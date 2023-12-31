package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, Term, isleader)
//   start agreement on a new log entry
// rf.GetState() (Term, isLeader)
//   ask a Raft for its current Term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	//	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.5840/labgob"
	"6.5840/labrpc"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type AppendEntriesArgs struct {
	// Self-defined, only for leader.
	// Some info
	// Leader shouldn't send heartbeat more than 10 times in a minute
	Term     int // The Term of leader
	LeaderId int
	// There maybe some another args in this struct
	// 11.13 Here we need some data structure to store log info in this struct and send through Go/rpc
	// Whether log is empty, follower should always upgrade its timer and term.
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []interface{} // We still don't know which data type should entries should store.
	// Above is the log info we have to store.
	LeaderCommit int // The latest index of leader committed
}

type AppendEntriesReply struct {
	Ok         bool // If a server acknowledge he is leader. Depends on Term number.
	LatestTerm int  // available when oj is false
	// If Ok is false, the leader should turn itself to follower instantly.
}

// A Go object implementing a single Raft peer.
type Raft struct {
	// Here we may have many Raft struct.
	// Each Raft struct is a raft server.
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()
	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	// Persistent status
	myTerm                int           // Increment only.
	myState               int           // 0 for follower || 1 for candidate || 2 for leader
	rpcFromLeaderLastTime time.Time     // only available as follower
	votedFor              int           // To save which server he had Voted for.
	myChannel             chan ApplyMsg // This channel is for communicate between peer server.
	// Volatile status
	lastLogCommitted int // To memory the index of the latest log to be committed (Commit could only happen on leader server)
	lastLogApplied   int // To memory the index of the latest log to be applied on state machine (On this machine)
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.myTerm, rf.myState == 2
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	CandidateTerm int
	CandidateID   int
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (2A).
	Voted      bool
	LatestTerm int // only available when follower didn't grant.
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).
	// This function is used for follower to tell candidate whether he votes.
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if args.CandidateTerm < rf.myTerm {
		reply.Voted = false
		reply.LatestTerm = rf.myTerm
		return
	}
	// If procedure comes to here, the Term of candidate's Term must greater than follower.
	// As paper said, now this raft server should be a follower.
	if args.CandidateTerm > rf.myTerm {
		// Now this server has to convert to follower state.
		rf.myState = 0
		rf.myTerm = args.CandidateTerm
		rf.votedFor = -1
		//fmt.Println("Change to follower,", rf.me, "changed term to", rf.myTerm)
	}
	// Whether we vote depends on if we have Voted yet.
	reply.LatestTerm = rf.myTerm
	if rf.votedFor == -1 || rf.votedFor == args.CandidateID {
		reply.Voted = true
		rf.votedFor = args.CandidateID
		rf.rpcFromLeaderLastTime = time.Now()
		return
	}
	reply.Voted = false
	return
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return. Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntire(server int, currentTerm int) bool {
	// Parameters above may be changed in Lab 2B.
	args := &AppendEntriesArgs{}
	reply := &AppendEntriesReply{}
	args.Term = currentTerm
	args.LeaderId = rf.me
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

// Self-defined
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	// This function is for reply message to client call it.
	// This function is used for communicate between leader and followers. In Lab 2A this function is used for
	// implement heartbeat function to refrain follower start voting and becoming candidate.
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.myState == 0 {
		rf.myTerm = args.Term
		//fmt.Println("because of heartbeat", rf.me, "change term into", rf.myTerm)
		rf.rpcFromLeaderLastTime = time.Now()
		reply.Ok = true
	} else if rf.myState == 1 {
		if rf.myTerm <= args.Term {
			// change to follower
			rf.myState = 0
			rf.myTerm = args.Term
			reply.Ok = true
		} else {
			reply.Ok = false
			reply.LatestTerm = rf.myTerm
			// Through paper, candidate should continue his state but turn into leader.
		}
	}
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// Term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	//index := -1
	//term := -1
	//isLeader := true

	// Your code here (2B).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	return 0, 0, rf.myState == 2

}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) leaderHeartBeat() {
	for {
		if rf.killed() || rf.myState != 2 {
			break
		}
		rf.mu.Lock()
		currentTerm := rf.myTerm
		// We can change this variable to atomic variable.
		rf.mu.Unlock()
		for i := range rf.peers {
			if i != rf.me {
				go func(server int) {
					rf.sendAppendEntire(server, currentTerm)
				}(i)
			}
		}
		time.Sleep(time.Millisecond * time.Duration(100))
	}
}

func (rf *Raft) newElection() {
	votes := 1
	stillVoting := true
	var voteMutex sync.Mutex
	rf.mu.Lock()
	// Now I am the candidate and send vote request rpc to other servers.
	rf.myTerm++
	//fmt.Println(rf.me, "term added!")
	rf.myState = 1
	rf.votedFor = rf.me
	args := &RequestVoteArgs{}
	args.CandidateTerm = rf.myTerm
	args.CandidateID = rf.me
	rf.mu.Unlock()
	for i := range rf.peers {
		if i != rf.me {
			go func(server int) {
				reply := &RequestVoteReply{}
				success := rf.sendRequestVote(server, args, reply)
				voteMutex.Lock()
				if success && reply.Voted && stillVoting {
					votes++
					if votes*2 > len(rf.peers) {
						stillVoting = false
						rf.mu.Lock()
						rf.myState = 2
						//fmt.Println(rf.me, "is leader now!")
						rf.mu.Unlock()
						go rf.leaderHeartBeat()
					}
				} else if success && !reply.Voted {
					rf.mu.Lock()
					rf.myState = 0
					rf.myTerm = reply.LatestTerm
					stillVoting = false
					rf.mu.Unlock()
				}
				voteMutex.Unlock()
			}(i)
		}
	}
}

func (rf *Raft) ticker() {
	var ms = 50 + (rand.Int63() % 300)
	for rf.killed() == false { // As long as this server haven't been closed, we will check whether timer expires.

		// Your code here (2A)
		// Check if a leader election should be started.
		// Let's check if we have received periodically sent RPC from leader
		// 1 second == 1000 milliseconds
		rf.mu.Lock()
		currentTime := time.Now()
		duration := currentTime.Sub(rf.rpcFromLeaderLastTime)
		if duration >= time.Duration(ms)*time.Millisecond && rf.myState != 2 {
			// Add my Term and send request to other raft server.
			go rf.newElection()
		}
		rf.mu.Unlock()
		// pause for a random amount of time between 50 and 350
		// milliseconds.
		ms = 50 + (rand.Int63() % 300)
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	// Go would initialize many variable to 0. So we don't have to set them manually.
	rf.rpcFromLeaderLastTime = time.Now()
	rf.myTerm = 0
	rf.votedFor = -1
	rf.myState = 0 // Now I am follower

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	// start ticker goroutine to start elections
	go rf.ticker()
	return rf
}
