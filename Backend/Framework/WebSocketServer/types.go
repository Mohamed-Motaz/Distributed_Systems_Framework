package main

import (
	cache "Framework/Cache"
	mq "Framework/MessageQueue"
	utils "Framework/Utils"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type Client struct {
	id              string
	finishedJobs    chan string
	webSocketConn   *websocket.Conn
	lastRequestTime int64
}

type WebSocketServer struct {
	requestHandler http.Handler
	cache          *cache.Cache
	queue          *mq.MQ
	clients        map[string]*Client
	mu             sync.Mutex
}

const (
	_MY_HOST            string        = "MY_HOST"
	_MY_PORT            string        = "MY_PORT"
	_CACHE_HOST         string        = "CACHE_HOST"
	_CACHE_PORT         string        = "CACHE_PORT"
	_MQ_HOST            string        = "MQ_HOST"
	_MQ_PORT            string        = "MQ_PORT"
	_MQ_USERNAME        string        = "MQ_USERNAME"
	_MQ_PASSWORD        string        = "MQ_PASSWORD"
	_LOCAL_HOST         string        = "127.0.0.1"
	MAX_IDLE_CACHE_TIME time.Duration = time.Hour * 24 * 30
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024 * 1024,
		CheckOrigin:     func(req *http.Request) bool { return true },
	}
	MyHost     string
	MyPort     string
	CacheHost  string
	CachePort  string
	MqHost     string
	MqPort     string
	MqUsername string
	MqPassword string
)

func init() {
	if !utils.IN_DOCKER {
		err := godotenv.Load("config.env")
		if err != nil {
			log.Fatal("Error loading config.env file")
		}
	}
	MyHost = strings.Replace(utils.GetEnv(_MY_HOST, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	MyPort = utils.GetEnv(_MY_PORT, "6666")
	CacheHost = strings.Replace(utils.GetEnv(_CACHE_HOST, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	CachePort = utils.GetEnv(_CACHE_PORT, "6379")
	MqHost = strings.Replace(utils.GetEnv(_MQ_HOST, _LOCAL_HOST), "_", ".", -1) //replace all "_" with ".""
	MqPort = utils.GetEnv(_MQ_PORT, "5672")
	MqUsername = utils.GetEnv(_MQ_USERNAME, "guest")
	MqPassword = utils.GetEnv(_MQ_PASSWORD, "guest")
}
