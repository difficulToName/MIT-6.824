package mr

import (
	"fmt"
	"log"
	"sync"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

// I am coordinator, so after I was init, I should wait for workers get info from me. ZYX

type Coordinator struct {
	// Your definitions here.
	fileMapped   int
	fileReduced  int
	files        []string
	mapStates    []int      // 0 for not work || 1 for processing || 2 for mapped
	reduceStates []int      // 0 for not work || 1 for processing || 2 for reduced
	nReduce      int        // intermediates
	lock         sync.Mutex // This variable is for goroutine-safe.
}

func (c *Coordinator) Report(reply *Reply, report *Report) {
	// This function is to prevent a worker crashes and capture a task long time.
	c.lock.Lock()
	if report.WorkType == 0 {
		c.mapStates[report.FileSequence] = 2
		c.fileMapped++
	} else {
		c.reduceStates[report.FileSequence] = 2
		c.fileReduced++
	}
	c.lock.Unlock()
	// This function should work with a goroutine, which turn those expired task to their origin state.
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) GetTask(reply *Reply, report *Report) error {
	c.lock.Lock()
	if c.fileMapped != len(c.mapStates) {
		taskNumber := -1
		for key, value := range c.mapStates {
			if value == 0 { // This task still doesn't been handled.
				taskNumber = key
				break
			}
		}
		// If there is file still not been mapped but can not allocate work.
		// Let worker sleep.
		if taskNumber == -1 {
			reply.WorkType = 2
			c.lock.Unlock()
		} else {
			reply.WorkType = 0
			reply.NReduce = c.nReduce
			reply.FileSequence = taskNumber
			reply.FileName = c.files[taskNumber]
			fmt.Println(c.files[taskNumber])
			c.mapStates[taskNumber] = 1
			c.lock.Unlock()
			go func() {
				time.Sleep(time.Duration(5) * time.Second)
				c.lock.Lock()
				if c.mapStates[taskNumber] == 1 {
					c.mapStates[taskNumber] = 0
				}
				c.lock.Unlock()
			}()
		}
	} else if c.fileReduced != len(c.reduceStates) {
		taskNumber := -1
		for key, value := range c.reduceStates {
			if value == 0 { // No worker handled or handling this file.
				taskNumber = key
				break
			}
			// Still, sleep.
			if taskNumber == -1 {
				reply.WorkType = 2
				c.lock.Unlock()
			} else {
				reply.WorkType = 1
				reply.FileSequence = taskNumber
				reply.NReduce = c.nReduce
				c.reduceStates[taskNumber] = 1
				c.lock.Unlock()
				go func() {
					time.Sleep(time.Duration(5) * time.Second)
					c.lock.Lock()
					if c.reduceStates[taskNumber] == 1 {
						c.reduceStates[taskNumber] = 0
					}
					c.lock.Unlock()
				}()
			}
		}
	} else {
		// All jobs get done, let's head for home!
		c.lock.Unlock()
		reply.WorkType = 3
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
	c.fileMapped = 0
	c.fileReduced = 0
	c.reduceStates = make([]int, len(files))
	c.mapStates = make([]int, len(files))
	c.nReduce = nReduce
	for _, filename := range files {
		c.files = append(c.files, filename)
	}
	//fmt.Println(c.files)
	c.server()
	return &c
}
