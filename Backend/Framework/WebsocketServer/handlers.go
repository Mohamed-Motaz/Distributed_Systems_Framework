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
func (webSocketServer *WebSocketServer) handleWebSocketConnections(res http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)
	clientId, ok := vars["clientId"]
	if !ok {
		//DONE respond with an error requiring an id
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Client didn't send the ID}")
		res.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(res).Encode(utils.HttpResponse{Success: false, Response: ("You must send the client Id")})
		return
	}

	upgradedConn, err := connectionUpgrader.Upgrade(res, req, nil)

	if err != nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to upgrade http request to websocket} -> error : %v", err)
		return
	}

	client := newClient(clientId, upgradedConn)

	clientData, err := webSocketServer.cache.Get(client.id)

	if err != nil && err != redis.Nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Unable to connect to cache at the moment} -> error : %v", err)
		webSocketServer.writeResp(client, WsResponse{JOB_REQUEST, utils.HttpResponse{Success: false, Response: ("Cache is down temporarily, please try again later")}})
		client.webSocketConn.Close() //need to close the connection because the cache is down, and I can't map the client to the server
		return
	}

	webSocketServer.mu.Lock()
	webSocketServer.clients[client.id] = client
	webSocketServer.mu.Unlock()

	if err != redis.Nil {
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
		webSocketServer.writeResp(client, WsResponse{JOB_REQUEST, utils.HttpResponse{Success: false, Response: ("Cache is down temporarily, please try again later")}})
		client.webSocketConn.Close() //need to close the connection because the cache is down, and I can't map the client to the server
		return
	}

	go webSocketServer.listenForJobs(client)
	go webSocketServer.sendSystemInfo(client)
	go webSocketServer.sendSystemProgress(client)
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
		Receiver: RPC.Receiver{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error while connecting lockServer} -> error : %+v", err)
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(utils.HttpResponse{Success: false, Response: "Unable to connect to lockserver"})

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with Adding files to lockServer} -> error : %+v", reply.ErrMsg)
		res.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(res).Encode(utils.HttpResponse{Success: false, Response: fmt.Sprintf("Error while adding files to the lockserver %+v", reply.ErrMsg)})
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(utils.HttpResponse{Success: true})
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
		Receiver: RPC.Receiver{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if !ok {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error while connecting lockServer} -> error : %+v", err)
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(utils.HttpResponse{Success: false, Response: "Unable to connect to lockserver"})

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with Deleting Binary file from lockServer} -> error : %+v", reply.ErrMsg)
		res.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(res).Encode(utils.HttpResponse{Success: false, Response: fmt.Sprintf("Error with Deleting Binary file from lockServer %+v", reply.ErrMsg)})
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(utils.HttpResponse{Success: true})
	}
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
		Receiver: RPC.Receiver{
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

func (webSocketServer *WebSocketServer) handleGetFinishedJobByIdRequests(res http.ResponseWriter, req *http.Request) {

	GetFinishedJobByIdRequest := GetFinishedJobByIdRequest{}

	err := json.NewDecoder(req.Body).Decode(&GetFinishedJobByIdRequest)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	var finishedJobs *cache.CacheValue
	finishedJobs, err = webSocketServer.cache.Get(GetFinishedJobByIdRequest.ClientId)

	if err == nil {

		for _, finishedJob := range finishedJobs.FinishedJobs {
			if finishedJob.JobId == GetFinishedJobByIdRequest.JobId {
				logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "{Finished job sent to client} -> job : %+v", finishedJob)
				res.WriteHeader(http.StatusOK)
				json.NewEncoder(res).Encode(utils.HttpResponse{Success: true, Response: finishedJob.JobResult})
				return
			}
		}

		logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Job ID not found")
		res.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(res).Encode(utils.HttpResponse{Success: false, Response: "Job ID not found"})

	} else if err == redis.Nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Client entry not present in cache")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(utils.HttpResponse{Success: false, Response: "Client entry not present in cache"})
	} else {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error while connecting to cache at the moment} -> error : %+v", err)
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(utils.HttpResponse{Success: false, Response: "Error while connecting to cache at the moment"})
	}
}

func (webSocketServer *WebSocketServer) handlePingRequests(res http.ResponseWriter, req *http.Request) {

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(utils.HttpResponse{Success: true, Response: "Pong"})
}
