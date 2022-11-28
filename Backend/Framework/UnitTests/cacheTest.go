package main

import (
	c "Framework/Cache"
	logger "Framework/Logger"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {

	cacheObj := c.NewCache(c.CreateCacheAddress("LocalHost", "6379"))

	err := cacheObj.Set("agina", "1", time.Second)
	if err != nil {
		logger.LogError(logger.CACHE, logger.ESSENTIAL, "error while setting key agina with value 1 %+v", err)
	}

	//time.Sleep(time.Second)
	
	val, err := cacheObj.Get("agina")
	if err == redis.Nil {
		logger.LogError(logger.CACHE, logger.ESSENTIAL, "key agina not present")
	} else if err != nil {
		logger.LogError(logger.CACHE, logger.ESSENTIAL, "shit there is a real error in cache %+v", err)
	} else {
		logger.LogInfo(logger.CACHE, logger.ESSENTIAL, "This is the value %+v", val)
	}

	time.Sleep(time.Second * 60)

}
