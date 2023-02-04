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
	"github.com/gorilla/websocket"
)

func newClient(webSocketConn *websocket.Conn) *Client {
	return &Client{
		id:            uuid.NewString(),
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

	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/submitJob", webSocketServer.handleJobRequests)
	serveMux.HandleFunc("/uploadBinary", webSocketServer.handleUploadBinaryRequests)
	serveMux.HandleFunc("/getAllBinaries", webSocketServer.handleGetAllBinariesRequests)
	serveMux.HandleFunc("/deleteBinary", webSocketServer.handleDeleteBinaryRequests)
	serveMux.HandleFunc("/getJobProgress", webSocketServer.handleGetJobProgressRequests)
	serveMux.HandleFunc("/getAllFinishedJobs", webSocketServer.handleGetAllFinishedJobsRequests)

	webSocketServer.requestHandler = serveMux

	go webSocketServer.listenAndServe()
	go webSocketServer.deliverJobs()

	return webSocketServer, nil
}

//todo ADD MIDDLEWARE TO LOG ALL REQUESTS

func (webSocketServer *WebSocketServer) listenAndServe() {

	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Listening on %v:%v", MyHost, MyPort)

	if err := http.ListenAndServe(MyHost+":"+MyPort, webSocketServer.requestHandler); err != nil {
		logger.FailOnError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Failed in listening on port} -> error : %v", err)
	}
}

func (webSocketServer *WebSocketServer) handleJobRequests(res http.ResponseWriter, req *http.Request) {

	upgradedConn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to upgrade http request to websocket} -> error : %v", err)
		return
	}

	client := newClient(upgradedConn)

	clientData := &cache.CacheValue{}
	clientData, err = webSocketServer.cache.Get(client.id)

	webSocketServer.mu.Lock()
	if err != nil && err != redis.Nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment} -> error : %v", err)
		webSocketServer.mu.Unlock()
		client.webSocketConn.Close()
		return
	}
	clientData.ServerID = webSocketServer.id
	webSocketServer.cache.Set(client.id, clientData, MAX_IDLE_CACHE_TIME)
	webSocketServer.clients[client.id] = client
	webSocketServer.mu.Unlock()

	go webSocketServer.listenForJobs(client)
}

func (webSocketServer *WebSocketServer) handleUploadBinaryRequests(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Content-Type", "application/json")

	uploadBinaryRequest := UploadBinaryRequest{}

	err := json.NewDecoder(req.Body).Decode(&uploadBinaryRequest)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	//map the dto to rpc args
	uploadBinaryRequestArgs := RPC.BinaryUploadArgs{
		FileType: utils.FileType(uploadBinaryRequest.FileType),
		File: utils.RunnableFile{
			RunCmd: uploadBinaryRequest.RunCmd,
			File: utils.File{
				Name:    uploadBinaryRequest.Name,
				Content: uploadBinaryRequest.Content,
			},
		},
	}

	reply := &RPC.FileUploadReply{}

	ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
		Name:         "LockServer.HandleAddBinaryFile",
		Args:         uploadBinaryRequestArgs,
		Reply:        &reply,
		SenderLogger: logger.WEBSOCKET_SERVER,
		Reciever: RPC.Reciever{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if ok && !reply.Err {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(true)
		return
	}

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with connect lockServer} -> error : %+v", err)

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with Adding files to lockServer} -> error : %+v", reply.ErrMsg)
	}
	res.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(res).Encode(false)
}

func (webSocketServer *WebSocketServer) handleGetAllBinariesRequests(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Content-Type", "application/json")

	reply := &RPC.GetBinaryFilesReply{}

	ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
		Name:         "LockServer.HandleGetBinaryFiles",
		Args:         nil,
		Reply:        &reply,
		SenderLogger: logger.WEBSOCKET_SERVER,
		Reciever: RPC.Reciever{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if ok && !reply.Err {
		json.NewEncoder(res).Encode(reply)
		return
	}

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with connect lockServer} -> error : %+v", err)

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with recieving files from lockServer} -> error : %+v", reply.ErrMsg)
	}
	json.NewEncoder(res).Encode(false)
}

func (webSocketServer *WebSocketServer) handleDeleteBinaryRequests(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Content-Type", "application/json")

	deleteBinaryRequest := DeleteBinaryRequest{}

	reply := &RPC.DeleteBinaryFileReply{}

	err := json.NewDecoder(req.Body).Decode(&deleteBinaryRequest)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	deleteBinaryRequestArgs := RPC.DeleteBinaryFileArgs{
		FileType: utils.FileType(deleteBinaryRequest.FileType),
		FileName: deleteBinaryRequest.FileName,
	}

	ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
		Name:         "LockServer.HandleDeleteBinaryFile",
		Args:         deleteBinaryRequestArgs,
		Reply:        &reply,
		SenderLogger: logger.WEBSOCKET_SERVER,
		Reciever: RPC.Reciever{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if ok && !reply.Err {
		json.NewEncoder(res).Encode(true)
		return
	}

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with connect lockServer} -> error : %+v", err)

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with Deleting Binary file from lockServer} -> error : %+v", reply.ErrMsg)
	}
	json.NewEncoder(res).Encode(false)
}

func (webSocketServer *WebSocketServer) handleGetJobProgressRequests(res http.ResponseWriter, req *http.Request) {

	GetJobRequest := GetJobProgressRequest{}

	err := json.NewDecoder(req.Body).Decode(&GetJobRequest)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	GetjobRequestArgs := RPC.GetJobProgressArgs{
		JobId: GetJobRequest.JobId,
	}

	reply := &RPC.GetJobProgressReply{}

	ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
		Name:         "LockServer.HandleGetJobProgress",
		Args:         GetjobRequestArgs,
		Reply:        &reply,
		SenderLogger: logger.WEBSOCKET_SERVER,
		Reciever: RPC.Reciever{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if ok {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(reply)
		return
	} else {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with connecting lockServer} -> error : %+v", err)
	}
	res.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(res).Encode(false)
}

func (webSocketServer *WebSocketServer) handleGetAllFinishedJobsRequests(res http.ResponseWriter, req *http.Request) {

	GetAllFinishedJobsRequest := GetAllFinishedJobsRequest{}

	err := json.NewDecoder(req.Body).Decode(&GetAllFinishedJobsRequest)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	finishedJobs := &cache.CacheValue{}
	finishedJobs, err = webSocketServer.cache.Get(GetAllFinishedJobsRequest.ClientId)

	if err == nil {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(finishedJobs)

	} else if err == redis.Nil {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode("No jobs Found")

	} else {
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(false)
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
		websocketServer.writeFinishedJob(client, utils.Error{Err: true, ErrMsg: fmt.Sprintf("Error with connecting lockServer")}) //send a message eshtem to client
		return false
	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with uploading files to lockServer} -> error : %+v", reply.ErrMsg)
		websocketServer.writeFinishedJob(client, utils.Error{Err: true, ErrMsg: fmt.Sprintf("Error with uploading files to lockServer")})
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
			webSocketServer.writeFinishedJob(client, utils.Error{Err: true, ErrMsg: fmt.Sprintf("Message queue is not available")})
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
				finishedJobObj.Nack(false, true)
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
