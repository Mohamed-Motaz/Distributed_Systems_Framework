package Cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
}

//Cache is a map of {ClientID, CacheValue}

type CacheValue struct {
	ServerID            string
	FinishedJobsResults []string
}
