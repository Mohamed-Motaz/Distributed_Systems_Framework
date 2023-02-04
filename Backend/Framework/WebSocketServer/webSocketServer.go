package main

import (
	cache "Framework/Cache"
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	"Framework/RPC"
	utils "Framework/Utils"
	"bytes"
	"encoding/json"
	"fmt"
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
		id:      uuid.NewString(),
		cache:   cache.NewCache(cache.CreateCacheAddress(CacheHost, CachePort)),
		queue:   mq.NewMQ(mq.CreateMQAddress(MqUsername, MqPassword, MqHost, MqPort)),
		clients: make(map[string]*Client),
		mu:      sync.Mutex{},
	}

	serveMux := mux.NewRouter()
	serveMux.HandleFunc("/openWS", webSocketServer.handleJobRequests)
	serveMux.HandleFunc("/uploadBinary", webSocketServer.handleUploadBinaryRequests).Methods("POST")
	serveMux.HandleFunc("/getAllBinaries", webSocketServer.handleGetAllBinariesRequests).Methods("POST")
	serveMux.HandleFunc("/deleteBinary", webSocketServer.handleDeleteBinaryRequests).Methods("POST")
	serveMux.HandleFunc("/getJobProgress", webSocketServer.handleGetJobProgressRequests).Methods("POST")
	serveMux.HandleFunc("/getAllFinishedJobs", webSocketServer.handleGetAllFinishedJobsRequests).Methods("POST")

	webSocketServer.requestHandler = cors.AllowAll().Handler(middlewareLogger(serveMux))

	go webSocketServer.listenAndServe()
	go webSocketServer.deliverJobs()

	return webSocketServer, nil
}

// Middleware to log all incoming requests
func middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.LogRequest(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Request received from %v to %v", r.RemoteAddr, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func (webSocketServer *WebSocketServer) listenAndServe() {

	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Listening on %v:%v", MyHost, MyPort)

	if err := http.ListenAndServe(MyHost+":"+MyPort, webSocketServer.requestHandler); err != nil {
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Failed in listening on port} -> error : %v", err)
	}
}

func (webSocketServer *WebSocketServer) writeFinishedJob(client *Client, finishedJob interface{}) {
	client.webSocketConn.WriteJSON(finishedJob)
}

func (websocketServer *WebSocketServer) sendOptionalFiles(client *Client, newJobRequest *JobRequest) bool {

	if len(newJobRequest.OptionalFilesZip.Content) == 0 {
		return true
	}

	optionalFilesUploadArgs := &RPC.OptionalFilesUploadArgs{
		JobId:    newJobRequest.JobId,
		FilesZip: newJobRequest.OptionalFilesZip,
	}

	reply := &RPC.FileUploadReply{}

	ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
		Name:         "LockServer.HandleAddOptionalFiles",
		Args:         optionalFilesUploadArgs,
		Reply:        &reply,
		SenderLogger: logger.WEBSOCKET_SERVER,
		Reciever: RPC.Reciever{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with connecting lockServer} -> error : %+v", err)
		websocketServer.writeFinishedJob(client, utils.Error{Err: true, ErrMsg: "Error with connecting lockServer"}) //send a message eshtem to client
		return false
	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with uploading files to lockServer} -> error : %+v", reply.ErrMsg)
		websocketServer.writeFinishedJob(client, utils.Error{Err: true, ErrMsg: "Error with uploading files to lockServer"})
		return false
	} else {
		logger.LogInfo(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Optional Files sent to lockServer successfully")
	}
	return true
}

func (webSocketServer *WebSocketServer) listenForJobs(client *Client) {

	defer func() {
		webSocketServer.mu.Lock()
		delete(webSocketServer.clients, client.id)
		webSocketServer.mu.Unlock()
	}()
	defer client.webSocketConn.Close()

	for {

		_, message, err := client.webSocketConn.ReadMessage()
		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Error with client while reading the  %v\n%v", client.webSocketConn.RemoteAddr(), err)
			webSocketServer.writeFinishedJob(client, utils.Error{Err: true, ErrMsg: fmt.Sprintf("Unable to read from the websocket with err: %+v", err)})
			return
		}

		newJobRequest := &JobRequest{}

		err = json.Unmarshal(message, newJobRequest)

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Error with client while unmarshaling json %v\n%v", client.webSocketConn.RemoteAddr(), err)
			webSocketServer.writeFinishedJob(client, utils.Error{Err: true, ErrMsg: fmt.Sprintf("Invalid format with err: %+v", err)})
			continue
		}

		//immediately attempt to send the optional files to the lockserver

		if !webSocketServer.sendOptionalFiles(client, newJobRequest) {
			continue
		}

		modifiedJobRequest := &mq.AssignedJob{}

		webSocketServer.modifyJobRequest(newJobRequest, modifiedJobRequest)

		jobToAssign := new(bytes.Buffer)

		err = json.NewEncoder(jobToAssign).Encode(modifiedJobRequest)

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Error with client %+v\n%+v", client.webSocketConn.RemoteAddr(), err)
			//server error, can't encode the job request at the moment
			continue
		}

		err = webSocketServer.queue.Enqueue(mq.ASSIGNED_JOBS_QUEUE, jobToAssign.Bytes())

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{New job not Enqeue to jobs assigned queue} -> error : %+v", err)
			webSocketServer.writeFinishedJob(client, utils.Error{Err: true, ErrMsg: "Message queue is not available"})
			//send to user telling him that mq is not available now
		} else {
			logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "New job successfully Enqeue to jobs assigned queue")
		}

	}
}

func (webSocketServer *WebSocketServer) deliverJobs() {

	finishedJobsChan, err := webSocketServer.queue.Dequeue(mq.FINISHED_JOBS_QUEUE)

	if err != nil {
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Server can't consume finished Jobs} -> error : %v", err)
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

			clientData := &cache.CacheValue{}
			clientData, err = webSocketServer.cache.Get(finishedJob.ClientId)

			if err != nil && err != redis.Nil {
				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment} -> error : %v", err)
				finishedJobObj.Ack(false) //todo: send the appropriate response to the user
				continue
			}

			if clientData.ServerID == webSocketServer.id {

				clientData.FinishedJobsResults = append(clientData.FinishedJobsResults, finishedJob.Content)
				webSocketServer.cache.Set(finishedJob.ClientId, clientData, MAX_IDLE_CACHE_TIME)

				webSocketServer.mu.Lock()
				client, ok := webSocketServer.clients[finishedJob.ClientId]
				webSocketServer.mu.Unlock()

				if ok {
					go webSocketServer.writeFinishedJob(client, finishedJob)
					logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Job sent to client} %+v\n%+v", client.webSocketConn.RemoteAddr(), finishedJob)
				} else {
					logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Connection with client may have been terminated}")
				}

				finishedJobObj.Ack(false)

			} else {
				finishedJobObj.Nack(false, true)
				logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Client is currently connecting to another server}")
			}

		}

		time.Sleep(time.Second * 5)
	}
}

func (webSocketServer *WebSocketServer) modifyJobRequest(jobRequest *JobRequest, modifiedJobRequest *mq.AssignedJob) {

	modifiedJobRequest.ClientId = jobRequest.ClientId
	modifiedJobRequest.JobId = jobRequest.JobId
	modifiedJobRequest.JobContent = jobRequest.JobContent
	modifiedJobRequest.DistributeBinaryName = jobRequest.DistributeBinaryName
	modifiedJobRequest.ProcessBinaryName = jobRequest.ProcessBinaryName
	modifiedJobRequest.AggregateBinaryName = jobRequest.AggregateBinaryName
}
