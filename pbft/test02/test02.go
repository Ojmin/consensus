package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

//
type NodeInfo struct {
	id     string
	path   string
	writer http.ResponseWriter
}

//
var nodeTable = make(map[string]string)

// 记录正常节点
var normalNodeMap = make(map[string]string)

//
var AuthSuccess = false

func (n *NodeInfo) authentication(r *http.Request) {
	if !AuthSuccess { // 第一次进入
		if len(r.Form["nodeid"][0]) > 0 {
			normalNodeMap[r.Form["nodeid"][0]] = "OK"
			if len(normalNodeMap) > len(normalNodeMap)/3 { // 达到拜占庭容错
				AuthSuccess = true
				n.broadcast(r.Form["wartime"][0], "/commit")
			}
		}
	}
}

func main() {
	userId := os.Args[1]
	fmt.Println(userId)
	nodeTable = map[string]string{
		"n0": "localhost:1110",
		"n1": "localhost:1111",
		"n2": "localhost:1112",
		"n3": "localhost:1113",
	}
	node := NodeInfo{id: userId, path: nodeTable[userId]}

	http.HandleFunc("/req", node.request)
	http.HandleFunc("/preprepare", node.preprepare)
	http.HandleFunc("/prepare", node.prepare)
	http.HandleFunc("/commit", node.commit)

	if err := http.ListenAndServe(node.path, nil); err != nil {
		fmt.Printf("listen and serve error: %v", err)
	}
}

//
func (n *NodeInfo) commit(w http.ResponseWriter, r *http.Request) {
	if w != nil {
		fmt.Println("拜占庭校验成功")
	}
	io.WriteString(n.writer, "OK")
}

//
func (n *NodeInfo) prepare(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("接收到广播", r.Form["wartime"][0])
	if len(r.Form["wartime"]) > 0 {

	}
}

//
func (n *NodeInfo) preprepare(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("接收到的广播", r.Form["wartime"][0])
	if len(r.Form["wartime"]) > 0 {
		n.broadcast(r.Form["wartime"][0], "/prepare")
	}
}

//
func (n *NodeInfo) request(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if len(r.Form["wartime"]) > 0 {
		n.writer = w
		fmt.Println("节点收到的参数信息为", r.Form["wartime"][0])
		// 广播数据
		n.broadcast(r.Form["wartime"][0], "/preprepare")
	}

}

// 节点广播
func (n *NodeInfo) broadcast(msg string, path string) {
	fmt.Println("广播 ", path)
	// 遍历所有节点进行广播
	for id, url := range nodeTable {
		// 判断是否是自己，如果是自己，跳出当此循环
		if id == n.id {
			continue
		}
		// 要进行分发的节点
		http.Get("http://" + url + path + "?wartime=" + msg + "&nodeid=" + n.id)
	}
}
