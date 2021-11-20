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

// s0(follower) - leader - currentTerm2
// s1(leader) - leader - currentTerm2
// 解决办法：两步检查（double check）

func (rf *Raft) AttemptElection() {
	rf.mu.Lock()
	rf.state = Candidate
	rf.currentTerm++
	rf.votedFor = rf.me
	log.Printf("[%d] attempting an election at term %d", rf.me, rf.currentTerm)
	votes := 1
	done := false
	term := rf.currentTerm
	rf.mu.Unlock()
	for _, server := range rf.peers {
		if server == rf.me {
			continue
		}
		go func(server int) {
			voteGranted := rf.CallRequestVote(server, term)
			if !voteGranted {
				return
			}
			rf.mu.Lock()
			defer rf.mu.Unlock()
			votes++
			log.Printf("[%d] got vote from %d", rf.me, server)
			if done || votes <= len(rf.peers)/2 {
				return
			}
			done = true
			// if rf.state != Candidate || rf.currentTerm != term {	// 解决办法
			// 	return
			// }
			log.Printf("[%d] we got enough votes, we are now the leader (currentTem=%d, state=%v)!", rf.me, rf.currentTerm, rf.state)
			rf.state = Leader
		}(server)
	}
}

func (rf *Raft) CallRequestVote(server int, term int) bool {
	log.Printf("[%d] sending request vote to %d", rf.me, server)
	args := RequestVoteArgs{
		Term:        term,
		CandidateID: rf.me,
	}
	var reply RequestVoteReply
	ok := rf.sendRequestVote(server, &args, &reply)
	log.Printf("[%d] finish sending request vote to %d", rf.me, server)
	if !ok {
		return false
	}
	// ... process the reply
	return true
}

func (rf *Raft) HandleRequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	log.Printf("[%d] received request vote from %d", rf.me, args.CandidateID)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	log.Printf("[%d] handling request vote from %d", rf.me, args.CandidateID)
	// ...
}
