package main

import (
	"fmt"
	"net/http"
	"net/rpc"
)

//
type Params struct {
	Width, Height int
}

//
type Rect struct{}

// Area
func (rect *Rect) Area(params Params, ret *int) error {
	*ret = params.Width * params.Height
	return nil
}

// Perimeter
func (rect *Rect) Perimeter(params Params, ret *int) error {
	*ret = 2 * (params.Width + params.Height)
	return nil
}

func main() {
	rect := new(Rect)

	if err := rpc.Register(rect); err != nil {
		fmt.Println("prc register error:", err)
	}

	rpc.HandleHTTP()

	if err := http.ListenAndServe(":9000", nil); err != nil {
		fmt.Println("listen and serve error:", err)
	}

}
