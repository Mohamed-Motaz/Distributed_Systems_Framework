package main

import (
	"Framework/Logger"
	logger "Framework/Logger"
	q "Framework/MessageQueue"
	"encoding/json"
	"time"
)

func main() {
	mqObj := q.NewMQ(q.CreateMQAddress(q.MqUsername, q.MqPassword, q.MqHost, q.MqPort))

	// ana kda b3ml job gdeda ya rewrew
	assignedJob := &q.AssignedJob{
		ClientId: "10",
		JobId:    "11",
		Content:  "REWREW",
	}

	// hn7wl el job de le array of bytes

	// marshal de bt7wl mn obj -> arr o bytes
	bytesArr, err := json.Marshal(assignedJob)

	if err != nil {
		logger.LogError(logger.MESSAGE_Q, logger.ESSENTIAL, "MSH 3REF A7WEL EL OBJ L ARR OF BYTES SHOOFO EL MOSHKLA FEN ", assignedJob)
	}

	// ro7 7otly fl queue bt3t el assigned jobs el job el gdedda de yala
	err = mqObj.Enqueue(q.ASSIGNED_JOBS_QUEUE, bytesArr)

	if err != nil {
		logger.LogError(logger.MESSAGE_Q, logger.ESSENTIAL, "yoh fe moshkla", q.ASSIGNED_JOBS_QUEUE)
	}

	time.Sleep(time.Second * 10)

	// dlwty ehna kda 3mlna l assign
	// tb dlwty yala n3ml  el dequeue

	// dequeue btrg3 error w channel

	ch, err := mqObj.Dequeue(q.ASSIGNED_JOBS_QUEUE)

	if err != nil {
		logger.LogError(logger.MESSAGE_Q, logger.ESSENTIAL, "fe zft moshkla", q.ASSIGNED_JOBS_QUEUE)
	}
	// dlwty el channel l mb3otly da struct 3ndo body 3shan ageb el data b2a .body

	data := <-ch
	consumedByteArr := data.Body
	// AssignedJob de struct shyel clientid, jobid, content
	// faana b3ml obj mn l struct da 3shan ashel feh
	receivedAssignedJob := &q.AssignedJob{}

	// w zi m2oltlko fo2 3ala el Marshal khlene aaolko dlwty 3la el Unmarshal
	// el unMarshal de bt7wl mn bytesArr -> obj el hwa struct
	err = json.Unmarshal(consumedByteArr, receivedAssignedJob)

	if err != nil {
		logger.LogError(logger.MESSAGE_Q, Logger.ESSENTIAL, "shofo mshklko b2a", string(consumedByteArr), err)
		data.Ack(false)
	}

	logger.LogInfo(logger.MESSAGE_Q, logger.ESSENTIAL, "Received this entity from %+v from the queue", receivedAssignedJob)
	data.Ack(false)
}
