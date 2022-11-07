package main

import (
	logger "Framework/Logger"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	m, err := NewWebSocketServer()

	if err != nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to create websocketServer} -> error : %v", err)
		return
	}

	fmt.Printf("This is the newly created websocketServer %+v\n", m)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	//now we should clean up
}
