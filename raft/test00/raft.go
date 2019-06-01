package main

import "time"

// 节点的三种状态
const (
	FOLLOWER  = iota // 跟随者
	CANDIDATE        // 候选者
	LEADER           // 领导者

	HBINTERVAL = 50 * time.Microsecond // 心跳间隔50毫秒
)

// 请求投票
type VoteRequest struct {
	CandidateTerm int // 候选人任期
	CandidateId   int // 候选人ID
}

// 请求投票应答
type VoteReplay struct {
	Term        int
	VoteGranted bool // true - 候选人接受了投票
}

// Raft协议节点
type Raft struct {
	me          int // 节点ID
	currentTerm int // 当前任期
	votedFor    int // 候选人
	state       int // 状态

}

func (rf *Raft) GetState() (int, int) { return rf.currentTerm, rf.state }
func (rf *Raft) IsLeader() bool       { return rf.state == LEADER }
