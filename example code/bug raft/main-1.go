package main

import (
	"log"
	"sync"
)

type State string

const (
	Follower  State = "follower"
	Candidate       = "candidate"
	Leader          = "leader"
)

type Raft struct {
	mu    sync.Mutex
	me    int
	peers []int
	state State

	currentTerm int
	votedFor    int
}

func (rf *Raft) AttemptElection() {
	rf.mu.Lock()
	rf.state = Candidate
	rf.currentTerm++
	rf.votedFor = rf.me
	log.Printf("[%d] attempting an election at term %d", rf.me, rf.currentTerm)
	rf.mu.Unlock()
	for _, server := range rf.peers {
		if server == rf.me {
			continue
		}
		go func(server int) {
			voteGranted := rf.CallRequestVote(server)
			if !voteGranted {
				return
			}
			// ... tally the votes
		}(server)
	}
}

func (rf *Raft) CallRequestVote(server int) bool {
	rf.mu.Lock()
	defer rf.mu.UnLock()
	log.Printf("[%d] sending request vote to %d", rf.me, server)
	args := RequestVoteArgs{
		Term:        rf.currentTerm,
		CandidateID: rf.me,
	}
	var reply RequestVoteReply
	ok := rf.sendRequestVote(server, &args, &reply) // RPC 调用时加锁，导致死锁（互相等待）
	log.Printf("[%d] finish sending request vote to %d", rf.me, server)
	if !ok {
		return false
	}
	// ... process the reply
	return true
}

func (rf *Raft) HandleRequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	log.Printf("[%d] received request vote from %d", rf.me, args.CandidateID)
	rf.mu.Lock() // 获取锁失败
	defer rf.mu.Unlock()
	log.Printf("[%d] handling request vote from %d", rf.me, args.CandidateID)
	// ...
}

// s0.CallRequestVote, acquire the lock
// s0.CallRequestVote, send RPC to s1
// s1.CallRequestVote, acquire the lock
// s1.CallRequestVote, send RPC to s0
// s0.Handler, s1.Handler trying to acquire lock

// 解决方案：RPC 调用时不要加锁
// func (rf *Raft) AttemptElection() {
// 	rf.mu.Lock()
// 	rf.state = Candidate
// 	rf.currentTerm++
// 	rf.votedFor = rf.me
// 	log.Printf("[%d] attempting an election at term %d", rf.me, rf.currentTerm)
// 	term := rf.currentTerm
// 	rf.mu.Unlock()
// 	for _, server := range rf.peers {
// 		if server == rf.me {
// 			continue
// 		}
// 		go func (server int)  {
// 			voteGranted := rf.CallRequestVote(server, term)
// 			if !voteGranted {
// 				return
// 			}
// 			// ... tally the votes
// 		}(server)
// 	}
// }
// func (rf *Raft) CallRequestVote(server int, term int) bool {
// 	log.Printf("[%d] sending request vote to %d", rf.me, server)
// 	args := RequestVoteArgs{
// 		Term: term,
// 		CandidateID: rf.me,
// 	}
// 	var reply RequestVoteReply
// 	ok := rf.sendRequestVote(server, &args, &reply)	// RPC 调用时加锁，导致死锁（互相等待）
// 	log.Printf("[%d] finish sending request vote to %d", rf.me, server)
// 	if !ok {
// 		return false
// 	}
// 	// ... process the reply
// 	return true
// }


