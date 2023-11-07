# MIT 6.824 NOTE
## By Zhang LuYao from DHU 2024.11.01

### 11.01 Read mrsequential.go and mark some hints on it.
#### Explain were added behind the line of code and marked with ZYX. 

Question: What is the internal of the dynamic file "wc.so", I have known the signature of mapf and reducef.
If they take "value" as time that word show up, why they don't use integer type as input?

:) FUCK ME! The wc.so have source code. Just under dir "mrapps".
The implementation is same as I thought, they use Itoa func, namely, turn a integer into string. 
But why they didn't just use this function directly but use a dynamic lib?

#### Implement a distributed Mapreduce system.
A coordinator and few workers.

Outline: 
1. Worker get tasks from coordinator.
2. Worker get data from one or more files, execute and write output into new files.
3. Coordinator should keep track if worker finish work in reasonable time, in this exp we set 10 seconds.
4. All communication should go through RPC.

To run this project, first you have to compile a dynamic lib file. (.dll for windows and .so for unix-like platform)

Before we continue, we have to ease ourselves into rpc in Golang.

Rpc: rpc in Go is an encapsulated package. Here are some important methods in it:
1. We pass a pointer to object to rpc by rpc.Register. Then client could call the method provided by object we passed to rpc.
2. So called rpc.HandleHTTP() means that you tell rpc package to handle those RPC request were sent to HTTP server's /rpc path. Maybe our service would be called by HTTP protocol?
However, we should notice that rpc implemented by Golang can be based on http but original RPC is not based on HTTP!
3. Use rpc.DialHTTP would init a rpc request through http protocol, the server must have opened their http to receive that.

Then some data structure. Before you do these labs you better grasp yourself some common sense of Golang.
Coordinator: To distribute tasks.

Rpc: Here we don't need u implement a rpc, that only a package for you to call to use. You could define structure in that file.

Worker: To mapping and reducing.

Here we start the task! From MapReduce can we know: coordinator should split file into pieces and tell workers come in which file should they handle.

### 11.03 Get rpc job done, now two process could communicate through rpc.
Now our task is figure out what map and what is reduce, now the only thing we know is the workers takes charge for map and reduce process. Now I have a good
 plan. First let's talk about worker. Workers are charge for both mapping and reducing. But there exist a sequence, namely, first map, then reduce. So let's use a slice
 to store all filename. Then we declare two integer and use them as pointer. First pointer illustrates "Next file to be mapped", second one illustrates "Next file to be reduced".
 That is easy to see:
1. No matter which pointer, their max value is len(filename), namely, the size of the slice of filename.
2. Pointer 2 should less than or equal to Pointer 1.
3. When Pointer 1 equals len(filename) all job get done.


### 11.06 Lab 1 finish!
The implementation before yesterday was legal in some degree. However that method can not handle some scenario like "worker died" or "worker
 failed to map or reduce" so we have to design a function and add a feature to let coordinator supervise the status of worker.

For coordinator, coordinator records a task's status, whether it was mapped or reduced, mapping or reducing. Whenever a worker picks up a task. 
Coordinator starts a goroutine then count 10 seconds to check if worker changes the status of current task. If the status of current task is still 
doing(like mapping or reducing), that means worker died or sth went wrong. So goroutine would manipulate that state to origin state(to be mapped or to be reduced).

Next let's talk about worker, how does a worker report "I have finish my current job" to coordinator? The answer is through go/rpc. After the job gets done, 
the worker call the corresponding function provided by coordinator, then coordinator would change that tasks state. So even goroutine checks the status of 
that task, goroutine wouldn't manipulate that task's status to (To be done).


