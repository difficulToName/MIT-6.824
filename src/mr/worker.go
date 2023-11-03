package mr

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)
import "log"
import "net/rpc"
import "hash/fnv"

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

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

	// Here we implement our worker. ZYX
	// Two parameter are sth like function pointer or std::function<> in C++. ZYX
	ask_message := AskTask{}
	ask_message.WorkerNumber = 10
	reply := Reply{}
	call_success := call("Coordinator.TaskDistribute", &ask_message, &reply)
	for {
		if !call_success || reply.WorkType == 2 {
			// Sth goes wrong.
			fmt.Printf("All job gets done and I will head home~")
		} else if reply.WorkType == 1 {
			fmt.Printf("I am reducing")
		} else {
			fmt.Printf("I am mapping")
		}
		time.Sleep(time.Second)
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
// This call function could be use directly.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

func mapping(filename string) []KeyValue {
	ret := []KeyValue
	file, errOpen := os.Open(filename)
	defer file.Close()
	if errOpen != nil {
		fmt.Println("Opening file ", filename, " error")
	}
	content, errRead := ioutil.ReadAll(file)
	if errRead != nil {
		fmt.Println("Reading file", filename, "error")
	}



}