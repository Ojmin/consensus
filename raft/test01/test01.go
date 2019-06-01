package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const raftCount = 3 // 节点个数

// 节点的三种状态
const (
	stateFollower  = iota // 跟随者
	stateCandidate        // 候选者
	stateLeader           // 领导者
)

var (
	leader     = Leader{Term: 0, LeaderId: -1} // 最初任期为0，-1代表没有leader
	shutdownCh = make(chan struct{})           // 主线程退出通道
)

// 领导者
type Leader struct {
	Term     int // 任期
	LeaderId int // 编号
}

// Raft节点
type Raft struct {
	mu              sync.Mutex // 锁
	me              int        // 节点编号
	curTerm         int        // 当前任期
	voteFor         int        // 为哪个节点投票以
	state           int        // 当前节点状态
	lastMsgTime     int64      // 发送最后一条消息时间
	curLeader       int        // 当前节点的领导
	timeout         int        // 超时时间
	msgCh           chan bool  // 消息通道
	electCh         chan bool  // 选举通道
	heartbeatCh     chan bool  // 心跳通道
	heartbeatRespCh chan bool  // 返回心跳信号
}

// 创建节点
func MakeNode(id int) *Raft {
	rf := &Raft{}
	rf.me = id
	rf.voteFor = -1          // 不投
	rf.state = stateFollower // follower状态
	rf.timeout = 0
	rf.curLeader = -1 // 没有领导
	rf.setTerm(0)     // 设置任期
	rf.msgCh = make(chan bool)
	rf.electCh = make(chan bool)
	rf.heartbeatCh = make(chan bool)
	rf.heartbeatRespCh = make(chan bool)
	rand.Seed(time.Now().UnixNano())

	return rf
}

// 选举
func (rf *Raft) election() {
	var result = false // 标识是否选举成功（选出leader）
	for {              // 循环投票
		timeout := randRange(150, 500)
		rf.lastMsgTime = milliSeconds() // 每个节点最后一条消息时间
		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			fmt.Printf("节点[%d]状态为: %d\n", rf.me, rf.state)
		}

		for !result {
			result = rf.electOneRand(&leader) // 选举leader，若选出leader，返回true
		}
	}
}

// 一次选举，选leader
func (rf *Raft) electOneRand(leader *Leader) bool {
	var (
		timeout          = int64(100)     // 超时时间
		votes            = 0              // 投票数量
		triggerHeartBeat = false          // 是否开始心跳
		success          = false          // 返回值
		last             = milliSeconds() // 当前时间
	)

	rf.mu.Lock()
	rf.becomeCandidate() // 成为候选人状态
	rf.mu.Unlock()

	fmt.Printf("节点[%d]开始选举Leader...\n", rf.me)
	for { // 选举
		for i := 0; i < raftCount; i++ { // 遍历所有节点，进行投票
			if i != rf.me { // 其他节点
				go func() {
					if leader.LeaderId < 0 { // 其他节点没有领导
						rf.electCh <- true
					}
				}()
			}
		}

		for i := 0; i < raftCount; i++ {
			select {
			case ok := <-rf.electCh: // 统计票数
				if ok {
					votes++
					success = votes > raftCount/2
					if success && !triggerHeartBeat { // 票数过半且未触发心跳
						triggerHeartBeat = true // 触发心跳
						rf.mu.Lock()
						rf.becomeLeader()
						rf.mu.Unlock()
						rf.heartbeatCh <- true // 向其他节点发送心跳信号
						fmt.Printf("[%d]成为Leader...\n", rf.me)
						fmt.Printf("Leader[%d]发送心跳信号\n", rf.me)
					}
				}
			}
		}

		// 不超时且票数过半且当前有领导
		if timeout+last < milliSeconds() && votes > raftCount/2 && rf.curLeader > -1 {
			break
		} else { // 没有选出leader
			select {
			case <-time.After(time.Duration(50 * time.Millisecond)):
			}
		}
	}

	return success
}

// 成为领导者
func (rf *Raft) becomeLeader() {
	rf.state = stateLeader
	rf.curLeader = rf.me
}

// 修改节点为candidate状态
func (rf *Raft) becomeCandidate() {
	rf.state = stateCandidate  // 状态设置为候选人
	rf.setTerm(rf.curTerm + 1) // 任期加一
	rf.voteFor = rf.me         // 为哪个节点（自己）投票
	rf.curLeader = -1          // 没有领导
}

// 发送心跳信号
// 只考虑leader没有挂的情况
func (rf *Raft) sendLeaderHeartbeat() {
	for {
		select {
		case <-rf.heartbeatCh: // 收到心跳
			rf.sendAppendEntryImpl()
		}
	}
}

// 给leader返回确认心跳
func (rf *Raft) sendAppendEntryImpl() {
	if rf.curLeader == rf.me {
		var count = 0                    // 确认信号的个数
		for i := 0; i < raftCount; i++ { // 设置返回确认信号的子节点
			if i != rf.me { // 当前不是本节点
				go func() {
					rf.heartbeatRespCh <- true // 子节点有返回
				}()
			}
		}

		for i := 0; i < raftCount; i++ { // 统计返回确认信号的子节点
			select {
			case ok := <-rf.heartbeatRespCh:
				if ok {
					count++
					if count > raftCount/2 {
						fmt.Printf("投票选举成功，校验心跳成功.\n")
						//log.Fatal("程序结束")
						shutdownCh <- struct{}{}
					}
				}
			}
		}
	}
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

// 启动节点
func (rf *Raft) Start() {
	go rf.election()
	go rf.sendLeaderHeartbeat()
}

func main() {
	// 创建三个节点
	nodes := make([]*Raft, 0, raftCount)
	for i := 0; i < raftCount; i++ {
		nodes = append(nodes, MakeNode(i))
	}

	for i := 0; i < raftCount; i++ {
		nodes[i].Start()
	}

	<-shutdownCh
}
