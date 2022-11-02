package main

import (
	logger "Framework/Logger"
	"Framework/RPC"
	"net/rpc"
	"time"

	"github.com/google/uuid"
)

//returns a pointer a worker and runs it
func NewWorker() *Worker {
	worker := &Worker{
		id: uuid.NewString(), //random id
	}

	logger.LogInfo(logger.WORKER, logger.ESSENTIAL, "Worker is now alive")

	go worker.work()

	return worker
}

func (worker *Worker) work() {
	//endless for loop that keeps asking for tasks from the master
	for {
		args := &RPC.GetTaskArgs{
			WorkerId: worker.id,
		}
		reply := &RPC.GetTaskReply{}

		ok := worker.callMaster("Master.HandleGetTasks", args, reply)

		if !ok {
			logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleGetTasks")
			continue
		}

		if !reply.TaskAvailable {
			logger.LogInfo(logger.WORKER, logger.DEBUGGING, "Master doesn't have available tasks")
			time.Sleep(10 * time.Second)
			continue
		}

		logger.LogInfo(logger.WORKER, logger.ESSENTIAL, "This is the response received from the master %+v", reply)
	}
}

//blocking
func (worker *Worker) callMaster(rpcName string, args *RPC.GetTaskArgs, reply *RPC.GetTaskReply) bool {
	ctr := 1
	successfullConnection := false
	var client *rpc.Client
	var err error

	//attempt to conncect to master
	for ctr <= 3 && !successfullConnection {
		client, err = rpc.DialHTTP("tcp", MasterHost+":"+MasterPort) //blocking
		if err != nil {
			logger.LogError(logger.WORKER, logger.ESSENTIAL, "Attempt number %v of dialing master failed with error: %v\n", ctr, err)
			time.Sleep(10 * time.Second)
		} else {
			successfullConnection = true
		}
		ctr++
	}
	if !successfullConnection {
		logger.FailOnError(logger.WORKER, logger.ESSENTIAL, "Error dialing http: %v\nFatal Error: Can't establish connection to master. Exiting now", err)
	}

	defer client.Close()

	err = client.Call(rpcName, args, reply)

	if err != nil {
		logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master with RPC with error: %v", err)
		return false
	}
	logger.LogInfo(logger.WORKER, logger.DEBUGGING, "Success dialing master")

	return true
}
