package main

import (
	"fmt"
)

//run using -> go run .
func main() {
	//run the master
	master := NewMaster()
	fmt.Printf("%+v", master)
}
