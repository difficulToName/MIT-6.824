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

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// Morris generously said I could use his code : ) ZYX
type ByKey []KeyValue

func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

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
	for {
		call_success := call("Coordinator.TaskDistribute", &ask_message, &reply)
		if !call_success || reply.WorkType == 2 {
			// Sth goes wrong.
			//fmt.Println("All job gets done and I will head home~")
			break
		} else if reply.WorkType == 1 {
			reducing(reply.FileSequence, reply.NReduce, reducef)
		} else {
			mapping(reply.FileDir, mapf, reply.NReduce, reply.FileSequence)
			//fmt.Println(reply.FileDir + "map finished.")
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

func mapping(filename string, mapf func(string, string) []KeyValue, nReduce int, fileSequence int) {
	var intermediate []KeyValue
	file, errOpen := os.Open(filename)
	defer file.Close()
	if errOpen != nil {
		fmt.Println("Opening file ", filename, " error")
		return
	}
	content, errRead := ioutil.ReadAll(file)
	if errRead != nil {
		fmt.Println("Reading file ", filename, " error")
		return
	}
	temp := mapf(filename, string(content))
	intermediate = append(intermediate, temp...)

	// Next we have to use bucket to store. ZYX
	// This part we can take guidance as reference. ZYX

	buckets := make([][]KeyValue, nReduce)
	for i := range buckets {
		buckets[i] = []KeyValue{}
	}

	for _, kv := range intermediate {
		// Here is ihash func comes in, ihash takes a string then output a integer. ZYX
		buckets[ihash(kv.Key)%nReduce] = append(buckets[ihash(kv.Key)%nReduce], kv)
	}

	// Now we have to store buckets on disk.
	// The guidance advise us to name intermediate file as
	// mr-X-Y, where X is Map task number and Y is the reduce task number. ZYX
	// For bucket we initialized,
	for i := range buckets {
		path := "mr-" + strconv.Itoa(fileSequence) + "-" + strconv.Itoa(i)
		jsonData, err := json.Marshal(buckets[i])
		if err != nil {
			fmt.Println("Convert slice to ", path, " failed.")
			return
		}
		ioutil.WriteFile(path, jsonData, os.ModePerm)
	}

}

func reducing(fileSequence int, nReduces int, reducef func(string, []string) string) {
	var intermediate []KeyValue
	for i := 0; i < nReduces; i++ {
		fileName := "mr-" + strconv.Itoa(fileSequence) + "-" + strconv.Itoa(i)
		//fmt.Println(fileName)
		jsonFile, err := ioutil.ReadFile(fileName)
		//fmt.Println("Origin file", len(jsonFile))
		if err != nil {
			fmt.Println("Open json file failed!")
			return
		}

		json.Unmarshal(jsonFile, &intermediate)

	}
	// Now we successfully get all json file to memory. ZYX
	// Code below is copy from "mrsequential.go" ZYX
	sort.Sort(ByKey(intermediate))
	fileName := "mr-out-" + strconv.Itoa(fileSequence)
	ofile, _ := os.Create(fileName)

	i := 0
	//fmt.Println(len(intermediate))
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
		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)
		i = j // i = j here to jump over same key lest re-computation. ZYX
	}
	ofile.Close()

}
