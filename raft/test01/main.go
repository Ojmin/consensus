package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const raftCount = 3

// 节点的三种状态
const (
	stateFollower  = iota // 跟随者
	stateCandidate        // 候选者
	stateLeader           // 领导者
)

// 最初任期为0，-1代表没有编号
var leader = Leader{0, -1}

// 领导者
type Leader struct {
	Term int // 任期
	Id   int // 编号
}

// Raft节点
type Raft struct {
	mu              sync.Mutex // 锁
	id              int        // 节点编号
	curTerm         int        // 当前任期
	voteFor         int        // 为哪个节点投票以
	state           int        // 当前节点状态，0-follower，1-candidate，2-leader
	lastMsgTime     int64      // 发送最后一条消息时间
	curLeader       int        // 当前节点的领导
	timeout         int        // 超时时间
	msgCh           chan bool  // 消息通道
	electCh         chan bool  // 选举通道
	heartBeatCh     chan bool  // 心跳通道
	heartBeatRespCh chan bool  // 返回心跳信号
}

// 创建节点
func MakeNode(id int) *Raft {
	rf := &Raft{}
	rf.id = id
	rf.voteFor = -1          // 不投
	rf.state = stateFollower // follower状态
	rf.curLeader = -1        // 没有领导
	rf.setTerm(0)            // 设置任期
	rf.msgCh = make(chan bool)
	rf.electCh = make(chan bool)
	rf.heartBeatCh = make(chan bool)
	rf.heartBeatRespCh = make(chan bool)
	rand.Seed(time.Now().UnixNano())

	go rf.election()
	go rf.sendLeaderHeartBeat()

	return rf
}

// 选举
func (rf *Raft) election() {
	var result bool

	// 循环投票
	for {
		timeout := randRange(150, 500)
		rf.lastMsgTime = milliSeconds() // 每个节点最后一条消息时间
		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			fmt.Printf("当前节点状态为:%d\n", rf.state)
		}

		result = false

		// 选出leader，停止循环，result设置为true
		for !result {
			result = rf.electOneRand()
		}

	}
}

// 一次选举
func (rf *Raft) electOneRand(leader *Leader) bool {
	var (
		timeout          = int64(100) // 超时时间
		votes            = 0          // 投票数量
		triggerHeartBeat = false      // 是否开始心跳
		success          = false      // 返回值
	)
	last := milliSeconds() // 当前时间

	// 成为候选人状态
	rf.mu.Lock()
	rf.becomeCandidate()
	rf.mu.Unlock()

	fmt.Println("开始选举领导者...")
	// 选举
	for {
		// 遍历所有节点，进行投票
		for i := 0; i < raftCount; i++ {

		}
	}

	return success
}

// 修改节点为candidate状态
func (rf *Raft) becomeCandidate() {
	rf.state = stateCandidate  // 状态设置为候选人
	rf.setTerm(rf.curTerm + 1) // 任期加一
	rf.voteFor = rf.id         // 为哪个节点（自己）投票
	rf.curLeader = -1          // 没有领导
}

// 心跳检查
func (rf *Raft) sendLeaderHeartBeat() {

}

// 设置任期
func (rf *Raft) setTerm(term int) {
	rf.curTerm = term
}

// 产生随机数
func randRange(min, max int64) int64 {
	return rand.Int63n(max-min) + min
}

// 获取当前时间的毫秒数
func milliSeconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func main() {
	// 创建三个节点，最初是follower状态
	// 如果出现candidate状态节点，则开始投票，产生leader
	for i := 0; i < raftCount; i++ {
		MakeNode(i)
	}

}
