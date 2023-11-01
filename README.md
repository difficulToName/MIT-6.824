# MIT 6.824 NOTE
## By Zhang LuYao from DHU 2024.11.01

### 11.01 Read mrsequential.go and mark some instructions on it.
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
4. All communication should through RPC.

To run this project, first you have to compile a dynamic lib file. (.dll for windows and .so for unix-like platform)

Before we continue, we have to ease ourselves into rpc and http or etc. in Golang.

rpc: rpc in Go is an encapsulated package. Here are some important methods in it:
