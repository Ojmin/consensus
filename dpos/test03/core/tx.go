package core

import (
	"bytes"
	"consensus/dpos/test03/database"
	"consensus/dpos/test03/utils"
	"encoding/gob"
	"github.com/boltdb/bolt"
	"log"
)

// 交易
type Transaction struct {
	Id         string
	From       string
	To         string
	Amount     float64
	TransferBy string
}

// 数据库中添加、删除、查询交易

// 添加交易
func AddTx(db *bolt.DB, transaction *Transaction) {
	if err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.TransactionsBucket)) // 交易表
		txid := []byte(transaction.Id)                           // 交易ID
		txdata := bucket.Get([]byte(txid))
		if txdata != nil { // 交易存在
			return nil
		}
		if err := bucket.Put(txid, utils.Serialize(txdata)); err != nil {
			log.Panic(err)
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
}

// 删除交易
func DelTx(db *bolt.DB, txid []byte) {
	if err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.TransactionsBucket))
		if err := bucket.Delete(txid); err != nil {
			log.Panic(err)
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
}

// 获取单笔交易
func GetTx(db *bolt.DB, txid []byte) *Transaction {
	var transaction *Transaction
	if err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.TransactionsBucket))
		if data := bucket.Get(txid); data != nil {
			transaction = DeserializeTx(data)
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
	return transaction
}

// 获得所有交易
func GetAllTxs(db *bolt.DB) []*Transaction {
	var all []*Transaction
	if err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.TransactionsBucket))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			t := DeserializeTx(v)
			all = append(all, t)
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}

	return all
}

// 反序列化
func DeserializeTx(data []byte) *Transaction {
	var tx Transaction
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&tx); err != nil {
		log.Panic(err)
	}
	return &tx
}
