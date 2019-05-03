package core

import (
	"bytes"
	"consensus/dpos/test03/database"
	"consensus/dpos/test03/utils"
	"encoding/gob"
	"github.com/boltdb/bolt"
	"log"
	"sort"
)

// 代理
type Delegate struct {
	Address    string
	LastHeight int64
	Candidates int               // 受托人数
	Votes      float64           // 票数
	IsForger   bool              // 是否是受托人状态
	Supporters map[string]Coiner //
}

// 代理人数组
type DelegateSlice []*Delegate

func (s DelegateSlice) Len() int           { return len(s) }
func (s DelegateSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s DelegateSlice) Less(i, j int) bool { return s[i].Votes < s[j].Votes }

// 根据投票数排序
func SortByVotes(delegates []*Delegate) []*Delegate {
	ds := DelegateSlice(delegates)
	sort.Stable(ds) // 相等票数的相对顺序不变
	return ds
}

// 新增或更新候选人
func AddDelegate(chain *BlockChain, delegate *Delegate, lastHeight int64) bool {
	var isAdded bool
	if err := chain.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		dData := bucket.Get([]byte(delegate.Address)) // 查询代理是否存在
		if dData != nil {                             // 代理存在
			d := DeserializeDelegate(dData) // 反序列化
			if d.LastHeight < lastHeight {
				delegate.IsForger = d.IsForger
				if err := bucket.Put([]byte(delegate.Address), utils.Serialize(delegate)); err != nil {
					log.Panic(err)
				}
			}
		} else { // 代理不存在
			if err := bucket.Put([]byte(delegate.Address), utils.Serialize(delegate)); err != nil {
				log.Panic(err)
			}
			isAdded = true
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
	return isAdded
}

// 反序列化Delegate
func DeserializeDelegate(data []byte) *Delegate {
	var delegate Delegate
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&delegate); err != nil {
		log.Panic(err)
	}
	return &delegate
}

// 更新受托人
func UpdateDelegate(bc *BlockChain, delegate *Delegate) {
	if err := bc.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		addrData := []byte(delegate.Address)
		dData := bucket.Get(addrData)
		if dData != nil {
			if err := bucket.Put(addrData, utils.Serialize(delegate)); err != nil {
				log.Panic(err)
			}
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
}

// 删除受托人
func DeleteDelegate(bc *BlockChain, address []byte) {
	if err := bc.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		if err := bucket.Delete(address); err != nil {
			log.Panic(err)
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
}

// 获取某个受托人
func GetDeledate(bc *BlockChain, address []byte) *Delegate {
	var delegate Delegate
	if err := bc.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		dData := bucket.Get(address)
		if dData != nil {
			delegate = *DeserializeDelegate(dData)
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
	return &delegate
}

// 获取所有的候选人
func GetAllDelegates(bc *BlockChain) []*Delegate {
	delegates := make([]*Delegate, 0)
	if err := bc.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			delegates = append(delegates, DeserializeDelegate(v))
		}
		return nil
	}); err != nil {
		log.Panic(err)
	}
	return delegates
}

// 获取受托人总数
func GetTotalNumOfDelegates(bc *BlockChain) int {
	delegates := GetAllDelegates(bc)
	return len(delegates)
}
