package Cache

import (
	logger "Framework/Logger"
	utils "Framework/Utils"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

//create a new Cache object
func NewCache(address string) *Cache {
	cache := &Cache{
		client: redis.NewClient(&redis.Options{
			Addr:     address,
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
		ctx: context.Background(),
	}

	//try to reconnect and sleep 10 seconds on failure
	var err error = fmt.Errorf("error")
	ctr := 0
	for ctr < 3 && err != nil {
		_, err := cache.client.Ping(cache.ctx).Result()
		ctr++
		if err != nil {
			time.Sleep(time.Second * 10)
		}
	}

	if err != nil {
		logger.FailOnError(logger.CACHE, logger.ESSENTIAL, "Unable to connect to caching layer with error %+v", err)
	} else {
		logger.LogInfo(logger.CACHE, logger.ESSENTIAL, "Successfully connected to caching layer")
	}

	if !utils.IN_DOCKER {
		//just for testing
		go cache.debug()
	}

	return cache
}

func CreateCacheAddress(host string, port string) string {
	return host + ":" + port
}

//gets the value of a key
func (c *Cache) Get(key string) (string, error) {
	return c.client.Get(c.ctx, key).Result()
}

//sets the value of a key
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, value, expiration).Err()
}

//prints out all cache contents every x amount of seconds
func (cache *Cache) debug() {
	time.Sleep(10 * time.Second)
	for {
		fmt.Println("About to print all redis keys and values\n")
		iter := cache.client.Scan(cache.ctx, 0, "*", 0).Iterator()
		for iter.Next(cache.ctx) {
			key := iter.Val()
			val, _ := cache.Get(key)
			fmt.Printf("%v -- %v\n", key, val)
		}
		if err := iter.Err(); err != nil {
			panic(err)
		}
		fmt.Println("\nDone printing all redis keys and values\n")

		time.Sleep(10 * time.Second)
	}

}
