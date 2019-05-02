package core

import (
	"consensus/dpos/test03/database"
	"consensus/dpos/test03/utils"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"log"
	"time"
)

// 区块链
type BlockChain struct {
	DB  *bolt.DB
	Tip []byte
}

// 添加区块到数据库
func (bc *BlockChain) AddBlock(b *Block) {
	if err := bc.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.BlocksBucket))
		curtHash := []byte(b.Hash)
		data := bucket.Get(curtHash)
		if data != nil { // 区块已在数据库中
			return nil
		}
		// 区块不存在
		data = utils.Serialize(b)                          // 序列化区块信息
		if err := bucket.Put(curtHash, data); err != nil { // 添加区块到数据库中
			log.Panic(err)
		}

		lastHash := bucket.Get([]byte(database.TipHash))
		var isPutLastHash = false
		if lastHash != nil {
			lastBlockData := bucket.Get(lastHash)
			lastBlock := DeserializeBlock(lastBlockData)
			if lastBlock.Index < b.Index {
				isPutLastHash = true
			}
		} else {
			isPutLastHash = true
		}
		if isPutLastHash {
			if err := bucket.Put([]byte(database.TipHash), curtHash); err != nil {
				log.Panic(err)
			}
			bc.Tip = curtHash
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
}

// 根据区块哈希获取区块
func (bc *BlockChain) GetBlock(blockHash []byte) (*Block, error) {
	var block Block
	if err := bc.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.BlocksBucket))
		blockData := bucket.Get(blockHash)
		if blockData == nil {
			return errors.New("block not exist")
		}
		block = *DeserializeBlock(blockData)
		return nil
	}); err != nil {
		log.Panic(err)
	}
	return &block, nil
}

// 获取所有区块
func GetAllBlocks(db *bolt.DB) []*Block {
	blocks := make([]*Block, 0)
	if err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.BlocksBucket))
		cursor := bucket.Cursor() // 遍历
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			b := DeserializeBlock(v)
			blocks = append(blocks, b)
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}

	return blocks
}

// 获取最新区块
func (bc *BlockChain) GetLastBlock() *Block {
	var lastBlock Block
	if err := bc.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.BlocksBucket))
		h := bucket.Get([]byte(database.TipHash))
		if h != nil {
			blockData := bucket.Get(h)
			lastBlock = *DeserializeBlock(blockData)
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}

	return &lastBlock
}

// 获取最大高度
func (bc *BlockChain) GetLastHeight() int64 {
	b := bc.GetLastBlock()
	return b.Index
}

// 生成创世区块
func CreateGenesisBlock(db *bolt.DB) *BlockChain {
	var curHash []byte
	if err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.BlocksBucket))
		h := bucket.Get([]byte(database.TipHash))
		if h == nil {
			genesis := Block{
				Index:     0,
				Timestamp: time.Now().Unix(),
			}
			h := CalcHash(&genesis)
			genesis.Hash = h
			curHash = []byte(h)
			if err := bucket.Put(curHash, utils.Serialize(genesis)); err != nil {
				log.Panic(err)
			}
			if err := bucket.Put([]byte(database.TipHash), curHash); err != nil {
				log.Panic(err)
			}
		} else {
			curHash = h
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
	return &BlockChain{
		DB:  db,
		Tip: curHash,
	}
}
