package main

import (
	logger "Framework/Logger"
	"Framework/RPC"
	utils "Framework/Utils"
	"net/rpc"
	"os"
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

		ok := worker.callMaster("Master.HandleGetTasks", getTaskArgs, getTaskReply)
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

		worker.handleTask(getTaskReply)

	}

}

func (worker *Worker) handleTask(getTaskReply *RPC.GetTaskReply) {
	stopHeartBeatsCh := make(chan bool)
	go worker.startHeartBeats(getTaskReply, stopHeartBeatsCh)
	defer func() {
		logger.LogInfo(logger.WORKER, logger.DEBUGGING, "Worker done with handleTask with this data %+v", getTaskReply)
		stopHeartBeatsCh <- true
	}()

	//now, need to run process
	fPath := "./process.txt"
	err := utils.CreateAndWriteToFile(fPath, []byte(getTaskReply.TaskContent))
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "error while creating the temporary file that contains the task contents for process locally on the worker %+v", err)
		//return fmt.Errorf("error while creating the temporary file that contains the task contents for process locally on the worker")
		//todo handle this error
		return
	}

	_, err = exec.Command("./" + getTaskReply.ProcessExe.Name).Output()
	if err != nil {
		logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to excute the client process with err: %+v", err)

		finishedTaskArgs := &RPC.FinishedTaskArgs{IsSuccess: false}
		finishedTaskReply := &RPC.FinishedTaskReply{}

		ok := worker.callMaster("Master.HandleFinishedTasks", finishedTaskArgs, finishedTaskReply)
		if !ok {
			logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleFinishedTasks")
		}
		return
	}

	//now need to read from this file the resulting data
	data, err := os.ReadFile(fPath)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "error while reading from the distribute process %+v", err)
		//return fmt.Errorf("error while reading from the distribute process")
		//todo handle this error
		return
	}

	finishedTaskArgs := &RPC.FinishedTaskArgs{
		TaskId:     getTaskReply.TaskId,
		JobId:      getTaskReply.JobId,
		TaskResult: string(data),
		IsSuccess:  true,
	}
	finishedTaskReply := &RPC.FinishedTaskReply{}

	ok := worker.callMaster("Master.HandleFinishedTasks", finishedTaskArgs, finishedTaskReply)
	if !ok {
		logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleFinishedTasks")
	}
}

func (worker *Worker) startHeartBeats(getTaskReply *RPC.GetTaskReply, stopHeartBeats chan bool) {
	logger.LogInfo(logger.WORKER, logger.DEBUGGING, "About to start sending heartbeats for this task %+v", getTaskReply)

	for {
		select {
		case <-stopHeartBeats:
			logger.LogInfo(logger.WORKER, logger.DEBUGGING, "Stopped sending heartbeats for this task %+v", getTaskReply)
			return
		default:
			time.Sleep(9 * time.Second)
			args := &RPC.WorkerHeartBeatArgs{
				WorkerId: worker.id,
				TaskId:   getTaskReply.TaskId,
				JobId:    getTaskReply.JobId,
			}
			reply := &RPC.WorkerHeartBeatReply{}
			ok := worker.callMaster("Master.HandleWorkerHeartBeats", args, reply)
			if !ok {
				logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleWorkerHeartBeats")
			}
		}
	}
}

//blocking
func (worker *Worker) callMaster(rpcName string, args interface{}, reply interface{}) bool {
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
