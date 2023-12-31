# MIT 6.824 NOTE
## By Zhang LuYao from DHU 2024.11.01

### 11.01 Read mrsequential.go and mark some hints on it.
#### Explain were added behind the line of code and marked with ZYX. 

Question: What is the internal of the dynamic file "wc.so", I have known the signature of mapf and reducef.
If they take "value" as time that word show up, why they don't use integer type as input?

:) FUCK ME! The wc.so have source code. Just under dir "mrapps".
The implementation is same as I thought, they use Itoa func, namely, turn an integer into string. 
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
Now our task is figure out what map and what is "reduce", now the only thing we know is the workers takes charge for map and reduce process. Now I have a good
 plan. First let's talk about worker. Workers are charge for both mapping and reducing. But there exist a sequence, namely, first map, then reduce. So let's use a slice
 to store all filename. Then we declare two integer and use them as pointer. First pointer illustrates "Next file to be mapped", second one illustrates "Next file to be reduced".
 That is easy to see:
1. No matter which pointer, their max value is len(filename), namely, the size of the slice of filename.
2. Pointer 2 should less than or equal to Pointer 1.
3. When Pointer 1 equals len(filename) all job get done.


### 11.06 Lab 1 finish!
The implementation before yesterday was legal in some degree. However, that method can not handle some scenario like "worker died" or "worker
 failed to map or reduce" so we have to design a function and add a feature to let coordinator supervise the status of worker.

For coordinator, coordinator records a task's status, whether it was mapped or reduced, mapping or reducing. Whenever a worker picks up a task. 
Coordinator starts a goroutine then count 10 seconds to check if worker changes the status of current task. If the status of current task is still 
doing(like mapping or reducing), that means worker died or sth went wrong. So goroutine would manipulate that state to origin state(to be mapped or to be reduced).

Next let's talk about worker, how does a worker report "I have finished my current job" to coordinator? The answer is through go/rpc. After the job gets done, 
the worker call the corresponding function provided by coordinator, then coordinator would change that tasks status. So even goroutine checks the status of 
that task, goroutine wouldn't manipulate that task's status to (To be done).

### 11.10 Read paper and hints. Lab 2A
For a server, it could only be at 3 status: follower, candidate or leader. When init a raft consensus, all server are be "follower" 
, they all have a timer to check last time it receives a rpc from leader, when that expires, it would increment its term and motivates 
other servers to vote for him. (The condition to convert from follower to leader is still unknown.) Suppose this server becomes leader. 
It should send signal periodically to inform its follower: "I am still alive and none of u dare to rebel!"
The method and detail can be found in hints and papers. 

### 11.11 Try to finish Lab 2A
In this lab, we have to implement two functions. One for transmit heartbeat signal from leader to followers. One for election, take charges for 
vote communicate between candidate and followers.
When consensus is at voting, followers will turn to candidate automatically when timer expires. Then it will ask other server to vote for him. As soon as 
it got votes amount greater than half amount of servers. It turns to leader and send AppendEntire RPC message to other servers. Server whose term not greater than 
leader's term should be his follower. Leader should send RPC periodically to his followers. Our thought is legal however we just can not pass the test.

### 11.12 Test Lab 2A passed
We added some fmt.Println in code and test code to find where the error is. The information given by tester is that servers have different terms. However, we, as a leader, had told 
follower servers to change their term they had already known. That's so wierd.

Through debugging we found the parameter latestTerm in heartbeat package, which after a server changed itself to leader sent had been wrongly set. After fix that bug we pass the test.

### 11.13 Lab 2B. Read paper and figure some fundamentals
Here, Lab 2B, we should implement this lab. Next I will write down some fundamentals to help u ease yourself into it.

In a raft consensus, all "write" command should be handled by master/leader node(server), and a read request could be handled by any server using raft. 
After leader get some command like "write", leader should send replication command to its followers, after the majority of followers had acknowledged that they had store or save 
that command/log, leader server could reply to the client: "Your command is implemented!".

In this Lab we will focus on how control communications between leader and its followers. Here are some details: 
1. If any of follower crashes or packet lose which causes follower didn't reply, leader should retry send packet or call rpc indefinitely. Even after it has responded to client until follower 
store all logs.
2. Any request packet should have at least two information: the term of current leader and the log information. We all know the term number plays a significant role to refrain its follower 
start next turn of vote. In another hand, it acts as a timestamp to let followers know whether this is the latest log it should accept.
3. Any packet should have a integer index to identify its position.
4. Like rule 1, leader node should commit a log entry after the majority of follower acknowledged that log. And that log should be marked with "committed".
5. Leader should maintain an integer which illustrates the latest index of log it had committed, and this parameter should be involved in the AppendEntries RPC.

