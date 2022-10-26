package MessageQueue

import (
	utils "Framework/Utils"
	"log"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

//queue names
const ASSIGNED_JOBS_QUEUE = "assignedJobs"
const FINISHED_JOBS_QUEUE = "finishedJobs"

//objects passed into and out of messageQ

type MQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	qMap map[string]*amqp.Queue
	mu   sync.Mutex
}
type AssignedJob struct {
	ClientId string `json:"clientId"`
	JobId    string `json:"jobId"`
	Content  string `json:"content"`
}
type FinishedJob struct {
	ClientId string `json:"clientId"`
	JobId    string `json:"jobId"`
	Content  string `json:"content"`
	Result   string `json:"result"`
}

const (
	_MQ_PORT     string = "MQ_PORT"
	_MQ_HOST     string = "MQ_HOST"
	_LOCAL_HOST  string = "127.0.0.1"
	_MQ_USERNAME string = "MQ_USERNAME"
	_MQ_PASSWORD string = "MQ_PASSWORD"
)

var MqHost string
var MqPort string
var MqUsername string
var MqPassword string

func init() {
	if !utils.IN_DOCKER {
		err := godotenv.Load("config.env")
		if err != nil {
			log.Fatal("Error loading config.env file")
		}
	}
	MqHost = strings.Replace(utils.GetEnv(_MQ_HOST, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	MqPort = utils.GetEnv(_MQ_PORT, "5672")
	MqUsername = utils.GetEnv(_MQ_USERNAME, "guest")
	MqPassword = utils.GetEnv(_MQ_PASSWORD, "guest")
}
