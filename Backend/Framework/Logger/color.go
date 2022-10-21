package Logger

import (
	utils "Framework/Utils"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

func init() {
	if true { //runtime.GOOS == "windows" {
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Purple = ""
		Cyan = ""
		Gray = ""
		White = ""
	}
	if !utils.IN_DOCKER {
		err := godotenv.Load("config.env")
		if err != nil {
			log.Fatal("Error loading config.env file")
		}
	}

	tmp, err := strconv.ParseInt(GetEnv(DEBUG_LOGS, "0"), 10, 16)
	if err != nil {
		panic(err)
	}
	DebugLogs = int16(tmp)

	tmp, err = strconv.ParseInt(GetEnv(ESSENTIAL_LOGS, "1"), 10, 16)
	if err != nil {
		panic(err)
	}
	EssentialLogs = int16(tmp)
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
