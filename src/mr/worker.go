package mr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"time"
)
import "log"
import "net/rpc"
import "hash/fnv"

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}
type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {
	// Your worker implementation here.
	for {
		ask := Ask{}
		not := Notification{}
		success := call("Coordinator.AskTask", &not, &ask)
		//if ask.TaskType == 1 {
		//	fmt.Println("In ", ask.ReduceSequence)
		//}
		if !success || ask.TaskType == 3 {
			break
		}
		if ask.TaskType == 0 {
			// This branch is for mapping.
			intermediate := []KeyValue{}
			file, err := os.Open(ask.FileName)
			if err != nil {
				log.Fatalln("Open file error!")
			}
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalln("Read file error!")
			}
			file.Close()

			// We use mapf, which provided by func
			kva := mapf(ask.FileName, string(content))
			intermediate = append(intermediate, kva...)

			// Then we make a bucket to store those intermediate files.
			buckets := make([][]KeyValue, ask.Buckets)
			for i := range buckets {
				buckets[i] = []KeyValue{}
			}
			for _, value := range intermediate {
				buckets[ihash(value.Key)%ask.Buckets] = append(buckets[ihash(value.Key)%ask.Buckets], value)
			}

			for i := range buckets {
				pathName := "mr-" + strconv.Itoa(ask.FileSequence) + "-" + strconv.Itoa(i)
				openFile, _ := ioutil.TempFile("", pathName+"*")
				encode := json.NewEncoder(openFile)
				for _, value := range buckets[i] {
					err := encode.Encode(&value)
					if err != nil {
						log.Fatalln("Encode failed")
					}
				}
				os.Rename(openFile.Name(), pathName)
				openFile.Close()
			}
			not.TaskType = 0
			not.FileNumber = ask.FileSequence
			call("Coordinator.Notify", &not, &ask)
		} else if ask.TaskType == 1 {
			// Reducing here!
			//fmt.Println(ask.ReduceSequence)
			intermediate := []KeyValue{}
			for i := 0; i < ask.FileAmount; i++ {
				fileName := "mr-" + strconv.Itoa(i) + "-" + strconv.Itoa(ask.ReduceSequence)
				file, err := os.Open(fileName)
				if err != nil {
					log.Fatalln("Open file failed!")
				}
				decode := json.NewDecoder(file)
				for {
					kv := KeyValue{}
					err := decode.Decode(&kv)
					if err != nil {
						break
					}
					intermediate = append(intermediate, kv)
				}
				file.Close()
			}
			sort.Sort(ByKey(intermediate))

			outputName := "mr-out-" + strconv.Itoa(ask.ReduceSequence)
			//fmt.Println(outputName)
			outputFile, _ := os.Create(outputName)

			i := 0
			for i < len(intermediate) {
				j := i + 1
				for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
					j++
				}
				// Here we got interval [i, j) ZYX.
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, intermediate[k].Value)
					// Take objects have same key into a slice. ZYX
				}
				output := reducef(intermediate[i].Key, values) // Pass that slice to reducef. ZYX

				// this is the correct format for each line of Reduce output.
				fmt.Fprintf(outputFile, "%v %v\n", intermediate[i].Key, output)

				i = j // i = j here to jump over same key lest re-computation. ZYX
			}
			outputFile.Close()

			not.TaskType = 1
			not.ReduceNumber = ask.ReduceSequence
			call("Coordinator.Notify", &not, &ask)

		} else {
			//fmt.Println("Sleep")
			time.Sleep(time.Second)
		}
	}

	// uncomment to send the Example RPC to the coordinator.
	//CallExample()

}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
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
