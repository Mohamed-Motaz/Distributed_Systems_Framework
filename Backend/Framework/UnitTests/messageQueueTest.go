package main

import (
	logger "Framework/Logger"
	q "Framework/MessageQueue"
	"bytes"
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

	//convert to byte array
	bytesArr := new(bytes.Buffer)
	err := json.NewEncoder(bytesArr).Encode(assignedJob)
	if err != nil {
		logger.LogError(logger.MESSAGE_Q, logger.ESSENTIAL, "Unable to publish message to queue %+v", q.ASSIGNED_JOBS_QUEUE)
	}

	mqObj.Publish(q.ASSIGNED_JOBS_QUEUE, bytesArr.Bytes())

	time.Sleep(time.Second * 10)

	//now get a channel to consume from the assigned jobs queue

	ch, err := mqObj.Consume(q.ASSIGNED_JOBS_QUEUE)
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
