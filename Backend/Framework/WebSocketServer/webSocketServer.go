package main

import (
	cache "Framework/Cache"
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	utils "Framework/Utils"
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

func newClient(id string, webSocketConn *websocket.Conn) *Client {
	return &Client{
		id:            id,
		finishedJobs:  make(chan string),
		webSocketConn: webSocketConn,
	}
}

func NewWebSocketServer() (*WebSocketServer, error) {

	webSocketServer := &WebSocketServer{
		id:        uuid.NewString(),
		cache:     cache.NewCache(cache.CreateCacheAddress(CacheHost, CachePort)),
		queue:     mq.NewMQ(mq.CreateMQAddress(MqUsername, MqPassword, MqHost, MqPort)),
		clients:   make(map[string]*Client),
		mu:        sync.Mutex{},
		writingMu: sync.Mutex{},
	}

	serveMux := mux.NewRouter()
	serveMux.HandleFunc("/openWS/{clientId}", webSocketServer.handleWebSocketConnections)
	serveMux.HandleFunc("/uploadBinary", webSocketServer.handleUploadBinaryRequests).Methods("POST")
	serveMux.HandleFunc("/deleteBinary", webSocketServer.handleDeleteBinaryRequests).Methods("POST")
	serveMux.HandleFunc("/getFinishedJobById", webSocketServer.handleGetFinishedJobByIdRequests).Methods("POST")
	serveMux.HandleFunc("/ping", webSocketServer.handlePingRequests).Methods("GET")

	webSocketServer.requestHandler = cors.AllowAll().Handler(middlewareLogger(serveMux))

	go webSocketServer.listenAndServe()
	go webSocketServer.deliverJobs()

	return webSocketServer, nil
}

func (webSocketServer *WebSocketServer) listenAndServe() {

	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Listening on %v:%v", MyHost, MyPort)

	if err := http.ListenAndServe(MyHost+":"+MyPort, webSocketServer.requestHandler); err != nil {
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Failed in listening on port} -> error : %v", err)
	}
}

// Middleware to log all incoming requests
func middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.LogRequest(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Request received from %v to %v", r.RemoteAddr, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// this method is responsible for listening on the specific websocket connection
// this method returns only when the connection is closed

func (webSocketServer *WebSocketServer) listenForJobs(client *Client) {

	defer func() {
		webSocketServer.mu.Lock()
		delete(webSocketServer.clients, client.id)
		webSocketServer.mu.Unlock()
	}()
	defer client.webSocketConn.Close()

	res := WsResponse{MsgType: JOB_REQUEST}

	for {
		_, message, err := client.webSocketConn.ReadMessage()

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Unable to read from the websocket %+v with err:\n%v", client.webSocketConn.RemoteAddr(), err)
			res.Response = utils.HttpResponse{Success: false, Response: ("Unable to read from the websocket")}
			webSocketServer.writeResp(client, res)
			return
		}

		newJobRequest := &JobRequest{}

		err = json.Unmarshal(message, newJobRequest)

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Error with client while unmarshaling json %v\n%v", client.webSocketConn.RemoteAddr(), err)
			res.Response = utils.HttpResponse{Success: false, Response: ("Invalid format")}
			webSocketServer.writeResp(client, res)
			continue
		}

		logger.LogInfo(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "This is the message received on the websocket connection %+v",
			struct {
				JobId              string
				JobContent         string
				DistributeBinaryId string
				ProcessBinaryId    string
				AggregateBinaryId  string
			}{JobId: newJobRequest.JobId, JobContent: newJobRequest.JobContent, DistributeBinaryId: newJobRequest.DistributeBinaryId, ProcessBinaryId: newJobRequest.ProcessBinaryId, AggregateBinaryId: newJobRequest.AggregateBinaryId},
		)

		resOfSendingOptionalFiles := webSocketServer.sendOptionalFilesToLockserver(client, newJobRequest)

		if !resOfSendingOptionalFiles.Response.Success {
			webSocketServer.writeResp(client, resOfSendingOptionalFiles)
			continue
		}

		modifiedJobRequest := &mq.AssignedJob{
			ClientId:  client.id,
			CreatedAt: time.Now(),
		}

		webSocketServer.modifyJobRequest(newJobRequest, modifiedJobRequest)
		logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "This is the modified job request %+v\nfor client %+v", modifiedJobRequest, client.webSocketConn.RemoteAddr())

		jobToAssign := new(bytes.Buffer)

		err = json.NewEncoder(jobToAssign).Encode(modifiedJobRequest)
		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Error with encoding data %+v for client %+v\n%+v", modifiedJobRequest, client.webSocketConn.RemoteAddr(), err)
			res.Response = utils.HttpResponse{Success: false, Response: ("Can't encode the job request and send it to the message queue at the moment")}
			webSocketServer.writeResp(client, res)
			webSocketServer.handleDeleteOptionalFiles(modifiedJobRequest.JobId)
			continue
		}

		err = webSocketServer.queue.Enqueue(mq.ASSIGNED_JOBS_QUEUE, jobToAssign.Bytes())

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{New job not enqeued to jobs assigned queue} -> error : %+v", err)
			res.Response = utils.HttpResponse{Success: false, Response: ("Message queue unavailable")}
			webSocketServer.writeResp(client, res)
			webSocketServer.handleDeleteOptionalFiles(modifiedJobRequest.JobId)
		} else {
			logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "New job successfully enqeued to jobs assigned queue")
		}

	}
}

