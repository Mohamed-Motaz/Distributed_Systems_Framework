package Database

import (
	utils "Framework/Utils"
	"log"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type DBWrapper struct {
	Db *gorm.DB
}

const (
	_DB_USER     string = "DB_USER"
	_DB_PASSWORD string = "DB_PASSWORD"
	_DB_PROTOCOL string = "DB_PROTOCOL"
	_DB_HOST     string = "DB_HOST"
	_DB_PORT     string = "DB_PORT"
	_DB_SETTINGS string = "DB_SETTINGS"
	_LOCAL_HOST  string = "127.0.0.1"
)

var DbUser string
var DbPassword string
var DbProtocol string
var DbHost string
var DbPort string
var DbSettings string

func init() {
	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatal("Error loading config.env file")
	}
	DbUser = utils.GetEnv(_DB_USER, "root")
	DbPassword = utils.GetEnv(_DB_PASSWORD, "TheRootPassword1234")
	DbProtocol = utils.GetEnv(_DB_PROTOCOL, "tcp")
	DbHost = utils.GetEnv(_DB_HOST, "localhost")
	DbPort = utils.GetEnv(_DB_PORT, "3306")
	DbSettings = utils.GetEnv(_DB_SETTINGS, "charset=utf8mb4&parseTime=True&loc=Local")
}
