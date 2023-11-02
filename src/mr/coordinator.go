package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"

// I am coordinator, so after I was init, I should wait for workers get info from me. ZYX
//

type Coordinator struct {
	// Your definitions here.

}

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
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
	ret := false

	// Your code here.

	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	// First we call this function out of this package. ZYX
	c := Coordinator{}
	// We have to split original file to
	// Your code here.

	c.server()
	return &c
}
