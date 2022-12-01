package main

import (
	cache "Framework/Cache"
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	"Framework/RPC"
	"bytes"
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
	serveMux.HandleFunc("/submitJob", webSocketServer.handleJobRequests)
	serveMux.HandleFunc("/addExe", webSocketServer.handleAddExeRequests)
	serveMux.HandleFunc("/getAllExes", webSocketServer.handleGetAllExesRequests)
	serveMux.HandleFunc("/deleteExe", webSocketServer.handleDeleteExeRequests)

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

func (webSocketServer *WebSocketServer) handleJobRequests(res http.ResponseWriter, req *http.Request) {

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

func (webSocketServer *WebSocketServer) handleAddExeRequests(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Content-Type", "application/json")

	addExeRequestArgs := AddExeRequest{};

	reply := &RPC.FileUploadReply{};

	err := json.NewDecoder(req.Body).Decode(&addExeRequestArgs);

	if err != nil {
        http.Error(res, err.Error(), http.StatusBadRequest)
        return
    }

	ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
		Name:         "LockServer.HandleAddExeFile",
		Args:         addExeRequestArgs,
		Reply:        &reply,
		SenderLogger: logger.WEBSOCKET_SERVER,
		Reciever: RPC.Reciever{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if ok && !reply.Error.IsFound{
		json.NewEncoder(res).Encode(true);
		return;
	}

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with connect lockServer} -> error : %+v", err)
		
	} else if reply.Error.IsFound {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with recieving files from lockServer} -> error : %+v", reply.Error.Msg)
	}
	json.NewEncoder(res).Encode(false);

}

func (webSocketServer *WebSocketServer) handleGetAllExesRequests(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Content-Type", "application/json")

	
}
func (webSocketServer *WebSocketServer) handleDeleteExeRequests(res http.ResponseWriter, req *http.Request) {

	res.Header().Set("Content-Type", "application/json")

}

func (webSocketServer *WebSocketServer) writeFinishedJob(client *Client, finishedJob interface{}) {
	client.webSocketConn.WriteJSON(finishedJob)
}

func (webSocketServer *WebSocketServer) assignJobs(client *Client) {

	defer delete(webSocketServer.clients, client.id)
	defer client.webSocketConn.Close()

	for {

		_, message, err := client.webSocketConn.ReadMessage()
		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Error with client %v\n%v", client.webSocketConn.RemoteAddr(), err)
			return
		}

		//update client time
		webSocketServer.mu.Lock()
		client.lastRequestTime = time.Now().Unix() //lock this operation since cleaner is running
		webSocketServer.mu.Unlock()                //and may check on c.connTime

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Error with client %v\n%v", client.webSocketConn.RemoteAddr(), err)
			return
		}

		newJobRequest := &JobRequest{}

		err = json.Unmarshal(message, newJobRequest)

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Error with client %v\n%v", client.webSocketConn.RemoteAddr(), err)
			return
		}

		optionalFilesUploadArgs := &RPC.OptionalFilesUploadArgs{
			JobId: newJobRequest.JobId,
			Files: newJobRequest.OptionalFiles,
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
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with connect lockServer} -> error : %+v", err)
			return
		} else if reply.Error.IsFound {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with uploading files to lockServer} -> error : %+v", reply.Error.Msg)
			return
		}

		logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Optional Files sent to lockServer successfully")

		modifiedJobRequest := &mq.AssignedJob{}

		webSocketServer.modifyJobRequest(newJobRequest, modifiedJobRequest)

		cachedJob, err := webSocketServer.cache.Get(modifiedJobRequest.JobContent)

		if err == nil {
			go webSocketServer.writeFinishedJob(client, cachedJob)
			continue
		}

		jobToAssign := new(bytes.Buffer)

		err = json.NewEncoder(jobToAssign).Encode(modifiedJobRequest)

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "Error with client %+v\n%+v", client.webSocketConn.RemoteAddr(), err)
			return
		}
		//message is viable and isnt present in cache, can now send it over to mq
		err = webSocketServer.queue.Enqueue(mq.ASSIGNED_JOBS_QUEUE, jobToAssign.Bytes())

		if err != nil {
			logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{New job not Enqeue to jobs assigned queue} -> error : %+v", err)
		} else {
			logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "New job successfully Enqeue to jobs assigned queue")
		}

	}
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

func (webSocketServer *WebSocketServer) modifyJobRequest(jobRequest *JobRequest, modifiedJobRequest *mq.AssignedJob) {

	modifiedJobRequest.ClientId = jobRequest.ClientId
	modifiedJobRequest.JobId = jobRequest.JobId
	modifiedJobRequest.JobContent = jobRequest.JobContent
	for _, optionalFile := range jobRequest.OptionalFiles {
		modifiedJobRequest.OptionalfilesNames = append(modifiedJobRequest.OptionalfilesNames, optionalFile.Name)
	}
	modifiedJobRequest.DistributeExeName = jobRequest.DistributeExeName
	modifiedJobRequest.ProcessExeName = jobRequest.ProcessExeName
	modifiedJobRequest.AggregateExeName = jobRequest.AggregateExeName
}