func (webSocketServer *WebSocketServer) deliverJobs() {

	finishedJobsChan, err := webSocketServer.queue.Dequeue(mq.FINISHED_JOBS_QUEUE)

	if err != nil {
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Server can't consume finished Jobs} -> error : %v", err)
	}

	res := WsResponse{MsgType: FINISHED_JOB}

	for {

		time.Sleep(time.Second * 1)

		for finishedJobObj := range finishedJobsChan {

			finishedJob := &mq.FinishedJob{}
			err := json.Unmarshal(finishedJobObj.Body, finishedJob)
			if err != nil {
				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to unMarshal job } -> error : %v\n Will be discarded", err)
				finishedJobObj.Ack(false)
				continue
			}

			if finishedJob.ClientId == "" {
				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Corrupt job with no clientId %+v }. Will be discarded", finishedJob.JobId)
				finishedJobObj.Ack(false)
				continue
			}

			//CASES
			//cache is dead:
			//                          p1 client is alive     			   --send the response, and ack
			//							p2 client isn't alive  			   --nack
			//
			//cache is alive:
			//                          p1 client is alive	   	           --send the response, cache, and ack
			//							p2 client is dead		           --cache only, and ack

			webSocketServer.mu.Lock()
			client, clientIsAlive := webSocketServer.clients[finishedJob.ClientId]
			webSocketServer.mu.Unlock()

			var clientData *cache.CacheValue
			clientData, err = webSocketServer.cache.Get(finishedJob.ClientId)

			if err != nil && err != redis.Nil { //case 1 -- cache is dead

				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment} -> error : %v", err)

				if clientIsAlive { //p1

					if finishedJob.Err {
						res.Response = utils.HttpResponse{Success: false, Response: ("There was an Error while processing the job")}
					} else {
						res.Response = utils.HttpResponse{Success: true, Response: *finishedJob}
						logger.LogInfo(logger.WEBSOCKET_SERVER, logger.LOG_INFO, "About to send job to client")
					}

					webSocketServer.writeResp(client, res)

					finishedJobObj.Ack(false)

				} else { //p2

					logger.LogInfo(logger.WEBSOCKET_SERVER, logger.LOG_INFO, "Cache and Client both are dead job will be Nacked")
					finishedJobObj.Nack(false, false)
				}
			} else { //case 2 -- cache is alive

				finishedJobToCache := &cache.FinishedJob{
					JobId:        finishedJob.JobId,
					JobResult:    finishedJob.Result,
					CreatedAt:    finishedJob.CreatedAt,
					TimeAssigned: finishedJob.TimeAssigned,
				}

				if err == redis.Nil {

					clientData = &cache.CacheValue{
						ServerID:     webSocketServer.id, //no lock errors because no one writes the server.id
						FinishedJobs: []cache.FinishedJob{},
					}
				}

				//p1
				if clientIsAlive {

					if finishedJob.Err {
						res.Response = utils.HttpResponse{Success: false, Response: ("There was an Error while processing the job")}
					} else {
						res.Response = utils.HttpResponse{Success: true, Response: *finishedJob}
						logger.LogInfo(logger.WEBSOCKET_SERVER, logger.LOG_INFO, "job sent to client")
					}
					webSocketServer.writeResp(client, res)
				}

				//p2 & p1
				logger.LogInfo(logger.WEBSOCKET_SERVER, logger.LOG_INFO, "About to cache job")
				clientData.FinishedJobs = append(clientData.FinishedJobs, *finishedJobToCache)
				err := webSocketServer.cache.Set(finishedJob.ClientId, clientData, MAX_IDLE_CACHE_TIME)

				if err != nil {
					logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment, job will not be cached} -> error : %v", err)
				} else {
					logger.LogInfo(logger.WEBSOCKET_SERVER, logger.LOG_INFO, "Job cached successfully")
				}

				finishedJobObj.Ack(false)

			}

		}
	}
}

func (webSocketServer *WebSocketServer) sendSystemInfo(client *Client) {

	for {

		webSocketServer.mu.Lock()
		_, found := webSocketServer.clients[client.id]
		webSocketServer.mu.Unlock()

		if !found {
			return
		}

		go webSocketServer.writeResp(client, webSocketServer.GetFinishedJobsIds(client))

		go webSocketServer.writeResp(client, webSocketServer.GetSystemBinaries())

		time.Sleep(time.Second * 5)

	}

}

func (webSocketServer *WebSocketServer) sendSystemProgress(client *Client) {

	for {

		webSocketServer.mu.Lock()
		_, found := webSocketServer.clients[client.id]
		webSocketServer.mu.Unlock()

		if !found {
			return
		}

		go webSocketServer.writeResp(client, webSocketServer.GetSystemProgress())

		time.Sleep(time.Second * 1)

	}

}
