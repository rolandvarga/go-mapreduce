package mr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net/rpc"
	"os"
	"path/filepath"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	rpctask "github/rolandvarga/go-mapreduce/internal/rpc"
)

const RETRY_INTERVAL = 200

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	running := true
	for running {
		task := GetTask()

		switch task.Type {
		case rpctask.MapTask:
			log.Infof("received a new map task: %#v\n", task)

			content, err := getFileContents(task.FileName)
			if err != nil {
				log.Errorf("error during map task: %v\n", err)
				break
			}

			pairs := mapf(task.FileName, string(content))

			for _, pair := range pairs {
				reduceId := ihash(pair.Key) % task.NReduce

				stagingFile := fmt.Sprintf("mr-%d-%d", task.MapId, reduceId)

				file, err := os.OpenFile(stagingFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Errorf("worker can't open mapping file: %v", err)
					break
				}

				enc := json.NewEncoder(file)
				err = enc.Encode(&pair)
				if err != nil {
					log.Errorf("worker can't write to intermediate file: %v", err)
					break
				}
				file.Close()
			}

			// tell coordinator when done with a task
			PostCompleteMapTask(task.FileName)

		case rpctask.ReduceTask:
			log.Printf("received a new reduce task: %#v\n", task)

			intermediatePairs := []KeyValue{}

			reduceFiles, err := filepath.Glob(fmt.Sprintf("mr-*-%d", task.ReduceId))
			log.Infof("parsing the following reduce files: %v\n", reduceFiles)
			if err != nil {
				log.Errorf("error finding reduce files: %v\n", err)
				break
			}
			for _, reduceFile := range reduceFiles {
				file, err := os.Open(reduceFile)
				if err != nil {
					log.Errorf("error reading reduce file: %v\n", err)
					break
				}

				// read file line by line
				fileScanner := bufio.NewScanner(file)
				for fileScanner.Scan() {
					var kv KeyValue
					err := json.Unmarshal(fileScanner.Bytes(), &kv)
					if err != nil {
						log.Errorf("error parsing reduce file: %v\n", err)
						break
					}
					intermediatePairs = append(intermediatePairs, kv)
				}
				file.Close()
			}

			// TODO we could create an RPC call on errors, so that reduce & map jobs can be retried faster
			outputFileName := fmt.Sprintf("mr-out-%d", task.ReduceId)
			tmpFile, err := ioutil.TempFile("", outputFileName)
			if err != nil {
				log.Errorf("error creating temp file: %v\n", err)
				break
			}

			// sort by key
			sort.Sort(ByKey(intermediatePairs))

			i := 0
			for i < len(intermediatePairs) {
				j := i + 1
				for j < len(intermediatePairs) && intermediatePairs[j].Key == intermediatePairs[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, intermediatePairs[k].Value)
				}
				output := reducef(intermediatePairs[i].Key, values)

				// this is the correct format for each line of Reduce output.
				fmt.Fprintf(tmpFile, "%v %v\n", intermediatePairs[i].Key, output)

				i = j
			}
			tmpFile.Close()
			os.Rename(tmpFile.Name(), outputFileName)

			PostCompleteReduceTask(task.ReduceId)
		case rpctask.NoTask:
			log.Debugf("no task given by coordinator; retrying in %d ms\n", RETRY_INTERVAL)
			time.Sleep(RETRY_INTERVAL * time.Millisecond)
		case rpctask.NoReply:
			log.Info("coordinator didn't respond and has most likely finished; exiting")
			os.Exit(0)
		default:
			log.Debugf("unknown task received from coordinator; retrying in %d ms\n", RETRY_INTERVAL)
			time.Sleep(RETRY_INTERVAL * time.Millisecond)
		}
	}
}

func GetTask() rpctask.TaskReply {
	arg := rpctask.MapTaskArg{}
	task := rpctask.TaskReply{}
	err := call("Coordinator.GetTask", &arg, &task)
	if err != nil {
		task.Type = rpctask.NoReply
	}
	return task
}

func PostCompleteMapTask(filename string) {
	arg := rpctask.MapTaskArg{FileName: filename}
	reply := rpctask.TaskReply{}

	if err := call("Coordinator.CompleteMapTask", &arg, &reply); err != nil {
		log.Errorf("error completing map task: %v\n", err)
	}
}

func PostCompleteReduceTask(id int) {
	arg := rpctask.ReduceTaskArg{ID: id}
	reply := rpctask.TaskReply{}

	if err := call("Coordinator.CompleteReduceTask", &arg, &reply); err != nil {
		log.Errorf("error completing reduce task: %v\n", err)
	}
}

// send an RPC request to the coordinator, wait for the response.
func call(rpcname string, args interface{}, reply interface{}) error {
	sockname := rpctask.CoordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		return fmt.Errorf("error dialing: +%v", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err != nil {
		return err
	}
	return nil
}

func getFileContents(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	file.Close()
	return content, nil
}
