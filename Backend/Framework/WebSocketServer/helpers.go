package main

import (
	cache "Framework/Cache"
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	"Framework/RPC"
	utils "Framework/Utils"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func (webSocketServer *WebSocketServer) writeResp(client *Client, res WsResponse) {
	client.webSocketConn.WriteJSON(res)
	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Resp sent to client} %+v\n%+v", client.webSocketConn.RemoteAddr(), res)
}

func (webSocketServer *WebSocketServer) modifyJobRequest(jobRequest *JobRequest, modifiedJobRequest *mq.AssignedJob) {

	modifiedJobRequest.JobId = jobRequest.JobId
	modifiedJobRequest.JobContent = jobRequest.JobContent
	modifiedJobRequest.DistributeBinaryId = jobRequest.DistributeBinaryId
	modifiedJobRequest.ProcessBinaryId = jobRequest.ProcessBinaryId
	modifiedJobRequest.AggregateBinaryId = jobRequest.AggregateBinaryId
}

func (websocketServer *WebSocketServer) sendOptionalFilesToLockserver(client *Client, newJobRequest *JobRequest) WsResponse {

	res := WsResponse{MsgType: JOB_REQUEST, Response: utils.HttpResponse{Success: true, Response: ""}}

	if len(newJobRequest.OptionalFilesZip.Content) == 0 {
		return res
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
		Receiver: RPC.Receiver{
			Name: "Lockserver",
			Port: LockServerPort,
			Host: LockServerHost,
		},
	})

	if !ok {

		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with connecting lockServer} -> error : %+v", err)
		res.Response = utils.HttpResponse{Success: false, Response: ("Error with connecting lockServer")}

	} else if reply.Err {

		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with uploading files to lockServer} -> error : %+v", reply.ErrMsg)
		res.Response = utils.HttpResponse{Success: false, Response: (fmt.Sprintf("Error with uploading files to lockServer: %+v", reply.ErrMsg))}

	} else {
		logger.LogInfo(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Optional Files sent to lockServer successfully")
	}

	return res
}

func (webSocketServer *WebSocketServer) GetFinishedJobsIds(client *Client) WsResponse {

	res := WsResponse{
		MsgType: FINISHED_JOBS_IDS,
	}

	var finishedJobs *cache.CacheValue
	finishedJobs, err := webSocketServer.cache.Get(client.id)

	if err == nil {

		finishedJobsIds := make([]string, 0)

		for _, finishedJob := range finishedJobs.FinishedJobs {
			finishedJobsIds = append(finishedJobsIds, finishedJob.JobId)
		}

		logger.LogInfo(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "{Finished jobs ids sent to client} -> jobs Ids : %+v", finishedJobsIds)
		res.Response = (utils.HttpResponse{Success: true, Response: finishedJobsIds})

	} else if err == redis.Nil {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.DEBUGGING, "Client entry not present in cache")
		res.Response = (utils.HttpResponse{Success: false, Response: "Client entry not present in cache"})
	} else {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error while connecting to cache at the moment} -> error : %+v", err)
		res.Response = (utils.HttpResponse{Success: false, Response: "Error while connecting to cache at the moment"})
	}
	return res
}

func (webSocketServer *WebSocketServer) GetSystemBinaries() WsResponse {

	res := WsResponse{
		MsgType: SYSTEM_BINARIES,
	}

	reply := &RPC.GetBinaryFilesReply{}

	ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
		Name:         "LockServer.HandleGetBinaryFiles",
		Args:         &RPC.GetBinaryFilesArgs{},
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
		res.Response = utils.HttpResponse{Success: false, Response: "Unable to connect to lockserver"}

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with recieving files from lockServer} -> error : %+v", reply.ErrMsg)
		res.Response = utils.HttpResponse{Success: false, Response: fmt.Sprintf("Error with recieving files from lockServer %+v", reply.ErrMsg)}

	} else {
		res.Response = utils.HttpResponse{Success: true, Response: reply}
	}
	return res
}

func (webSocketServer *WebSocketServer) GetSystemProgress() WsResponse {

	res := WsResponse{
		MsgType: SYSTEM_PROGRESS,
	}

	reply := &RPC.GetSystemProgressReply{}

	ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
		Name:         "LockServer.HandleGetSystemProgress",
		Args:         &RPC.GetSystemProgressArgs{},
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
		res.Response = (utils.HttpResponse{Success: false, Response: "Unable to connect to lockserver"})

	} else if reply.Err {
		logger.LogError(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error with Getting system progress from lockServer} -> error : %+v", reply.ErrMsg)
		res.Response = (utils.HttpResponse{Success: false, Response: fmt.Sprintf("Error with Getting system progress from lockServer %+v", reply.ErrMsg)})
	} else {
		res.Response = (utils.HttpResponse{Success: true, Response: reply.Progress})
	}
	return res
}
