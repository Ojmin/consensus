package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	raftCount = 2

	FOLLOWER = iota
	CANDIDATE
	LEADER
)

var (
	clientWriter  http.ResponseWriter
	leader        = Leader{Term: 0, Id: -1}
	mysqlMessage  = make(map[string]string) // 处理数据库信息
	messageId     = 1                       // 消息数组下标
	nodeTable     = make(map[string]string) // 存储每个节点中的键值对
	bufferMessage = make(map[string]string)
)

//
type nodeInfo struct {
	id   string
	port string
}

//
type Leader struct {
	Term int
	Id   int
}

//
type Raft struct {
	node            nodeInfo
	mu              sync.Mutex
	me              int
	currentTerm     int
	votedFor        int
	state           int
	timeout         int
	currentLeader   int
	lastMessageTime int64
	messageCh       chan bool
	electCh         chan bool // 发送选举
	heartbeatCh     chan bool // 发送心跳
	heartbeatReCh   chan bool // 返回心跳
}

//
type Param struct {
	Msg   string
	MsgId string
	Arg   Leader
}

//
func randomRange(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}

//
func Make(id int) *Raft {
	rf := Raft{}
	rf.me = id
	rf.votedFor = -1
	rf.state = FOLLOWER
	rf.timeout = 0
	rf.currentLeader = -1
	rf.setTerm(0)
	rf.messageCh = make(chan bool)
	rf.electCh = make(chan bool)
	rf.heartbeatCh = make(chan bool)
	rf.heartbeatReCh = make(chan bool)

	go rf.election()
	go rf.sendLeaderHeartbeat()

	return &rf
}

//
func (rf *Raft) setTerm(term int) {
	rf.currentTerm = term
}

//
func (rf *Raft) election() {
	var result bool
	for {
		timeout := randomRange(1500, 3000)
		rf.lastMessageTime = milliSeconds()
		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
		}
		if !result {
			result = rf.electionOneRound(&leader)
		}
	}

}

//
func (rf *Raft) electionOneRound(args *Leader) bool {
	if args.Id > -1 && args.Id != rf.me {
		fmt.Printf("%d已经是leader,终止%d选举\n", args.Id, rf.me)
		return true
	}

	var (
		timeout          int64 = 2000
		votes                  = 0
		triggerHeartbeat       = false
		last                   = milliSeconds()
		success                = false
	)

	rf.mu.Lock()
	rf.becomeCandidate()
	rf.mu.Unlock()

	fmt.Printf("candidate=%d start electing leader\n", rf.me)
	for {
		fmt.Printf("candidate=%d send request vote to server\n", rf.me)
		go func() {
			rf.broadcast(Param{Msg: "send request vote"}, "Raft.ElectingLeader", func(ok bool) {
				rf.electCh <- ok
			})
		}()

		for i := 0; i < raftCount-1; i++ {
			fmt.Printf("candidate=%d waiting for select for i=%d\n", rf.me, i)
			select {
			case ok := <-rf.electCh:
				if ok {
					votes++
					success = votes >= raftCount || rf.currentLeader > -1
					if success && !triggerHeartbeat {
						fmt.Println("ok...", args)
						triggerHeartbeat = true
						rf.mu.Lock()
						rf.becomeLeader()
						args.Term = rf.currentTerm + 1
						args.Id = rf.me
						rf.mu.Unlock()
						fmt.Printf("candidate=%d becomes leader\n", rf.currentLeader)
						rf.heartbeatCh <- true
					}
				}

			}

			fmt.Printf("candidate=%d complete for selecting for i=%d\n", rf.me, i)
		}

		if timeout+last < milliSeconds() || votes >= raftCount/2 || rf.currentLeader > -1 {
			break
		} else {
			select {
			case <-time.After(time.Duration(500) * time.Millisecond):
			}
		}
	}

	return success
}

//
func (rf *Raft) sendLeaderHeartbeat() {
	for {
		select {
		case <-rf.heartbeatCh:
			rf.sendAppendEntryImpl()
		}
	}

}

//
func (rf *Raft) sendAppendEntryImpl() {
	if rf.currentLeader == rf.me {
		var count = 0
		go func() {
			param := Param{
				Msg: "leader heartbeat",
				Arg: Leader{Term: rf.currentTerm, Id: rf.me},
			}
			rf.broadcast(param, "Raft.Heartbeat", func(ok bool) {
				rf.heartbeatCh <- ok
			})
		}()

		for i := 0; i < raftCount-1; i++ {
			select {
			case ok := <-rf.heartbeatCh:
				if ok {
					count++
					if count >= raftCount/2 {
						rf.mu.Lock()
						rf.lastMessageTime = milliSeconds()
						fmt.Println("接收到了子节点的返回信息")
						rf.mu.Unlock()
					}
				}
			}
		}
	}

}

