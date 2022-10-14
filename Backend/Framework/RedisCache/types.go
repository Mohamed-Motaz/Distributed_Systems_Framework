package RedisCache

import (
	utils "Server/Utils"
	"context"
	"log"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

//a wrapper for redis

type Cache struct {
	client *redis.Client   //caching layer
	ctx    context.Context //context for redis
}

const (
	CACHE_PORT string = "CACHE_PORT"
	CACHE_HOST string = "CACHE_HOST"
	LOCAL_HOST string = "127.0.0.1"
)

var CacheHost string
var CachePort string

func init() {
	if !utils.IN_DOCKER {
		err := godotenv.Load("config.env")
		if err != nil {
			log.Fatal("Error loading config.env file")
		}
	}

	CacheHost = strings.Replace(utils.GetEnv(CACHE_HOST, LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	CachePort = utils.GetEnv(CACHE_PORT, "6379")

}
