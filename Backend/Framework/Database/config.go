package Database

import (
	utils "Server/Utils"
	"log"

	"github.com/joho/godotenv"
)

//types
const (
	DB_USER     string = "DB_USER"
	DB_PASSWORD string = "DB_PASSWORD"
	DB_PROTOCOL string = "DB_PROTOCOL"
	DB_HOST     string = "DB_HOST"
	DB_PORT     string = "DB_PORT"
	DB_SETTINGS string = "DB_SETTINGS"
	LOCAL_HOST  string = "127.0.0.1"
)

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
	DbUser = utils.GetEnv(DB_USER, "root")
	DbPassword = utils.GetEnv(DB_PASSWORD, "instabug")
	DbProtocol = utils.GetEnv(DB_PROTOCOL, "tcp")
	DbHost = utils.GetEnv(DB_HOST, "localhost")
	DbPort = utils.GetEnv(DB_PORT, "3306")
	DbSettings = utils.GetEnv(DB_SETTINGS, "charset=utf8mb4&parseTime=True&loc=Local")
}
