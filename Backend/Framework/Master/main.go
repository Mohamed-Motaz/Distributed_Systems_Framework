package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

//run the master --> go run .
func main() {

	m := NewMaster()

	fmt.Printf("This is the newly created master %+v\n", m)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	//now we should clean up
	
}
