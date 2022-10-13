package rpc

import (
	"os"
	"strconv"
)

type TaskType int

const (
	MapTask TaskType = iota
	ReduceTask
	NoTask
	NoReply
)

type TaskReply struct {
	FileName string
	NReduce  int
	Type     TaskType
	MapId    int
	ReduceId int
}

type MapTaskArg struct {
	FileName string
}

type ReduceTaskArg struct {
	ID int
}

func CoordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
