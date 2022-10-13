package coordinator

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	rpctask "github/rolandvarga/go-mapreduce/internal/rpc"
)

const TASK_CHECK_FREQUENCY = 1
const TASK_EXPIRATION_INTERVAL = 10
const MAP_PARTITION_COUNT = 3

type State int

const (
	Mapping State = iota
	Busy
	Reducing
	Finished
)

type Coordinator struct {
	nReduce     int
	mu          sync.Mutex
	MapTasks    map[string]TaskInfo
	ReduceTasks map[int]TaskInfo
	State       State
}

type TaskInfo struct {
	StartTime  time.Time
	InProgress bool
}

func (c *Coordinator) GetTask(arg *rpctask.MapTaskArg, reply *rpctask.TaskReply) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.updateState()

	switch c.State {
	case Mapping:
		// reply with Map jobs as long as there's MapTasks
		if filename, ok := c.GetNextMapItem(); ok {
			log.Infof("sending out file as map task: %s\n", filename)

			reply.Type = rpctask.MapTask
			reply.NReduce = c.nReduce
			reply.FileName = filename
			reply.MapId = c.getMapPartition()

			return nil
		}
		// all tasks are in progress
		reply.Type = rpctask.NoTask
		return nil

	case Reducing:
		if id, ok := c.GetNextReduceItem(); ok {
			log.Infof("sending out reduce id for processing: %d\n", id)

			reply.Type = rpctask.ReduceTask
			reply.ReduceId = id

			return nil
		}
		// all tasks are in progress
		reply.Type = rpctask.NoTask
		return nil

	case Busy, Finished:
		reply.Type = rpctask.NoTask
	default:
		reply.Type = rpctask.NoTask
	}
	return nil
}

func (c *Coordinator) CompleteMapTask(arg *rpctask.MapTaskArg, reply *rpctask.TaskReply) error {
	c.mu.Lock()
	delete(c.MapTasks, arg.FileName)
	c.mu.Unlock()
	return nil
}

func (c *Coordinator) CompleteReduceTask(arg *rpctask.ReduceTaskArg, reply *rpctask.TaskReply) error {
	c.mu.Lock()
	delete(c.ReduceTasks, arg.ID)
	c.mu.Unlock()
	return nil
}

func (c *Coordinator) updateState() {
	totalMapTasks := len(c.MapTasks)
	totalReduceTasks := len(c.ReduceTasks)

	availableReduceTasks := c.GetAvailableReduceItemCount()
	availableMapTasks := c.GetAvailableMapItemCount()

	if totalMapTasks > 0 {
		if availableMapTasks > 0 {
			// map tasks available for serving
			c.State = Mapping
			return
		}
		// all map tasks have been given out
		c.State = Busy
		return
	}

	// all map tasks done, ready to reduce
	if totalReduceTasks > 0 {
		if availableReduceTasks > 0 {
			c.State = Reducing
			return
		}
		// all reduce tasks have been given out
		c.State = Busy
		return
	}

	// all tasks done
	if totalMapTasks == 0 && totalReduceTasks == 0 {
		c.State = Finished
		return
	}
}

// partition the mapping tasks
func (c *Coordinator) getMapPartition() int {
	taskToComplete := 0
	for _, task := range c.MapTasks {
		if !task.InProgress {
			taskToComplete++
		}
	}
	return taskToComplete%MAP_PARTITION_COUNT + 1
}

func (c *Coordinator) GetAvailableMapItemCount() int {
	count := 0
	for _, taskInfo := range c.MapTasks {
		if taskInfo.InProgress == false {
			count++
		}
	}
	return count
}

func (c *Coordinator) GetNextMapItem() (string, bool) {
	for filename, taskInfo := range c.MapTasks {
		if taskInfo.InProgress == false {
			tmpTaskInfo := c.MapTasks[filename]
			tmpTaskInfo.InProgress = true
			tmpTaskInfo.StartTime = time.Now()

			c.MapTasks[filename] = tmpTaskInfo

			return filename, true
		}
	}
	return "", false
}

func (c *Coordinator) GetAvailableReduceItemCount() int {
	count := 0
	for _, taskInfo := range c.ReduceTasks {
		if taskInfo.InProgress == false {
			count++
		}
	}
	return count
}

func (c *Coordinator) GetNextReduceItem() (int, bool) {
	for id, taskInfo := range c.ReduceTasks {
		if taskInfo.InProgress == false {
			tmpTaskInfo := c.ReduceTasks[id]
			tmpTaskInfo.InProgress = true
			tmpTaskInfo.StartTime = time.Now()

			c.ReduceTasks[id] = tmpTaskInfo

			return id, true
		}
	}
	return -1, false
}

func (c *Coordinator) expireTasks(expireCh chan []string) {
	ticker := time.NewTicker(TASK_CHECK_FREQUENCY * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			expired := []string{}
			c.mu.Lock()
			for filename, taskInfo := range c.MapTasks {
				if taskInfo.InProgress && isExpired(taskInfo) {
					tmpTaskInfo := c.MapTasks[filename]
					tmpTaskInfo.InProgress = false
					c.MapTasks[filename] = tmpTaskInfo

					expired = append(expired, filename)
				}
			}
			for id, taskInfo := range c.ReduceTasks {
				if taskInfo.InProgress && isExpired(taskInfo) {
					tmpTaskInfo := c.ReduceTasks[id]
					tmpTaskInfo.InProgress = false
					c.ReduceTasks[id] = tmpTaskInfo

					expired = append(expired, fmt.Sprintf("%d", id))
				}
			}
			c.mu.Unlock()
			expireCh <- expired
		}
	}
}

func isExpired(task TaskInfo) bool {
	return time.Now().UTC().After(task.StartTime.Add(time.Duration(TASK_EXPIRATION_INTERVAL * time.Second)))
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server(errChan chan error) {
	rpc.Register(c)
	rpc.HandleHTTP()

	sockname := rpctask.CoordinatorSock()

	// if sockname exists, remove it
	if _, err := os.Stat(sockname); err == nil {
		err = os.Remove(sockname)
		if err != nil {
			errChan <- fmt.Errorf("unable to remove socket: %v", err)
		}
	}

	log.Infof("starting coordinator on socket '%v'\n", sockname)

	listener, err := net.Listen("unix", sockname)
	if err != nil {
		errChan <- fmt.Errorf("listen error: %v", err)
	}
	go func() {
		errChan <- http.Serve(listener, nil)
	}()
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.MapTasks) == 0 && len(c.ReduceTasks) == 0 {
		return true
	}

	return false
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	c.nReduce = nReduce

	reduceTasks := make(map[int]TaskInfo)
	for i := 0; i < c.nReduce; i++ {
		reduceTasks[i] = TaskInfo{StartTime: time.Now(), InProgress: false}
	}
	c.ReduceTasks = reduceTasks

	mapTasks := make(map[string]TaskInfo)
	for _, file := range files {
		mapTasks[file] = TaskInfo{StartTime: time.Now(), InProgress: false}
	}
	c.MapTasks = mapTasks

	errChan := make(chan error)
	c.server(errChan)

	// The coordinator should notice if a worker hasn't completed its task
	// in a reasonable amount of time, and give the same task to a different worker.
	expireCh := make(chan []string)
	go c.expireTasks(expireCh)

	for !c.Done() {
		select {
		case expired := <-expireCh:
			if len(expired) > 0 {
				log.Infof("expired the following tasks: %v\n", expired)
			}
		case err := <-errChan:
			if err != nil {
				log.Errorf("received an error in the coordinator chain: %v", err)
				os.Exit(1)
			}
		}
	}
	return &c
}
