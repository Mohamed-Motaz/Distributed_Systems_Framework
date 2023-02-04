package Cache

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
}

type CacheValue struct{
	ServerID string
	FinishedJobsResults []string
}
