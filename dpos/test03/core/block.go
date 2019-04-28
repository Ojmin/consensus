package core

import (
	"bytes"
	"consensus/dpos/test03/utils"
	"encoding/gob"
	"log"
)

// 区块
type Block struct {
	// 区块头
	Index     int64
	Timestamp int64
	Hash      string
	PrevHash  string

	// 区块体
	Txs []*Transaction
}

// 反序列化
func DeserializeBlock(data []byte) *Block {
	var b Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&b); err != nil {
		log.Panic(err)
	}
	return &b
}

// 计算哈希
func CalcHash(b *Block) string {
	str := string(b.Index) + b.PrevHash + string(b.Timestamp)
	return utils.CalculateHash(str)
}
