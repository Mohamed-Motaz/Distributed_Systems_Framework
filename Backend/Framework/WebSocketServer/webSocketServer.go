package main

import (
	cache "Framework/Cache"
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	"encoding/json"
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
	go webSocketServer.deliverJob();
	//go releaser
	
	return webSocketServer, nil;
}

func (webSocketServer *WebSocketServer) listenAndServe(){

	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Listening on %v:%v", MyHost, MyPort)

	if err := http.ListenAndServe(MyHost + ":" + MyPort, webSocketServer.requestHandler) ; err != nil{
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Failed in listening on port} -> error : %v", err)
	}

}

func (webSocketServer *WebSocketServer) handleRequest(res http.ResponseWriter, req *http.Request){

	upgradedConn, err := upgrader.Upgrade(res, req, nil);

	if err != nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to upgrade http request to websocket} -> error : %v", err);
		return;
	}

	client, err := newClient(upgradedConn);
	
	if err != nil {
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to create client} -> error : %v", err);
		return;
	}

	webSocketServer.mu.Lock();
	webSocketServer.clients[client.id] = client;
	webSocketServer.mu.Unlock();

	//go read
}

func (webSocketServer *WebSocketServer) deliverJob() {

	finishedJobsChan , err := webSocketServer.queue.Dequeue(mq.FINISHED_JOBS_QUEUE); 

	if err != nil{
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Server can't consume finished Jobs} -> error : %v", err)
		return;
	}

	for {

		for finishedJob := range finishedJobsChan{
			body := finishedJob.Body;
			data := &mq.FinishedJob{}
			err := json.Unmarshal(body, data)
			if err != nil {
				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to unMarshal job %v } -> error : %v\n Will be discarded", string(body),err) 
				finishedJob.Ack(false)
				continue
			}

			webSocketServer.mu.Lock();

			if client, ok := webSocketServer.clients[data.ClientId]; ok {

				webSocketServer.mu.Unlock();

				go func (client *Client,data interface{}){
					client.webSocketConn.WriteJSON(data);
				} (client, data)

				finishedJob.Ack(false)
				logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Job sent to client} %+v\n%+v", client.webSocketConn.RemoteAddr(), data) 

			}else{
				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Connection with client may have been terminated}") 
				finishedJob.Nack(false, true)  
			} 
			
		}
		time.Sleep(time.Second * 5);
	}

}

