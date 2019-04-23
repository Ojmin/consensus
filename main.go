package main

import (
	"consensus/dpos"
	"fmt"
)

func main() {
	dpos.CreateNode()
	for _, node := range dpos.Nodes {
		fmt.Println(node.Name)
	}
	fmt.Println("-------------------")
	dpos.Vote()
	producers := dpos.SortNodes()
	fmt.Println("-------------------")
	for _, node := range producers {
		fmt.Println(node.Name, node.Votes)
	}

	first := dpos.GenesisBlock()
	last := first
	for i := 0; i < len(producers); i++ {
		fmt.Printf("[%s %d] produce new block\n", producers[i].Name, producers[i].Votes)
		last = producers[i].NewBlock(last, []byte(fmt.Sprintf("new block %d", i)))
	}
}
