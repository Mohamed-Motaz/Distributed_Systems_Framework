package main

import (
	logger "Framework/Logger"
	q "Framework/MessageQueue"
	"encoding/json"
	"time"
)

func main() {
	mqObj := q.NewMQ(q.CreateMQAddress(q.MqUsername, q.MqPassword, q.MqHost, q.MqPort))

	assignedJob := &q.AssignedJob{
		ClientId: "10",
		JobId:    "11",
		Content:  "This is the content",
	}
	//[0,1,1,1,1,1,0,0,01,01,101,1]

	//convert to byte array
	bytesArr, err := json.Marshal(assignedJob)
	if err != nil {
		logger.LogError(logger.MESSAGE_Q, logger.ESSENTIAL, "Unable to convert struct %+v to byte array", assignedJob)
	}

	err = mqObj.Enqueue(q.ASSIGNED_JOBS_QUEUE, bytesArr)
	if err != nil {
		logger.LogError(logger.MESSAGE_Q, logger.ESSENTIAL, "Unable to push message %+v to q %+v", assignedJob, q.ASSIGNED_JOBS_QUEUE)

	}

	time.Sleep(time.Second * 10)

	//now get a channel to consume from the assigned jobs queue

	ch, err := mqObj.Dequeue(q.ASSIGNED_JOBS_QUEUE)
	if err != nil {
		logger.LogError(logger.MESSAGE_Q, logger.ESSENTIAL, "Unable to consume message from queue %+v", q.ASSIGNED_JOBS_QUEUE)
	}

	data := <-ch
	consumedByteArr := data.Body
	receivedAssignedJob := &q.AssignedJob{}

	err = json.Unmarshal(consumedByteArr, receivedAssignedJob)
	if err != nil {
		logger.LogError(logger.MESSAGE_Q, logger.ESSENTIAL, "Unable to unmarshal entity %v with error %v\nWill discard it", string(consumedByteArr), err)
		data.Ack(false)
	}

	logger.LogInfo(logger.MESSAGE_Q, logger.ESSENTIAL, "Received this entity from %+v from the queue", receivedAssignedJob)

	data.Ack(false) //ack after everything is done. This should be blocking other workers which is definetly a bottleneck

}
