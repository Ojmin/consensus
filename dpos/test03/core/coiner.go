package core

import (
	"consensus/dpos/test03/utils"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// 持币者
type Coiner struct {
	Address   string
	Amount    float64
	Candidate string
	IsVote    bool
}

// 初始化
func InitCoiners(num int) []Coiner {
	fmt.Println("开始初始化持币者")
	var (
		coiners []Coiner
		amount  = 1.23
	)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < num; i++ {
		coiner := Coiner{
			Address:   strconv.FormatInt(int64(i), 10),
			Amount:    utils.Decimal(amount * 10),
			Candidate: "",
			IsVote:    true,
		}
		amount += rand.Float64() // 增加持币量
		coiners = append(coiners, coiner)
	}
	fmt.Println("初始化持币者完成，个数为：", len(coiners))
	return coiners
}

// 持币者投票
func Vote(coiner Coiner, delegates []*Delegate, voteIndex int, m *sync.Mutex) *Coiner {
	if voteIndex < len(delegates) {
		m.Lock()
		candidate := delegates[voteIndex]             // 选择一个代理
		candidate.Votes += coiner.Amount              // 将coiner的持币量加给代理
		candidate.Supporters[coiner.Address] = coiner // 在代理中记录投票者

		coiner.Candidate = candidate.Address
		coiner.IsVote = true
		m.Unlock()
	}
	return &coiner
}

// 取消投票
func CancelVote(coiner Coiner, delegate Delegate, m *sync.Mutex) *Delegate {
	m.Lock()
	delegate.Votes -= coiner.Amount
	m.Unlock()
	return &delegate
}

// 循环投票
func RoundVote(coiners []Coiner, bc *BlockChain, m *sync.Mutex) {
	delegates := GetAllDelegates(bc)
	for _, coiner := range coiners {
		if len(delegates) >= utils.LimitDelegateNum && coiner.IsVote {
			index := rand.Intn(len(delegates))
			Vote(coiner, delegates, index, m)
			UpdateDelegate(bc, delegates[index])
		}
	}
}
