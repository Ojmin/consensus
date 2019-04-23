package dpos

import (
	"fmt"
	"math/rand"
	"time"
)

// 网络节点
var Nodes = make([]*Node, 10)

// 全节点
type Node struct {
	Name  string
	Votes int
}

// 代理产生新块
func (node *Node) NewBlock(lastBlock *Block, data []byte) *Block {
	b := &Block{
		Index:     lastBlock.Index + 1,
		PrevHash:  lastBlock.Hash,
		TimeStamp: time.Now().String(),
		Data:      data,
		Delegate:  node,
	}
	b.SetHash()
	return b
}

// 创建节点
func CreateNode() {
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("nodes[%d]", i)
		Nodes[i] = &Node{
			Name:  name,
			Votes: 0,
		}
	}
}

// 简单模拟投票
func Vote() {
	for i := 0; i < 10; i++ {
		rand.Seed(time.Now().UnixNano())
		time.Sleep(100000)
		votes := rand.Intn(10000)
		Nodes[i].Votes = votes // 每个节点的票数就是随机数
		fmt.Printf("nodes[%d]:%d\n", i, votes)
	}
}

// 选出票数最多的三个
func SortNodes() []*Node {
	n := Nodes
	for i := 0; i < len(n); i++ {
		for j := 0; j < len(n)-1; j++ {
			if n[j].Votes < n[j+1].Votes {
				n[j], n[j+1] = n[j+1], n[j]
			}
		}
	}
	return n[:3]
}
