package Cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
}

//Cache is a map of {ClientID, CacheValue}

type FinishedJob struct {
	JobId        string    `json:"jobId"`
	JobResult    string    `json:"jobResult"`
	CreatedAt    time.Time `json: "createdAt"`
	TimeAssigned time.Time `json: "timeAssigned"`
}

// DONE: add createdAt and timeAssigned
type CacheValue struct {
	ServerID     string        `json: "serverId"`
	FinishedJobs []FinishedJob `json: "finishedJobs"`
}
