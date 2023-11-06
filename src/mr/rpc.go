package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
)
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
type Ask struct {
	TaskType int // 0 for map || 1 for reduce || 2 for sleep and 3 for exit

	FileName       string // Just for mapping
	FileSequence   int
	ReduceSequence int // For finding in reducing and naming in mapping

	Buckets    int // How much buckets you want?
	FileAmount int
}

type Notification struct {
	TaskType int // Same as above

	FileNumber   int
	ReduceNumber int // Same as above
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
