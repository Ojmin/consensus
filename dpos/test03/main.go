package main

import (
	"consensus/dpos/test03/core"
	"consensus/dpos/test03/database"
	"consensus/dpos/test03/utils"
	"fmt"
	"log"
	"sync"
	"time"
)

// 声明一些持币者
var (
	coiners []core.Coiner
)

func main() {
	var (
		mutex sync.Mutex
	)
	utils.LoadEnv(".env") // 加载配置文件
	port := utils.GetEnvValue("PORT")
	core.NodeAddress = fmt.Sprintf("localhost:%s", port)
	log.Println("开启服务节点：", core.NodeAddress)

	core.InitCoiners(utils.CoinNum)  // 初始化持币者
	db, err := database.InitDB(port) // 连接数据库
	if err != nil {
		log.Panic(err)
	}
	chain := core.CreateGenesisBlock(db) // 创建区块链
	// delegatesNum := core.GetTotalNumOfDelegates(chain) // 获得代理数量
	lastHeight := chain.GetLastHeight() // 获得最大高度
	delegate := &core.Delegate{         // 创建代理
		Address:    core.NodeAddress,
		LastHeight: lastHeight,
		Candidates: 0,
		Votes:      0,
		IsForger:   false,
		Supporters: make(map[string]core.Coiner),
	}
	core.AddDelegate(chain, delegate, lastHeight)
	core.SendToDelegates(chain, delegate) // 将自己是候选人广播
	go core.GenerateTx(chain)

	core.RoundVote(coiners, chain, &mutex) // 循环投票
	delegates := core.GetAllDelegates(chain)
	core.SortByVotes(delegates) // 排序

	if len(delegates) >= utils.LimitDelegateNum {
		time.Sleep(time.Second * 3)
		core.DelegateFlag = true
	} else {
		log.Println("受托人人数不够")
	}

	core.StartServer(chain) // 启动服务
}
