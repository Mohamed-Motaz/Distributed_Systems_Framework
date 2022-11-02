package main

import (
	logger "Framework/Logger"
	"os"
	"os/signal"
	"syscall"
)

//run using -> go run .
func main() {
	//run the master
	NewMaster()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	sig := <-signalCh //block until user exits
	logger.LogInfo(logger.SERVER, logger.ESSENTIAL, "Received a quit sig %+v\nCleaning up resources. Goodbye", sig)
}
