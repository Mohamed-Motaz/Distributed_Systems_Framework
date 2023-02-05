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

type FinishedJob struct {
	JobId                string    `json:"jobId"`
	JobResult            string    `json:"jobResult"`
}

type CacheValue struct {
	ServerID     string        `json:"serverId"`
	FinishedJobs []FinishedJob `json:"finishedJobs"`
}
