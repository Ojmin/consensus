package dpos

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"
)

var (
	blockchain []*Block                               // 区块链
	delegates  = []string{"aaa", "bbb", "ccc", "ddd"} // 委托人
)

func RandDelegates() {
	rand.Seed(time.Now().Unix())
	r := rand.Intn(3) // 0-2的随机数
	t := delegates[r]
	delegates[r] = delegates[3]
	delegates[3] = t
}

// 区块
type Block struct {
	Index     int
	TimeStamp string
	Hash      string
	PrevHash  string
	Data      []byte
	Delegate  *Node
}

func GenesisBlock() *Block {
	b := &Block{
		Index:     0,
		TimeStamp: time.Now().String(),
		Hash:      "",
		PrevHash:  "",
		Data:      []byte("genesis block"),
		Delegate:  nil,
	}
	b.SetHash()
	return b
}

func (b *Block) SetHash() {
	hashcode := b.PrevHash + b.TimeStamp + hex.EncodeToString(b.Data) + strconv.Itoa(b.Index)
	sha := sha256.New()
	sha.Write([]byte(hashcode))
	hashed := sha.Sum(nil)
	b.Hash = hex.EncodeToString(hashed)
}
