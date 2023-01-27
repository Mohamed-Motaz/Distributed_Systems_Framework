package MessageQueue

import (
	utils "Framework/Utils"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

//todo  handle the lifetime for a job in the message queue

//queue names
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
	ClientId         string     `json:"clientId"`
	JobId            string     `json:"jobId"`
	JobContent       string     `json:"jobContent"`
	DistributeBinary utils.File `json:"distributeBinary"` //the content is not passed in the mq
	ProcessBinary    utils.File `json:"processBinary"`    //the content is not passed in the mq
	AggregateBinary  utils.File `json:"aggregateBinary"`  //the content is not passed in the mq
}
type FinishedJob struct {
	ClientId string `json:"clientId"`
	JobId    string `json:"jobId"`
	Content  string `json:"content"`
	Result   string `json:"result"`
	utils.Error
}
