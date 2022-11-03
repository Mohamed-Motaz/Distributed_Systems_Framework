package main

import (
	logger "Framework/Logger"
	"Framework/RPC"
	"net/rpc"
	"os/exec"
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
		getTaskArgs := &RPC.GetTaskArgs{
			WorkerId: worker.id,
		}
		getTaskReply := &RPC.GetTaskReply{}

		ok := worker.call("Master.HandleGetTasks", getTaskArgs, getTaskReply)
		if !ok {
			logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleGetTasks")
			continue
		}

		if !getTaskReply.TaskAvailable {
			logger.LogInfo(logger.WORKER, logger.DEBUGGING, "Master doesn't have available tasks")
			time.Sleep(10 * time.Second)
			continue
		}

		logger.LogInfo(logger.WORKER, logger.ESSENTIAL, "This is the response received from the master %+v", getTaskReply)

		// Now ready to call the process.exe
		// TODO: we will need to write the task to a file and make sure the process can read from a file
		time.Sleep(3*time.Second)

		worker.handleTask(getTaskReply)

		time.Sleep(10*time.Second)

	}

}

func (worker *Worker) handleTask(getTaskReply *RPC.GetTaskReply) {
	output, err := exec.Command(ProcessExeCmd).Output()
	if err != nil {
		logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to excute the client process with err: %+v",err)	

		finishedTaskArgs := &RPC.FinishedTaskArgs{IsSuccess: false}
		finishedTaskReply := &RPC.FinishedTaskReply{}

		ok := worker.call("Master.HandleFinishedTasks", finishedTaskArgs, finishedTaskReply)
		if !ok {
			logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleFinishedTasks")	
		}
		return
	}

	finishedTaskArgs := &RPC.FinishedTaskArgs{
		TaskId: getTaskReply.TaskId,
		JobId: getTaskReply.JobId,
		TaskResult: string(output),
		IsSuccess: true,
	}
	finishedTaskReply := &RPC.FinishedTaskReply{}

	ok := worker.call("Master.HandleFinishedTasks", finishedTaskArgs, finishedTaskReply)	
	if !ok {
		logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleFinishedTasks")	
	}
}

//blocking
func (worker *Worker) call(rpcName string, args interface{}, reply interface{}) bool {
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

	return true
}
