package MessageQueue

import (
	utils "Framework/Utils"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

//todo  handle the lifetime for a job in the message queue

// queue names
const (
	ASSIGNED_JOBS_QUEUE = "assignedJobs"
	FINISHED_JOBS_QUEUE = "finishedJobs"
)

//objects passed into and out of messageQ

type MQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	qMap map[string]*amqp.Queue
	mu   sync.Mutex
}
type AssignedJob struct {
	ClientId           string    `json:"clientId"`
	JobId              string    `json:"jobId"`
	JobContent         string    `json:"jobContent"`
	DistributeBinaryId string    `json:"distributeBinaryId"`
	ProcessBinaryId    string    `json:"processBinaryId"`
	AggregateBinaryId  string    `json:"aggregateBinaryId"`
	CreatedAt          time.Time `json:"createdAt"`
}

type FinishedJob struct {
	ClientId     string    `json:"clientId"`
	JobId        string    `json:"jobId"`
	Content      string    `json:"content"`
	Result       string    `json:"result"`
	CreatedAt    time.Time `json:"createdAt"`
	TimeAssigned time.Time `json:"timeAssigned"`
	utils.Error
}
