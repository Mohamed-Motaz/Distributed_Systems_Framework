package main

import (
	database "Framework/Database"
	utils "Framework/Utils"
	"log"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type LockServer struct {
	id              string
	databaseWrapper *database.DBWrapper
	address         string // to enable master connect with the lock server
	mu              sync.Mutex
}

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
)

var MyHost string
var MyPort string
var DbUser string
var DbPassword string
var DbProtocol string
var DbHost string
var DbPort string
var DbSettings string

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
	DbHost = strings.Replace(utils.GetEnv(_DB_HOST,_LOCAL_HOST ), "_", ".", -1) //replace all "_" with ".""
	DbPort = utils.GetEnv(_DB_PORT, "3306")
	DbSettings = utils.GetEnv(_DB_SETTINGS, "charset=utf8mb4&parseTime=True&loc=Local") 
}
