package core

import (
	"consensus/dpos/test03/database"
	"consensus/dpos/test03/utils"
	"github.com/boltdb/bolt"
	"log"
)

// 区块链
type BlockChain struct {
	DB   *bolt.DB
	Hash []byte
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

		lastHash := bucket.Get([]byte(database.LastHash))
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
			if err := bucket.Put([]byte(database.LastHash), curtHash); err != nil {
				log.Panic(err)
			}
			bc.Hash = curtHash
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
}
