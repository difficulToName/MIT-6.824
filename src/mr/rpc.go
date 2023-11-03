package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.
// This file is for user-defined data struct exclusive for rcp communication. ZYX
// For example, worker send A to coordinator and get reply from coordinator which structure is B. ZYX

type AskTask struct {
	WorkerNumber int
}

type Reply struct {
	WorkType int // Here we define 0 as map, 1 as reduce and 2 as sleep or job done
	FileDir  string
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
