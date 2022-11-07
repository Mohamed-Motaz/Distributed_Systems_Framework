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

func newClient(webSocketConn *websocket.Conn) (*Client, error) {
	return &Client{
		id:              uuid.NewString(),
		finishedJobs:    make(chan string),
		webSocketConn:   webSocketConn,
		lastRequestTime: time.Now().Unix(),
	}, nil
}

func NewWebSocketServer() (*WebSocketServer, error) {

	webSocketServer := &WebSocketServer{
		cache:   cache.NewCache(cache.CreateCacheAddress(CacheHost, CachePort)),
		queue:   mq.NewMQ(mq.CreateMQAddress(MqUsername, MqPassword, MqHost, MqPort)),
		clients: make(map[string]*Client),
		mu:      sync.Mutex{},
	}

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", webSocketServer.handleRequests)
	webSocketServer.requestHandler = serveMux

	go webSocketServer.listenAndServe()
	go webSocketServer.deliverJobs()
	go webSocketServer.idleConnCloser()

	return webSocketServer, nil
}

func (webSocketServer *WebSocketServer) listenAndServe() {

	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Listening on %v:%v", MyHost, MyPort)

	if err := http.ListenAndServe(MyHost+":"+MyPort, webSocketServer.requestHandler); err != nil {
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Failed in listening on port} -> error : %v", err)
	}

}

func (webSocketServer *WebSocketServer) handleRequests(res http.ResponseWriter, req *http.Request) {

	upgradedConn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to upgrade http request to websocket} -> error : %v", err)
		return
	}

	client, err := newClient(upgradedConn)

	if err != nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to create client} -> error : %v", err)
		return
	}

	webSocketServer.mu.Lock()
	webSocketServer.clients[client.id] = client
	webSocketServer.mu.Unlock()

	go webSocketServer.assignJobs(client)
}

func (webSocketServer *WebSocketServer) writeFinishedJob(client *Client, finishedJob interface{}) {
	client.webSocketConn.WriteJSON(finishedJob)
}

func (webSocketServer *WebSocketServer) deliverJobs() {

	finishedJobsChan, err := webSocketServer.queue.Dequeue(mq.FINISHED_JOBS_QUEUE)

	if err != nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Server can't consume finished Jobs} -> error : %v", err)
		return
	}

	for {

		for finishedJobObj := range finishedJobsChan {

			finishedJob := &mq.FinishedJob{}
			err := json.Unmarshal(finishedJobObj.Body, finishedJob)
			if err != nil {
				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to unMarshal job %v } -> error : %v\n Will be discarded", string(finishedJobObj.Body), err)
				finishedJobObj.Ack(false)
				continue
			}

			webSocketServer.mu.Lock()

			if client, ok := webSocketServer.clients[finishedJob.ClientId]; ok {

				webSocketServer.mu.Unlock()

				go webSocketServer.writeFinishedJob(client, finishedJob)

				finishedJobObj.Ack(false)
				logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Job sent to client} %+v\n%+v", client.webSocketConn.RemoteAddr(), finishedJob)

			} else {
				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Connection with client may have been terminated}")
				finishedJobObj.Nack(false, true) //handle this case
			}

			webSocketServer.cache.Set(finishedJob.Content, finishedJob, MAX_IDLE_CACHE_TIME)
		}

		time.Sleep(time.Second * 5)
	}
}

func (webSocketServer *WebSocketServer) idleConnCloser() {

	for {

		for clientId, client := range webSocketServer.clients {

			if time.Now().Unix()-client.lastRequestTime > int64(time.Hour) {
				webSocketServer.mu.Lock()
				client.webSocketConn.Close()
				delete(webSocketServer.clients, clientId)
				webSocketServer.mu.Unlock()
			}
		}
		time.Sleep(time.Second * 5)
	}

}
