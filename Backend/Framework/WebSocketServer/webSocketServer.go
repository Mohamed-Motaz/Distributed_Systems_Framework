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
	serveMux.HandleFunc("/openWS/{clientId}", webSocketServer.handleJobRequests)
	serveMux.HandleFunc("/uploadBinary", webSocketServer.handleUploadBinaryRequests).Methods("POST")
	serveMux.HandleFunc("/getAllBinaries", webSocketServer.handleGetAllBinariesRequests).Methods("POST")
	serveMux.HandleFunc("/deleteBinary", webSocketServer.handleDeleteBinaryRequests).Methods("POST")
	serveMux.HandleFunc("/getJobProgress", webSocketServer.handleGetSystemProgressRequests).Methods("POST")
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

func (webSocketServer *WebSocketServer) writeFinishedJob(client *Client, finishedJob mq.FinishedJob) {
	client.webSocketConn.WriteJSON(finishedJob)
	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Job sent to client} %+v\n%+v", client.webSocketConn.RemoteAddr(), finishedJob)
}

func (webSocketServer *WebSocketServer) writeError(client *Client, err utils.Error) {
	client.webSocketConn.WriteJSON(err)
	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error sent to client} %+v\n%+v", client.webSocketConn.RemoteAddr(), err)
}

// this method shows whether or not the server succeeding in sending the optionalFilesZip if there are any
// this method is responsible for sending the errors it encounters to the client
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

	if !ok { //can't establish connection to the lockserver
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with connecting lockServer} -> error : %+v", err)
		websocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Error with connecting lockServer"}) //send a message eshtem to client
		return false
	} else if reply.Err { //establish a connection to the lockserver, but the operation fails
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with uploading files to lockServer} -> error : %+v", reply.ErrMsg)
		websocketServer.writeError(client, utils.Error{Err: true, ErrMsg: fmt.Sprintf("Error with uploading files to lockServer: %+v", reply.Err)})
		return false
	} else {
		logger.LogInfo(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Optional Files sent to lockServer successfully")
	}
	return true
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

	for {

		_, message, err := client.webSocketConn.ReadMessage()
		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Unable to read from the websocket %+v with err:\n%v", client.webSocketConn.RemoteAddr(), err)
			webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Unable to read from the websocket"})
			return
		}

		newJobRequest := &JobRequest{}

		err = json.Unmarshal(message, newJobRequest)

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Error with client while unmarshaling json %v\n%v", client.webSocketConn.RemoteAddr(), err)
			webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Invalid format"})
			continue
		}

		//immediately attempt to send the optional files to the lockserver
		if !webSocketServer.sendOptionalFiles(client, newJobRequest) { //no need to send the errors since this method is responsible for this
			continue
		}

		modifiedJobRequest := &mq.AssignedJob{}

		webSocketServer.modifyJobRequest(newJobRequest, modifiedJobRequest)

		jobToAssign := new(bytes.Buffer)
		err = json.NewEncoder(jobToAssign).Encode(modifiedJobRequest)
		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Error with encoding data %+v for client %+v\n%+v", modifiedJobRequest, client.webSocketConn.RemoteAddr(), err)
			webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Can't encode the job request and send it to the message queue at the moment"})
			//DONE, send an rpc to the lockserver telling it to delete the files
			webSocketServer.handleDeleteOptionalFiles(modifiedJobRequest.JobId);
			continue
		}

		err = webSocketServer.queue.Enqueue(mq.ASSIGNED_JOBS_QUEUE, jobToAssign.Bytes())

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{New job not enqeued to jobs assigned queue} -> error : %+v", err)
			webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Message queue unavailable"})
			//DONE, send an rpc to the lockserver telling it to delete the files
			webSocketServer.handleDeleteOptionalFiles(modifiedJobRequest.JobId);
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

	for {
		time.Sleep(time.Second * 5)

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

			webSocketServer.mu.Lock()
			client, clientIsAlive := webSocketServer.clients[finishedJob.ClientId]
			webSocketServer.mu.Unlock()

			//DONE: send the appropriate response to the user

			if clientIsAlive {
				go webSocketServer.writeFinishedJob(client, *finishedJob)
				//cache is not working at the moment
				if err != nil && err != redis.Nil {
					logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment} -> error : %v", err)
					finishedJobObj.Ack(false)
					continue
				}
			}

			if clientData.ServerID == webSocketServer.id {

				clientData.FinishedJobsResults = append(clientData.FinishedJobsResults, finishedJob.Content)
				err := webSocketServer.cache.Set(finishedJob.ClientId, clientData, MAX_IDLE_CACHE_TIME)

				if err != nil {
					logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment} -> error : %v", err)
				}
				finishedJobObj.Ack(false)

			} else {
				finishedJobObj.Nack(false, true)
			}

		}

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
