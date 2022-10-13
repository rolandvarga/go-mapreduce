# go-mapreduce
This is a MapReduce implementation I did for Lab 1 of the `MIT 6.824: Distributed Systems` course.
Lab spec: https://pdos.csail.mit.edu/6.824/labs/lab-mr.html

### running MR
The coordinator is expecting text files with the `*.txt`. The app in `mrapps/wc.go` was supplied by the course and is used for this example.

```bash
# running the coordinator
> go run -race mrcoordinator.go ~/_Kompi/GO/src/g.csail.mit.edu/6.824/src/main/*.txt

INFO[2022-10-13T10:56:50-04:00] starting coordinator on socket '/var/tmp/824-mr-501'
INFO[2022-10-13T10:58:11-04:00] sending out file as map task: pg-being_ernest.txt
INFO[2022-10-13T10:58:12-04:00] sending out file as map task: pg-huckleberry_finn.txt
INFO[2022-10-13T10:58:15-04:00] sending out file as map task: pg-dorian_gray.txt
INFO[2022-10-13T10:58:17-04:00] sending out file as map task: pg-frankenstein.txt
INFO[2022-10-13T10:58:19-04:00] sending out file as map task: pg-metamorphosis.txt
INFO[2022-10-13T10:58:19-04:00] sending out file as map task: pg-grimm.txt
INFO[2022-10-13T10:58:22-04:00] sending out file as map task: pg-sherlock_holmes.txt
INFO[2022-10-13T10:58:25-04:00] sending out file as map task: pg-tom_sawyer.txt
INFO[2022-10-13T10:58:26-04:00] sending out reduce id for processing: 4
INFO[2022-10-13T10:58:27-04:00] sending out reduce id for processing: 5
INFO[2022-10-13T10:58:27-04:00] sending out reduce id for processing: 7
INFO[2022-10-13T10:58:27-04:00] sending out reduce id for processing: 2
INFO[2022-10-13T10:58:28-04:00] sending out reduce id for processing: 0
INFO[2022-10-13T10:58:28-04:00] sending out reduce id for processing: 3
INFO[2022-10-13T10:58:29-04:00] sending out reduce id for processing: 9
INFO[2022-10-13T10:58:29-04:00] sending out reduce id for processing: 1
INFO[2022-10-13T10:58:29-04:00] sending out reduce id for processing: 8
INFO[2022-10-13T10:58:30-04:00] sending out reduce id for processing: 6
```

```bash
# starting the workers
> go build -race -buildmode=plugin ../mrapps/wc.go && go run -race mrworker.go wc.so

INFO[2022-10-13T10:58:11-04:00] received a new map task: rpc.TaskReply{FileName:"pg-being_ernest.txt", NReduce:10, Type:0, MapId:2, ReduceId:0}
INFO[2022-10-13T10:58:12-04:00] received a new map task: rpc.TaskReply{FileName:"pg-huckleberry_finn.txt", NReduce:10, Type:0, MapId:1, ReduceId:0}
INFO[2022-10-13T10:58:15-04:00] received a new map task: rpc.TaskReply{FileName:"pg-dorian_gray.txt", NReduce:10, Type:0, MapId:3, ReduceId:0}
INFO[2022-10-13T10:58:17-04:00] received a new map task: rpc.TaskReply{FileName:"pg-frankenstein.txt", NReduce:10, Type:0, MapId:2, ReduceId:0}
INFO[2022-10-13T10:58:19-04:00] received a new map task: rpc.TaskReply{FileName:"pg-metamorphosis.txt", NReduce:10, Type:0, MapId:1, ReduceId:0}
INFO[2022-10-13T10:58:19-04:00] received a new map task: rpc.TaskReply{FileName:"pg-grimm.txt", NReduce:10, Type:0, MapId:3, ReduceId:0}
INFO[2022-10-13T10:58:22-04:00] received a new map task: rpc.TaskReply{FileName:"pg-sherlock_holmes.txt", NReduce:10, Type:0, MapId:2, ReduceId:0}
INFO[2022-10-13T10:58:25-04:00] received a new map task: rpc.TaskReply{FileName:"pg-tom_sawyer.txt", NReduce:10, Type:0, MapId:1, ReduceId:0}
INFO[2022-10-13T10:58:26-04:00] received a new reduce task: rpc.TaskReply{FileName:"", NReduce:0, Type:1, MapId:0, ReduceId:4}
INFO[2022-10-13T10:58:26-04:00] parsing the following reduce files: [mr-1-4 mr-2-4 mr-3-4]
INFO[2022-10-13T10:58:27-04:00] received a new reduce task: rpc.TaskReply{FileName:"", NReduce:0, Type:1, MapId:0, ReduceId:5}
INFO[2022-10-13T10:58:27-04:00] parsing the following reduce files: [mr-1-5 mr-2-5 mr-3-5]
INFO[2022-10-13T10:58:27-04:00] received a new reduce task: rpc.TaskReply{FileName:"", NReduce:0, Type:1, MapId:0, ReduceId:7}
INFO[2022-10-13T10:58:27-04:00] parsing the following reduce files: [mr-1-7 mr-2-7 mr-3-7]
INFO[2022-10-13T10:58:27-04:00] received a new reduce task: rpc.TaskReply{FileName:"", NReduce:0, Type:1, MapId:0, ReduceId:2}
INFO[2022-10-13T10:58:27-04:00] parsing the following reduce files: [mr-1-2 mr-2-2 mr-3-2]
INFO[2022-10-13T10:58:28-04:00] received a new reduce task: rpc.TaskReply{FileName:"", NReduce:0, Type:1, MapId:0, ReduceId:0}
INFO[2022-10-13T10:58:28-04:00] parsing the following reduce files: [mr-1-0 mr-2-0 mr-3-0]
INFO[2022-10-13T10:58:28-04:00] received a new reduce task: rpc.TaskReply{FileName:"", NReduce:0, Type:1, MapId:0, ReduceId:3}
INFO[2022-10-13T10:58:28-04:00] parsing the following reduce files: [mr-1-3 mr-2-3 mr-3-3]
INFO[2022-10-13T10:58:29-04:00] received a new reduce task: rpc.TaskReply{FileName:"", NReduce:0, Type:1, MapId:0, ReduceId:9}
INFO[2022-10-13T10:58:29-04:00] parsing the following reduce files: [mr-1-9 mr-2-9 mr-3-9]
INFO[2022-10-13T10:58:29-04:00] received a new reduce task: rpc.TaskReply{FileName:"", NReduce:0, Type:1, MapId:0, ReduceId:1}
INFO[2022-10-13T10:58:29-04:00] parsing the following reduce files: [mr-1-1 mr-2-1 mr-3-1]
INFO[2022-10-13T10:58:29-04:00] received a new reduce task: rpc.TaskReply{FileName:"", NReduce:0, Type:1, MapId:0, ReduceId:8}
INFO[2022-10-13T10:58:29-04:00] parsing the following reduce files: [mr-1-8 mr-2-8 mr-3-8]
INFO[2022-10-13T10:58:30-04:00] received a new reduce task: rpc.TaskReply{FileName:"", NReduce:0, Type:1, MapId:0, ReduceId:6}
INFO[2022-10-13T10:58:30-04:00] parsing the following reduce files: [mr-1-6 mr-2-6 mr-3-6]
INFO[2022-10-13T10:58:32-04:00] coordinator didn't respond and has most likely finished; exiting
```