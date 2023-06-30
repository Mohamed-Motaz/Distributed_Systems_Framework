package main

import (
	utils "Framework/Utils"
	"log"
	"strings"

	"github.com/joho/godotenv"
)

type Worker struct {
	id               string
	ProcessBinary    utils.RunnableFile
	OptionalFilesZip utils.File
	JobId            string
}

const (
	_MASTER_HOST     string = "MASTER_HOST"
	_MASTER_PORT     string = "MASTER_PORT"
	_FILES_TO_IGNORE string = "FILES_TO_IGNORE"
	_LOCAL_HOST      string = "127.0.0.1"
)

var (
	MasterHost        string
	MasterPort        string
	FileNamesToIgnore []string
)

func init() {
	if !utils.IN_DOCKER {
		err := godotenv.Load("config.env")
		if err != nil {
			log.Fatal("Error loading config.env file")
		}
	}

	MasterHost = strings.Replace(utils.GetEnv(MasterHost, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	MasterPort = utils.GetEnv(_MASTER_PORT, "5555")
	FileNamesToIgnore = strings.Split(utils.GetEnv(_FILES_TO_IGNORE, ""), ",")
}
