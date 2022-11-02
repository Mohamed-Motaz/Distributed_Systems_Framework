package main

import (
	logger "Framework/Logger"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

func NewMaster() *Master {
	master := &Master{
		//q: mq.NewMQ(mq.CreateMQAddress(MqUsername, MqPassword, MqHost, MqPort)),
	}

	//start any go routines here
	go master.server()
	return master
}

//
// start a thread that listens for RPCs
//
func (master *Master) server() error {
	rpc.Register(master)
	rpc.HandleHTTP()

	masterAddr := createMasterAddress()
	os.Remove(masterAddr)
	listener, err := net.Listen("tcp", masterAddr)

	if err != nil {
		logger.FailOnError(logger.MASTER, logger.ESSENTIAL, "Error while listening on socket: %v", err)
	} else {
		logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Listening on socket: %v", masterAddr)
	}

	go http.Serve(listener, nil)
	return nil
}

func createMasterAddress() string {
	return MyHost + ":" + MyPort
}
