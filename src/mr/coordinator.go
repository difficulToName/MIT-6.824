package mr

import (
	"log"
	"sync"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

type Coordinator struct {
	// Your definitions here.
	ReduceAmount   int
	MapAmount      int
	FileNames      []string
	MapFinished    int
	ReduceFinished int
	MapStatus      []int
	ReduceStatus   []int
	mutex          sync.Mutex
}

func (c *Coordinator) Notify(not *Notification, ask *Ask) error {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	if not.TaskType == 0 {
		//fmt.Println("Map finish" + strconv.Itoa(not.ReduceNumber))
		c.MapStatus[not.FileNumber] = 2
		c.MapFinished++
		//fmt.Println("Amount ", strconv.Itoa(c.MapAmount), " Finished ", c.MapFinished)
	} else {
		//fmt.Println("Reduce finish" + strconv.Itoa(not.ReduceNumber))
		c.ReduceStatus[not.ReduceNumber] = 2
		c.ReduceFinished++
	}
	return nil
}

// Your code here -- RPC handlers for the worker to call.
func (c *Coordinator) AskTask(not *Notification, ask *Ask) error {
	// Fuck Fuck Fuck!
	// Coordinator should reply tasks through parameter 2, namely, ask
	// In this function coordinator return task through Ask.
	// Because worker is stateless, so ask should be 0.
	defer c.mutex.Unlock()
	c.mutex.Lock()
	if c.MapFinished < c.MapAmount {
		taskIndex := -1
		for index, value := range c.MapStatus {
			if value == 0 {
				taskIndex = index
				break
			}
		}
		if taskIndex == -1 {
			ask.TaskType = 2
		} else {
			ask.TaskType = 0
			ask.FileName = c.FileNames[taskIndex]
			ask.FileSequence = taskIndex
			ask.Buckets = c.ReduceAmount
			//fmt.Println("Coor map task " + strconv.Itoa(taskIndex) + " allo")
			c.MapStatus[taskIndex] = 1
			go func() {
				defer c.mutex.Unlock()
				time.Sleep(time.Second * time.Duration(10))
				c.mutex.Lock()
				if c.MapStatus[taskIndex] == 1 {
					c.MapStatus[taskIndex] = 0
				}
			}()
		}
	} else if c.MapFinished == c.MapAmount && c.ReduceFinished < c.ReduceAmount {
		taskIndex := -1
		for index, value := range c.ReduceStatus {
			if value == 0 {
				taskIndex = index
				break
			}
		}
		if taskIndex == -1 {
			ask.TaskType = 2
		} else {
			ask.TaskType = 1
			ask.ReduceSequence = taskIndex
			ask.Buckets = c.ReduceAmount
			ask.FileAmount = len(c.FileNames)

			c.ReduceStatus[taskIndex] = 1
			//fmt.Println("Coor reduce task " + strconv.Itoa(ask.ReduceSequence) + " allo!")
			go func() {
				defer c.mutex.Unlock()
				time.Sleep(time.Second * time.Duration(10))
				c.mutex.Lock()
				if c.ReduceStatus[taskIndex] == 1 {
					c.ReduceStatus[taskIndex] = 0
				}
			}()
		}
	} else {
		// All jobs done.
		ask.TaskType = 3
	}
	//if ask.TaskType == 1 {
	//	fmt.Println("Leave out ", ask.ReduceSequence)
	//}
	return nil
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	//fmt.Println("Coor finish!")
	return c.ReduceFinished == c.ReduceAmount
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}
	// Your code here.
	c.FileNames = files
	c.MapAmount = len(files)
	c.MapStatus = make([]int, c.MapAmount)
	c.ReduceAmount = nReduce
	c.ReduceStatus = make([]int, nReduce)
	c.server()
	return &c
}
