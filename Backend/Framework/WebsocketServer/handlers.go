package main

import (
	cache "Framework/Cache"
	logger "Framework/Logger"
	"Framework/RPC"
	utils "Framework/Utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// deals with a ws upgrade connection
func (webSocketServer *WebSocketServer) handleJobRequests(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)
	clientId, ok := vars["clientId"]
	if !ok {
		//DONE respond with an error requiring an id
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Client didn't send the ID}")
		res.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(res).Encode(utils.Error{Err: true, ErrMsg: ("You must send the client Id")})
		return
	}

	upgradedConn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to upgrade http request to websocket} -> error : %v", err)
		return
	}

	client := newClient(clientId, upgradedConn)

	clientData, err := webSocketServer.cache.Get(client.id)

	webSocketServer.mu.Lock()
	if err != nil && err != redis.Nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment} -> error : %v", err)
		webSocketServer.mu.Unlock()
		webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Cache is down temporarily, please try again later"})
		client.webSocketConn.Close() //need to close the connection because the cache is down, and I can't map the client to the server
		return
	}

	webSocketServer.clients[client.id] = client
	webSocketServer.mu.Unlock()

	if clientData != nil {
		clientData.ServerID = webSocketServer.id
		//leave the finishedJobsResults as it is
	} else {
		clientData = &cache.CacheValue{
			ServerID:     webSocketServer.id,
			FinishedJobs: make([]cache.FinishedJob, 0),
		}
	}

	err = webSocketServer.cache.Set(client.id, clientData, MAX_IDLE_CACHE_TIME)

	if err != nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment} -> error : %v", err)
		webSocketServer.writeError(client, utils.Error{Err: true, ErrMsg: "Cache is down temporarily, please try again later"})
		client.webSocketConn.Close() //need to close the connection because the cache is down, and I can't map the client to the server
		return
	}

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
		FileType: uploadBinaryRequest.FileType,
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

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error while connecting lockServer} -> error : %+v", err)
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(utils.Error{Err: true, ErrMsg: "Unable to connect to lockserver"})

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with Adding files to lockServer} -> error : %+v", reply.ErrMsg)
		res.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(res).Encode(utils.Error{Err: true, ErrMsg: fmt.Sprintf("Error while adding files to the lockserver %+v", reply.ErrMsg)})
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(utils.Success{Success: true})
	}
}

// DONE fix the rest
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

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error while connecting lockServer} -> error : %+v", err)
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(utils.Error{Err: true, ErrMsg: "Unable to connect to lockserver"})

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with recieving files from lockServer} -> error : %+v", reply.ErrMsg)
		res.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(res).Encode(utils.Error{Err: true, ErrMsg: fmt.Sprintf("Error with recieving files from lockServer %+v", reply.ErrMsg)})
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(utils.Success{Success: true, Response: reply})
	}
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
		FileType: deleteBinaryRequest.FileType,
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

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error while connecting lockServer} -> error : %+v", err)
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(utils.Error{Err: true, ErrMsg: "Unable to connect to lockserver"})

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with Deleting Binary file from lockServer} -> error : %+v", reply.ErrMsg)
		res.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(res).Encode(utils.Error{Err: true, ErrMsg: fmt.Sprintf("Error with Deleting Binary file from lockServer %+v", reply.ErrMsg)})
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(utils.Success{Success: true})
	}
}

func (webSocketServer *WebSocketServer) handleGetSystemProgressRequests(res http.ResponseWriter, req *http.Request) {

	GetSystemProgressRequest := GetSystemProgressRequest{}

	err := json.NewDecoder(req.Body).Decode(&GetSystemProgressRequest)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	reply := &RPC.GetSystemProgressReply{}

	ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
		Name:         "LockServer.HandleGetSystemProgress",
		Args:         &RPC.GetSystemProgressArgs{},
		Reply:        &reply,
		SenderLogger: logger.WEBSOCKET_SERVER,
		Reciever: RPC.Reciever{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error while connecting lockServer} -> error : %+v", err)
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(utils.Error{Err: true, ErrMsg: "Unable to connect to lockserver"})

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with Getting system progress from lockServer} -> error : %+v", reply.ErrMsg)
		res.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(res).Encode(utils.Error{Err: true, ErrMsg: fmt.Sprintf("Error with Getting system progress from lockServer %+v", reply.ErrMsg)})
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(utils.Success{Success: true, Response: reply})
	}
}

func (webSocketServer *WebSocketServer) handleGetAllFinishedJobsRequests(res http.ResponseWriter, req *http.Request) {

	GetAllFinishedJobsRequest := GetAllFinishedJobsRequest{}

	err := json.NewDecoder(req.Body).Decode(&GetAllFinishedJobsRequest)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var finishedJobs *cache.CacheValue
	finishedJobs, err = webSocketServer.cache.Get(GetAllFinishedJobsRequest.ClientId)

	if err == nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "{Finished jobs sent to client} -> jobs : %+v", finishedJobs)
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(utils.Success{Success: true, Response: finishedJobs})

	} else if err == redis.Nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "No jobs found")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(utils.Success{Success: true, Response: "No jobs Found"})
	} else {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error while connecting to cache at the moment} -> error : %+v", err)
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(utils.Error{Err: true, ErrMsg: "Error while connecting to cache at the moment"})
	}
}

func (websocketServer *WebSocketServer) handleSendOptionalFiles(client *Client, newJobRequest *JobRequest) bool {

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
		websocketServer.writeError(client, utils.Error{Err: true, ErrMsg: fmt.Sprintf("Error with uploading files to lockServer: %+v", reply.ErrMsg)})
		return false
	} else {
		logger.LogInfo(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Optional Files sent to lockServer successfully")
	}
	return true
}

func (webSocketServer *WebSocketServer) handleDeleteOptionalFiles(jobId string) {

	reply := RPC.DeleteOptionalFilesReply{}

	deleteOptionalFilesArgs := RPC.DeleteOptionalFilesArgs{
		JobId: jobId,
	}

	ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
		Name:         "LockServer.HandleDeleteOptionalFiles",
		Args:         deleteOptionalFilesArgs,
		Reply:        &reply,
		SenderLogger: logger.WEBSOCKET_SERVER,
		Reciever: RPC.Reciever{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error while connecting lockServer} -> error : %+v", err)
	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with Deleting Optional file from lockServer} -> error : %+v", reply.ErrMsg)
	} else {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "{Optional Files Deleted}")
	}

}
