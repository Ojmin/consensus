package core

import (
	"bytes"
	"consensus/dpos/test03/utils"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"
)

const (
	protocol = "tcp" // 协议名
	Length   = 12    // 读数据长度
)

var (
	NodeAddress  string  // 开启节点地址
	DelegateFlag = false // 是否是受托人
	BlockFlag    = true  // 是否可以产块
)

// 开启服务
func StartServer(chain *BlockChain) {
	log.Println("开启服务的节点地址：", NodeAddress)
	listener, err := net.Listen(protocol, NodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer listener.Close()
	//循环出块
	go Forge(chain)
	// 接收广播
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConn(conn, chain)
	}
}

// 接收处理连接
func handleConn(conn net.Conn, chain *BlockChain) {
	time.Sleep(time.Second)
	var businessType = ""
	switch businessType {
	case utils.Block:
		handleBlock(conn, chain)
	case utils.Delegate:
		handleDelegate(conn, chain)
	case utils.Transfer:
		handleTx(conn, chain)
	default:
		log.Println("连接错误")
	}
}

// 处理交易
func handleTx(conn net.Conn, chain *BlockChain) {
	req, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	data := utils.Serialize(req[Length:])
	tx := DeserializeTx(data)
	AddTx(chain.DB, tx)
	SendTx(chain, tx)
}

// 发送交易
func SendTx(chain *BlockChain, tx *Transaction) {
	data := utils.Serialize(tx)
	reqData := append(utils.ConvertStrToBytes(utils.Transfer), data...) // 拼接请求数据
	delegates := GetAllDelegates(chain)
	for _, d := range delegates {
		if d.Address == NodeAddress {
			continue
		}
		sendData(d.Address, reqData)
	}
}

// 处理代理
func handleDelegate(conn net.Conn, chain *BlockChain) {
	req, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	data := utils.Serialize(req[Length:])
	delegate := DeserializeDelegate(data)
	lastHeight := chain.GetLastHeight()
	ok := AddDelegate(chain, delegate, lastHeight)
	if ok { // 添加成功广播到其他节点
		SendToDelegates(chain, delegate)
	} else {
		log.Panic("添加代理失败")
	}
}

// 发送到其他代理
func SendToDelegates(chain *BlockChain, delegate *Delegate) {
	delegates := GetAllDelegates(chain)
	for _, d := range delegates {
		if d.Address == delegate.Address {
			continue
		}
		sendDelegate(d.Address, delegate)
	}
}

// 发送代理
func sendDelegate(addr string, delegate *Delegate) {
	data := utils.Serialize(delegate)
	req := append(utils.ConvertStrToBytes(utils.Delegate), data...)
	sendData(addr, req)
}

// 处理区块
func handleBlock(conn net.Conn, chain *BlockChain) {
	req, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	blockData := utils.Serialize(req[Length:])
	block := DeserializeBlock(blockData)
	if DelegateFlag { // 是受托人，收到广播后停止出块
		BlockFlag = false
	}
	isOk := ValidateBlock(chain.GetLastBlock(), block)
	if isOk {
		chain.AddBlock(block)
		log.Println("验证通过：", block.Hash)
		Broadcast(chain, block) // 广播给其他候选人
	} else { // 验证失败
		log.Println("验证区块失败：", block.Hash)
	}
	BlockFlag = true
}

// 产块
func Forge(chain *BlockChain) {
	// 循环出块
	for {
		time.Sleep(time.Second * 3)
		if DelegateFlag && BlockFlag {
			txs := GetAllTxs(chain.DB)                              // 得到所有交易
			block := NewBlock(chain.GetLastBlock(), txs, BlockFlag) // 创建新区块
			if block != nil {
				chain.AddBlock(block)
				Broadcast(chain, block)
			}
		}
	}
}

// 广播区块
func Broadcast(chain *BlockChain, block *Block) {
	delegates := GetAllDelegates(chain)
	for _, d := range delegates {
		if d.Address == NodeAddress {
			continue
		}
		sendBlock(d.Address, block)
	}
}

// 发送区块
func sendBlock(address string, block *Block) {
	data := utils.Serialize(block)
	reqData := append(utils.ConvertStrToBytes(utils.Block), data...) // 拼接请求数据
	sendData(address, reqData)
}

// 发送数据
func sendData(address string, data []byte) {
	conn, err := net.Dial(protocol, address)
	if err != nil {
		log.Panic(err)
	}
	if _, err := io.Copy(conn, bytes.NewBuffer(data)); err != nil {
		log.Panic(err)
	}
}

//
func GenerateTx(chain *BlockChain) {
	for i := 0; i < 50; i++ {
		tx := &Transaction{
			Id:         "id:" + utils.GetUuid(),
			From:       "from:zhangsan",
			To:         "to:lisi",
			Amount:     float64(20.00),
			TransferBy: "较易发生地址:" + NodeAddress,
		}
		AddTx(chain.DB, tx)
		log.Println("交易信息：", tx)
		// 广播交易到其他节点
		SendTx(chain, tx)
	}
}
