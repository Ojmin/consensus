package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// 区块
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
	Validator string
}

var (
	Blockchain []Block                                // 区块链
	delegates  = []string{"aaa", "bbb", "ccc", "ddd"} // 委托人
)

// 生成区块
func GenerateBlock(old Block, bpm int, addr string) Block {
	b := Block{
		Index:     old.Index + 1,
		BPM:       bpm,
		PrevHash:  old.Hash,
		Timestamp: time.Now().String(),
		Validator: addr,
	}
	b.Hash = CalcHash(b)
	return b
}

// 计算哈希
func CalcHash(b Block) string {
	str := b.PrevHash + b.Timestamp + b.Validator + strconv.Itoa(b.BPM) + strconv.Itoa(b.Index)
	sha := sha256.New()
	sha.Write([]byte(str))
	return hex.EncodeToString(sha.Sum(nil))
}

// 打乱委托人顺序
func RandDelegates() {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(3) // 0-2的随机数
	t := delegates[r]
	delegates[r] = delegates[3]
	delegates[3] = t
}

func main() {
	fmt.Println(delegates)
	RandDelegates()
	fmt.Println(delegates)
	firstBlock := Block{}
	Blockchain = append(Blockchain, firstBlock)
	n := 0
	for {
		time.Sleep(time.Second * 3)
		nextBlock := GenerateBlock(firstBlock, 1, delegates[n])
		n++
		n %= 4
		Blockchain = append(Blockchain, nextBlock)
		firstBlock = nextBlock
		fmt.Println(Blockchain)
	}

}
