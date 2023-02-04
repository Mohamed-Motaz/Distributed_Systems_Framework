package main

import (
	database "Framework/Database"
	"Framework/RPC"
	utils "Framework/Utils"
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type LockServer struct {
	id            string
	db            *database.DBWrapper
	mxLateJobTime time.Duration
	mastersState  map[string]RPC.CurrentJobProgress // key -> masterId, value -> CJP 
}
type FolderName string

const (
	_MY_HOST string = "MY_HOST"
	_MY_PORT string = "MY_PORT"

	_DB_USER     string = "DB_USER"
	_DB_PASSWORD string = "DB_PASSWORD"
	_DB_PROTOCOL string = "DB_PROTOCOL"
	_DB_HOST     string = "DB_HOST"
	_DB_PORT     string = "DB_PORT"
	_DB_SETTINGS string = "DB_SETTINGS"

	_LOCAL_HOST string = "127.0.0.1"

	PROCESS_BINARY_FOLDER_NAME FolderName = "Process"

	DISTRIBUTE_BINARY_FOLDER_NAME FolderName = "Distribute"
	AGGREGATE_BINARY_FOLDER_NAME  FolderName = "Aggregate"
	BINARY_FILES_FOLDER_NAME      FolderName = "BinaryFiles"
	OPTIONAL_FILES_FOLDER_NAME    FolderName = "OptionalFiles"
)

var (
	MyHost     string
	MyPort     string
	DbUser     string
	DbPassword string
	DbProtocol string
	DbHost     string
	DbPort     string
	DbSettings string
)

func init() {
	if !utils.IN_DOCKER {
		err := godotenv.Load("config.env")
		if err != nil {
			log.Fatal("Error loading config.env file")
		}
	}
	MyHost = strings.Replace(utils.GetEnv(_MY_HOST, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	MyPort = utils.GetEnv(_MY_PORT, "7777")
	DbUser = utils.GetEnv(_DB_USER, "root")
	DbPassword = utils.GetEnv(_DB_PASSWORD, "TheRootPassword1234")
	DbProtocol = utils.GetEnv(_DB_PROTOCOL, "tcp")
	DbHost = strings.Replace(utils.GetEnv(_DB_HOST, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	DbPort = utils.GetEnv(_DB_PORT, "3306")
	DbSettings = utils.GetEnv(_DB_SETTINGS, "charset=utf8mb4&parseTime=True&loc=Local")
}
