package Cache

import (
	utils "Framework/Utils"
	"context"
	"log"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
}

const (
	_CACHE_HOST string = "CACHE_HOST"
	_CACHE_PORT string = "CACHE_PORT"
	_LOCAL_HOST string = "127.0.0.1"
)

var CacheHost string
var CachePort string

func init() {
	if !utils.IN_DOCKER { //if i am not in docker, I will read the data from config.env
		err := godotenv.Load("config.env")
		if err != nil {
			log.Fatal("Error loading config.env file")
		}
	}

	CacheHost = strings.Replace(utils.GetEnv(_CACHE_HOST, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	CachePort = utils.GetEnv(_CACHE_PORT, "6379")
}
