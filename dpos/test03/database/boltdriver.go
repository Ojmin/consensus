package database

import (
	"fmt"
	"github.com/boltdb/bolt"
)

// 将交易信息、候选人记录、区块等存到数据库

const (
	dbFile = "blockchain%s.db" // 数据库名

	// 表名
	BlocksBucket       = "blocks"
	DelegatesBucket    = "delegates"
	TransactionsBucket = "transactions"
	TipHash            = "tiphash"
)

// 初始化
func InitDB(nodeId string) (*bolt.DB, error) {
	df := fmt.Sprintf(dbFile, nodeId)
	db, err := bolt.Open(df, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open error:%v", err)
	}

	//建表
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(BlocksBucket)); err != nil {
			return fmt.Errorf("cannot create blocks bucket:%v", err)
		}
		return nil
	})
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(DelegatesBucket)); err != nil {
			return fmt.Errorf("cannot create delegates bucket:%v", err)
		}
		return nil
	})
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(TransactionsBucket)); err != nil {
			return fmt.Errorf("cannot create transactions bucket:%v", err)
		}
		return nil
	})

	return db, nil
}
