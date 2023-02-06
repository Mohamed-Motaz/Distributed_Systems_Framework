package main

import (
	mq "Framework/MessageQueue"
	logger "Framework/Logger"
	utils "Framework/Utils"
)


func (webSocketServer *WebSocketServer) writeFinishedJob(client *Client, finishedJob mq.FinishedJob) {
	client.webSocketConn.WriteJSON(finishedJob)
	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Job sent to client} %+v\n%+v", client.webSocketConn.RemoteAddr(), finishedJob)
}

func (webSocketServer *WebSocketServer) writeError(client *Client, err utils.Error) {
	client.webSocketConn.WriteJSON(err)
	logger.LogInfo(logger.WEBSOCKET_SERVER, logger.ESSENTIAL, "{Error sent to client} %+v\n%+v", client.webSocketConn.RemoteAddr(), err)
}

func (webSocketServer *WebSocketServer) modifyJobRequest(jobRequest *JobRequest, modifiedJobRequest *mq.AssignedJob) {

	modifiedJobRequest.JobId = jobRequest.JobId
	modifiedJobRequest.JobContent = jobRequest.JobContent
	modifiedJobRequest.DistributeBinaryName = jobRequest.DistributeBinaryName
	modifiedJobRequest.ProcessBinaryName = jobRequest.ProcessBinaryName
	modifiedJobRequest.AggregateBinaryName = jobRequest.AggregateBinaryName
}