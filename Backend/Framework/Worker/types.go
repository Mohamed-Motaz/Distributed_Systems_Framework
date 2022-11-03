package main

import (
	utils "Framework/Utils"
	"log"
	"strings"

	"github.com/joho/godotenv"
)

type Worker struct {
	id string
}

const (
	_MASTER_HOST     string = "MASTER_HOST"
	_MASTER_PORT     string = "MASTER_PORT"
	_PROCESS_EXE_CMD string = "PROCESS_EXE_CMD"
	_LOCAL_HOST      string = "127.0.0.1"
)

var MasterHost string
var MasterPort string
var ProcessExeCmd string

func init() {
	if !utils.IN_DOCKER {
		err := godotenv.Load("config.env")
		if err != nil {
			log.Fatal("Error loading config.env file")
		}
	}

	MasterHost = strings.Replace(utils.GetEnv(MasterHost, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	MasterPort = utils.GetEnv(_MASTER_PORT, "5555")
	ProcessExeCmd = utils.GetEnv(_PROCESS_EXE_CMD, "ClientProcess.exe")

}
