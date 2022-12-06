package main

import (
	logger "Framework/Logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	_, err := NewWebSocketServer()

	if err != nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to create websocketServer} -> error : %v", err)
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	//now we should clean up

}
