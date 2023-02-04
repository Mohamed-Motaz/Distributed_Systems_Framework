package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// run the lock server --> go run .
func main() {

	m := NewLockServer()

	fmt.Printf("This is the newly created lock server %+v\n", m)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	//now we should clean up
}
