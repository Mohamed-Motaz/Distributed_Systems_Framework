package main

import (
	cache "Framework/Cache"
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func newClient (webSocketConn *websocket.Conn) (*Client, error){
	return &Client {
		id: uuid.NewString(),
		finishedJobs: make(chan string),
		webSocketConn: webSocketConn,
		connStartTime: time.Now().Unix(),
	},nil;
}

func newWebSocketServer() (*WebSocketServer,error){

	webSocketServer := &WebSocketServer{
		cache : cache.NewCache(cache.CreateCacheAddress(CacheHost,CachePort)),
		queue : mq.NewMQ(mq.CreateMQAddress(MqUsername, MqPassword, MqHost, MqPort)),
		clients : make(map[string]*Client),
		mu: sync.Mutex{},
	}

	serveMux := http.NewServeMux();
	serveMux.HandleFunc("/",webSocketServer.handleRequest);
	webSocketServer.requestHandler = serveMux;
	
	go webSocketServer.listenAndServe();
	//go write
	//go releaser
	
	return webSocketServer, nil;
}

func (webSocketServer *WebSocketServer) listenAndServe(){

	logger.LogInfo(logger.SERVER, logger.ESSENTIAL, "Listening on %v:%v", MyHost, MyPort)

	if err := http.ListenAndServe(MyHost + ":" + MyPort, webSocketServer.requestHandler) ; err != nil{
		logger.FailOnError(logger.SERVER, logger.ESSENTIAL, "{Failed in listening on port} -> error : %v", err)
	}

}

func (webSocketServer *WebSocketServer) handleRequest(res http.ResponseWriter, req *http.Request){

	upgradedConn, err := upgrader.Upgrade(res, req, nil);

	if err != nil {
		logger.LogError(logger.SERVER, logger.ESSENTIAL, "{Unable to upgrade http request to websocket} -> error : %v", err);
		return;
	}

	client, err := newClient(upgradedConn);
	
	if err != nil {
		logger.FailOnError(logger.SERVER, logger.ESSENTIAL, "{Unable to create client} -> error : %v", err);
		return;
	}

	webSocketServer.mu.Lock();
	webSocketServer.clients[client.id] = client;
	webSocketServer.mu.Unlock();

	//go read

}

