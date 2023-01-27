package main

import (
	common "Framework/Common"
	logger "Framework/Logger"
	"Framework/RPC"
	utils "Framework/Utils"
	"time"

	"github.com/google/uuid"
)

// returns a pointer a worker and runs it
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

		rpcConn := &RPC.RpcConnection{
			Name:         "Master.HandleGetTasks",
			Args:         getTaskArgs,
			Reply:        &getTaskReply,
			SenderLogger: logger.WORKER,
			Reciever: RPC.Reciever{
				Name: "Master",
				Port: MasterPort,
				Host: MasterHost,
			},
		}
		ok, err := RPC.EstablishRpcConnection(rpcConn)
		if !ok {
			logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleGetTasks with error -> %v", err)
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
	data, err := common.BinarycuteProcess(logger.MASTER, utils.ProcessBinary,
		utils.File{Name: "process.txt", Content: []byte(getTaskReply.TaskContent)},
		getTaskReply.ProcessBinary)

	if err != nil {
		logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to excute the client process with err: %+v", err)

		finishedTaskArgs := &RPC.FinishedTaskArgs{Error: utils.Error{Err: true, ErrMsg: "Error while binarycuting process binary on the worker"}}
		finishedTaskReply := &RPC.FinishedTaskReply{}

		rpcConn := &RPC.RpcConnection{
			Name:         "Master.HandleFinishedTasks",
			Args:         finishedTaskArgs,
			Reply:        &finishedTaskReply,
			SenderLogger: logger.WORKER,
			Reciever: RPC.Reciever{
				Name: "Master",
				Port: MasterPort,
				Host: MasterHost,
			},
		}
		ok, err := RPC.EstablishRpcConnection(rpcConn)

		if !ok {
			logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleFinishedTasks with error -> %v", err)
		}
		return
	}

	finishedTaskArgs := &RPC.FinishedTaskArgs{
		TaskId:     getTaskReply.TaskId,
		JobId:      getTaskReply.JobId,
		TaskResult: string(data),
		Error:      utils.Error{Err: false},
	}
	finishedTaskReply := &RPC.FinishedTaskReply{}

	rpcConn := &RPC.RpcConnection{
		Name:         "Master.HandleFinishedTasks",
		Args:         finishedTaskArgs,
		Reply:        &finishedTaskReply,
		SenderLogger: logger.WORKER,
		Reciever: RPC.Reciever{
			Name: "Master",
			Port: MasterPort,
			Host: MasterHost,
		},
	}
	ok, err := RPC.EstablishRpcConnection(rpcConn)
	if !ok {
		logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleFinishedTasks with error -> %v", err)
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

			rpcConn := &RPC.RpcConnection{
				Name:         "Master.HandleWorkerHeartBeats",
				Args:         args,
				Reply:        &reply,
				SenderLogger: logger.WORKER,
				Reciever: RPC.Reciever{
					Name: "Master",
					Port: MasterPort,
					Host: MasterHost,
				},
			}
			ok, err := RPC.EstablishRpcConnection(rpcConn)
			if !ok {
				logger.LogError(logger.WORKER, logger.ESSENTIAL, "Unable to call master HandleWorkerHeartBeats with error -> %v", err)
			}
		}
	}
}