//
func (rf *Raft) becomeLeader() {
	rf.state = LEADER
	fmt.Printf("%d成为了leader.", rf.me)
	rf.currentLeader = rf.me
}

//
func (rf *Raft) becomeCandidate() {
	if rf.state == 0 || rf.currentLeader == -1 {
		rf.state = CANDIDATE
		rf.setTerm(rf.currentTerm + 1)
		rf.votedFor = rf.me
		rf.currentLeader = -1
	}

}

//
func (rf *Raft) getRequest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return
	}

	if len(r.Form["age"]) > 0 {
		clientWriter = w
		fmt.Println("主节点广播客户端请求age:", r.Form["age"][0])
	}
	param := Param{
		Msg:   r.Form["age"][0],
		MsgId: strconv.Itoa(messageId),
	}
	messageId++
	if leader.Id == rf.me {
		rf.sendMessageToOtherNodes(param)
	} else {
		leaderId := nodeTable[strconv.Itoa(leader.Id)]
		rpcClient, err := rpc.DialHTTP("tcp", "127.0.0.1"+leaderId)
		if err != nil {
			log.Fatal(err)
		}
		rpcClient.Call()
	}

}

//
func (rf *Raft) raftRegisterRPC(port string) {
	err := rpc.Register(rf)
	if err != nil {
		log.Println(err)
	}
	rpc.HandleHTTP()
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Println(err)
	}

}

//
func milliSeconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//
func (rf *Raft) sendMessageToOtherNodes(param Param) {
	bufferMessage[param.MsgId] = param.Msg
	if rf.currentLeader == rf.me { // 只有leader才能给其他节点发消息
		var count = 0
		fmt.Println("leader发送数据...")
		go func() {
			rf.broadcast(param, "Raft.LogDataCopy", func(ok bool) {
				rf.messageCh <- ok
			})
		}()

		for i := 0; i < raftCount-1; i++ {
			fmt.Println("等待其他节点回应...")
			select {
			case ok := <-rf.messageCh:
				if ok {
					count++
					if count >= raftCount/2 {
						rf.mu.Lock()
						rf.lastMessageTime = milliSeconds()
						mysqlMessage[param.MsgId] = bufferMessage[param.MsgId]
						delete(bufferMessage, param.MsgId)
						if clientWriter != nil {
							_, _ = fmt.Fprintf(clientWriter, "OK")
						}
						fmt.Println("leader发送数据结束")
						rf.mu.Unlock()
					}
				}
			}
		}
	}

}

//
func (rf *Raft) broadcast(param Param, path string, fun func(ok bool)) {
	for nodeId, port := range nodeTable {
		if nodeId == rf.node.id { // 不给自己广播
			continue
		}

		rpcClient, err := rpc.DialHTTP("tcp", "127.0.0.1"+port)
		if err != nil {
			fun(false)
			continue
		}

		var bo = false
		err = rpcClient.Call(path, param, &bo)
		if err != nil {
			fun(false)
			continue
		}
		fun(bo)
	}

}

//
func (rf *Raft) ElectingLeader(param Param, result *bool) error {
	*result = true
	rf.lastMessageTime = milliSeconds()

	return nil
}

//
func (rf *Raft) Heartbeat(param Param, result *bool) error {
	fmt.Println("\nrpc:heartbeat:", rf.me, param.Msg)
	if param.Arg.Term < rf.currentTerm {
		*result = false
	} else {
		leader := param.Arg
		fmt.Printf("%d收到leader%d的心跳\n", rf.currentLeader, leader.Id)
		*result = true
		rf.mu.Lock()
		rf.currentLeader = leader.Id
		rf.votedFor = leader.Id
		rf.state = FOLLOWER
		rf.lastMessageTime = milliSeconds()
		fmt.Printf("server=%d learned the leader=%d\n", rf.me, rf.currentLeader)
		rf.mu.Unlock()
	}

	return nil
}

//
func (rf *Raft) LogDataCopy(param Param, result *bool) error {
	fmt.Println("\nrpc:LogDataCopy:", rf.me, param.Msg)
	bufferMessage[param.MsgId] = param.Msg
	*result = true
	return nil
}

//
func main() {
	if len(os.Args) > 1 {
		userId := os.Args[1] // 接收终端输入信息
		id, _ := strconv.Atoi(userId)
		fmt.Println(id)
		nodeTable = map[string]string{ // 定义节点ID和端口
			"1": "9000",
			"2": "9001",
		}
		node := nodeInfo{id: userId, port: nodeTable[userId]}
		rf := Make(id)
		rf.node = node

		go func() {
			rf.raftRegisterRPC(node.port) // 注册rpc
		}()

		if userId == "1" {
			go func() {
				http.HandleFunc("/req", rf.getRequest)
				fmt.Println("监听8080")
				if err := http.ListenAndServe(":8080", nil); err != nil {
					fmt.Println(err)
					return
				}
			}()
		}
	}

}
