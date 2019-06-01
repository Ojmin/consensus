package main

import (
	"fmt"
	"net/rpc"
)

//
type Params struct {
	Width, Height int
}

func main() {
	client, err := rpc.DialHTTP("tcp", "192.168.1.8:9000")
	if err != nil {
		fmt.Println("dial http error:", err)
	}

	perimeter, area := 0, 0
	params := Params{25, 30}

	// Area
	err = client.Call("Rect.Area", params, &area)
	if err != nil {
		fmt.Println("call Area error:", err)
	}
	fmt.Println("面积为:", area)

	// Perimeter
	err = client.Call("Rect.Perimeter", params, &perimeter)
	if err != nil {
		fmt.Println("call Area error:", err)
	}
	fmt.Println("周长为:", perimeter)

}
