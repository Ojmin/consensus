package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// 全节点
type Node struct {
	Name  string
	Votes int
}

// 区块
type Block struct {
	Index     int
	Timestamp string
	PrevHash  string
	Hash      string
	Data      []byte
	Delegate  *Node
}

// 创世区块
func GenesisBlock() Block {
	genesis := Block{
		Index:     0,
		Timestamp: time.Now().String(),
		Data:      []byte("genesis block"),
	}
	genesis.Hash = CalcHash(genesis)
	return genesis
}

// 计算哈希
func CalcHash(b Block) string {
	str := b.PrevHash + b.Timestamp + strconv.Itoa(b.Index) + hex.EncodeToString(b.Data)
	sha := sha256.New()
	sha.Write([]byte(str))
	return hex.EncodeToString(sha.Sum(nil))
}

// 生成新区块
func (node *Node) GenerateBlock(prev Block, data []byte) Block {
	b := Block{
		Index:     prev.Index + 1,
		Timestamp: time.Now().String(),
		PrevHash:  prev.Hash,
		Data:      data,
		Delegate:  node, // 产块的代理
	}
	b.Hash = CalcHash(b)
	return b
}

var (
	nodes = make([]Node, 10)
)

// 创建节点
func InitNodes() {
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("节点:%d, 票数:", i)
		nodes[i] = Node{Name: name, Votes: 0}
	}
}

// 简单模拟投票
func Votes() {
	for i := 0; i < 10; i++ {
		rand.Seed(time.Now().UnixNano())
		time.Sleep(1)
		votes := rand.Intn(10000)
		nodes[i].Votes = votes
		fmt.Printf("节点[%d], 票数[%d]\n", i, votes)
	}
}

// 选票数最多的前三
func SortNodes() []Node {
	n := nodes
	for i := 0; i < len(n); i++ {
		for j := 0; j < len(n)-1; j++ {
			if n[j].Votes < n[j+1].Votes {
				n[j], n[j+1] = n[j+1], n[j]
			}
		}
	}
	return n[:3]
}

func main() {
	InitNodes()
	fmt.Println("创建的节点列表:")
	fmt.Println(nodes)
	fmt.Println()
	fmt.Println("节点票数:")

	Votes()
	n := SortNodes()
	fmt.Println()
	fmt.Println("获胜者:")
	fmt.Println(n)
	fmt.Println()
	first := GenesisBlock()
	last := first
	fmt.Println("开始生成区块:")

	for i := 0; i < len(n); i++ {
		fmt.Printf("%s %d 生成新区块\n", n[i].Name, n[i].Votes)
		last = n[i].GenerateBlock(last, []byte(fmt.Sprintf("new block %d", i)))
	}
}
