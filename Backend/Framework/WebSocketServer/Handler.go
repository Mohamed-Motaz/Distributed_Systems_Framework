package main

import (
	cache "Framework/Cache"
	logger "Framework/Logger"
	"Framework/RPC"
	utils "Framework/Utils"
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis/v8"
)

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
		Args:         &RPC.GetBinaryFilesArgs{},
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
