package mr

import (
	"log"
	"sync"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

// I am coordinator, so after I was init, I should wait for workers get info from me. ZYX

type Coordinator struct {
	// Your definitions here.
	lock        sync.Mutex // This variable is for goroutine-safe.
	files       []string
	fileMapped  int // This pointer points to the next file should be mapped. ZYX
	fileReduced int // This pointer points to same slice as above but shouldn't be front / right of above. ZYX
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) TaskDistribute(ask *AskTask, reply *Reply) error {
	// That is easy, if there is file unmapped we hand it to worker. ZYX
	// When all files are mapped, then we start to reduce. ZYX
	defer c.lock.Unlock()
	c.lock.Lock()
	if c.fileMapped != len(c.files) {
		reply.WorkType = 0 // Map it! ZYX
		reply.FileDir = c.files[c.fileMapped]
		c.fileMapped++
	} else if c.fileReduced != len(c.files) {
		reply.WorkType = 1 // Reduce it! ZYX
		reply.FileDir = c.files[c.fileReduced]
		c.fileReduced++
	} else {
		reply.WorkType = 2 // Just break. ZYX
	}
	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock() // I am a string. The string is made by system. You can render it as fixed. ZYX
	os.Remove(sockname)           // Maybe for remove last rested file? ZYX
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	return c.fileReduced == len(c.files)
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	// First we call this function out of this package. ZYX
	c := Coordinator{}
	// At here we should notify Coordinator the filename we want to operate. ZYX
	// Here we are in a function scope, so we just copy name in our struct. ZYX
	// (Whisper) Doesn't golang got sth like std::move in C++? ZYX
	for _, filename := range files {
		c.files = append(c.files, filename)
	}
	c.server()
	return &c
}
