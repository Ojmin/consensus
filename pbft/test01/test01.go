package main

import "fmt"

//
type Node struct {
	Name string
	Data int
}

//
var nodes = [4]Node{
	{"a", 0},
	{"b", 0},
	{"c", 0},
	{"d", 0},
}

func main() {
	nodes[0].Data = 1
	for i := 0; i < len(nodes); i++ {
		if i == 3 {
			continue
		}
		for j := 0; j < len(nodes); j++ {
			if i == j {
				continue
			}
			nodes[j].Data = nodes[j].Data + 1
		}
		fmt.Println(nodes)
	}
}
