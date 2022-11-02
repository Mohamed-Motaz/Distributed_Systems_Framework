package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

//run the worker --> go run .
func main() {

	w := NewWorker()

	fmt.Printf("This is the newly created worker %+v\n", w)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	//now we should clean up
}
