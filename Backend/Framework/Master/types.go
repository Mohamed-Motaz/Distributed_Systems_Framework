package main

import (
	mq "Framework/MessageQueue"
	utils "Framework/Utils"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Master struct {
	id                string
	currentJob        string
	currentJobId      string
	currentTasks      []Task
	workersTimers     []WorkerAndHisTimer
	maxHeartBeatTimer time.Duration
	isRunning         bool //do i currently have a job
	mu                sync.Mutex
	q                 *mq.MQ
}

type WorkerAndHisTimer struct {
	lastHeartBeat time.Time
	workerId      string
}

type Task struct {
	content string
	id      string
}

const (
	_MY_HOST          string = "MY_HOST"
	_MY_PORT          string = "MY_PORT"
	_LOCK_SERVER_HOST string = "LOCK_SERVER_HOST"
	_LOCK_SERVER_PORT string = "LOCK_SERVER_PORT"
	_MQ_HOST          string = "MQ_HOST"
	_MQ_PORT          string = "MQ_PORT"
	_MQ_USERNAME      string = "MQ_USERNAME"
	_MQ_PASSWORD      string = "MQ_PASSWORD"
	_LOCAL_HOST       string = "127.0.0.1"
)

var MyHost string
var MyPort string
var LockServerHost string
var LockServerPort string
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
	MyHost = strings.Replace(utils.GetEnv(_MY_HOST, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	MyPort = utils.GetEnv(_MY_PORT, "5555")
	LockServerHost = strings.Replace(utils.GetEnv(_LOCK_SERVER_HOST, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	LockServerPort = utils.GetEnv(_LOCK_SERVER_PORT, "7777")
	MqHost = strings.Replace(utils.GetEnv(_MQ_HOST, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	MqPort = utils.GetEnv(_MQ_PORT, "5672")
	MqUsername = utils.GetEnv(_MQ_USERNAME, "guest")
	MqPassword = utils.GetEnv(_MQ_PASSWORD, "guest")
}
