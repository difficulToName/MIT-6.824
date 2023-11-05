package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

type Report struct {
	FileSequence int
	WorkType     int // 0 for mapping task || 1 for reducing task
}

type Reply struct {
	WorkType int
	// 0 for map || 1 for reduce || 2 for sleep || 3 for exit
	FileName     string
	NReduce      int
	FileSequence int
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
