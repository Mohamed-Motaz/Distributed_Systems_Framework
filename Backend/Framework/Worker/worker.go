package main

import (
	logger "Framework/Logger"
	"Framework/RPC"
	utils "Framework/Utils"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// returns a pointer a worker and runs it
func NewWorker() *Worker {

	worker := &Worker{
		id:               uuid.NewString(), //random id
		ProcessBinary:    utils.RunnableFile{},
		OptionalFilesZip: utils.File{},
		JobId:            "",
	}

	logger.LogInfo(logger.WORKER, logger.ESSENTIAL, "Worker is now alive")

	utils.KeepFilesThatMatch(FileNamesToIgnore)

	go worker.askForWork()

	return worker
}

func (worker *Worker) askForWork() {
	//endless for loop that keeps asking for tasks from the master

	for {


		getTaskArgs := &RPC.GetTaskArgs{
			WorkerId:        worker.id,
			ProcessBinaryId: worker.ProcessBinary.Id,
			JobId:           worker.JobId,
		}

		getTaskReply := &RPC.GetTaskReply{}

		rpcConn := &RPC.RpcConnection{
			Name:         "Master.HandleGetTasks",
			Args:         getTaskArgs,
			Reply:        &getTaskReply,
			SenderLogger: logger.WORKER,
			Receiver: RPC.Receiver{
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
			time.Sleep(5 * time.Second)
			continue
		}

		if getTaskReply.ProcessBinary.Id == worker.ProcessBinary.Id {
			//process already exists on disk
			getTaskReply.ProcessBinary = worker.ProcessBinary

		} else {

			utils.KeepFilesThatMatch(FileNamesToIgnore)

			//write process on disk
			if err := utils.CreateAndWriteToFile(getTaskReply.ProcessBinary.Name, getTaskReply.ProcessBinary.Content); err != nil {
				worker.sendFinishedTaskAsErr(fmt.Sprintf("Error while creating the process binary zip file %+v", err), "Error while creating the process binary zip file")
				continue
			}
			if err := utils.UnzipSource(getTaskReply.ProcessBinary.Name, ""); err != nil {
				worker.sendFinishedTaskAsErr(fmt.Sprintf("Unable to unzip the client process with err: %+v", err), "Error while unzipping process binary on the worker")
				continue
			}
		}
		worker.ProcessBinary = getTaskReply.ProcessBinary

		if getTaskReply.JobId == worker.JobId {
			getTaskReply.OptionalFilesZip = worker.OptionalFilesZip

		} else {

			if err := utils.CreateAndWriteToFile(getTaskReply.OptionalFilesZip.Name, getTaskReply.OptionalFilesZip.Content); err != nil {
				worker.sendFinishedTaskAsErr(fmt.Sprintf("Error while creating the optional files zip file %+v", err), "Error while creating the optional files zip file")
				continue
			}
			if err := utils.UnzipSource(getTaskReply.OptionalFilesZip.Name, ""); err != nil {
				worker.sendFinishedTaskAsErr(fmt.Sprintf("Unable to unzip the client optional files with err: %+v", err), "Error while unzipping the optional files on the worker")
				continue
			}
		}
		worker.JobId = getTaskReply.JobId
		worker.OptionalFilesZip = getTaskReply.OptionalFilesZip

		logger.LogInfo(logger.WORKER, logger.ESSENTIAL, "This is the response received from the master %+v", struct {
			TaskAvailable        bool
			TaskContent          string
			ProcessBinaryName    string
			OptionalFilesZipName string
			TaskId               string
			JobId                string
		}{TaskAvailable: getTaskReply.TaskAvailable, TaskContent: getTaskReply.TaskContent, ProcessBinaryName: getTaskReply.ProcessBinary.Name,
			OptionalFilesZipName: getTaskReply.OptionalFilesZip.Name, TaskId: getTaskReply.TaskId, JobId: getTaskReply.TaskId},
		)

		//do this needs a goroutine
		go worker.doWork(getTaskReply)

	}

}

func (worker *Worker) doWork(getTaskReply *RPC.GetTaskReply) {
	stopHeartBeatsCh := make(chan bool)
	go worker.startHeartBeats(getTaskReply, stopHeartBeatsCh)

	defer func() {
		logger.LogInfo(logger.WORKER, logger.DEBUGGING, "Worker done with handleTask with this data %+v", struct {
			TaskAvailable        bool
			TaskContent          string
			ProcessBinaryName    string
			OptionalFilesZipName string
			TaskId               string
			JobId                string
		}{TaskAvailable: getTaskReply.TaskAvailable, TaskContent: getTaskReply.TaskContent, ProcessBinaryName: getTaskReply.ProcessBinary.Name,
			OptionalFilesZipName: getTaskReply.OptionalFilesZip.Name, TaskId: getTaskReply.TaskId, JobId: getTaskReply.TaskId})
		stopHeartBeatsCh <- true
	}()

	//now, need to run process
	data, err := utils.ExecuteProcess(logger.MASTER, utils.ProcessBinary,
		utils.File{Name: "process.txt", Content: []byte(getTaskReply.TaskContent)},
		getTaskReply.ProcessBinary)

	if err != nil {
		worker.sendFinishedTaskAsErr(fmt.Sprintf("Unable to execute the client process with err: %+v", err), "Error while execute process binary on the worker")
		return
	}

	finishedTaskArgs := &RPC.FinishedTaskArgs{
		TaskId:     getTaskReply.TaskId,
		JobId:      getTaskReply.JobId,
		TaskResult: string(data),
		Error:      utils.Error{Err: false},
	}
	
	worker.sendFinishedTask(finishedTaskArgs)
}

func (worker *Worker) startHeartBeats(getTaskReply *RPC.GetTaskReply, stopHeartBeats chan bool) {
	logger.LogInfo(logger.WORKER, logger.DEBUGGING, "About to start sending heartbeats for this task %+v", struct {
		TaskAvailable        bool
		TaskContent          string
		ProcessBinaryName    string
		OptionalFilesZipName string
		TaskId               string
		JobId                string
	}{TaskAvailable: getTaskReply.TaskAvailable, TaskContent: getTaskReply.TaskContent, ProcessBinaryName: getTaskReply.ProcessBinary.Name,
		OptionalFilesZipName: getTaskReply.OptionalFilesZip.Name, TaskId: getTaskReply.TaskId, JobId: getTaskReply.TaskId})

	for {
		select {
		case <-stopHeartBeats:
			logger.LogInfo(logger.WORKER, logger.DEBUGGING, "Stopped sending heartbeats for this task %+v", struct {
				TaskAvailable        bool
				TaskContent          string
				ProcessBinaryName    string
				OptionalFilesZipName string
				TaskId               string
				JobId                string
			}{TaskAvailable: getTaskReply.TaskAvailable, TaskContent: getTaskReply.TaskContent, ProcessBinaryName: getTaskReply.ProcessBinary.Name,
				OptionalFilesZipName: getTaskReply.OptionalFilesZip.Name, TaskId: getTaskReply.TaskId, JobId: getTaskReply.TaskId})
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
				Receiver: RPC.Receiver{
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

func (worker *Worker) sendFinishedTask(finishedTaskArgs *RPC.FinishedTaskArgs){

	finishedTaskReply := &RPC.FinishedTaskReply{}

	rpcConn := &RPC.RpcConnection{
		Name:         "Master.HandleFinishedTasks",
		Args:         finishedTaskArgs,
		Reply:        &finishedTaskReply,
		SenderLogger: logger.WORKER,
		Receiver: RPC.Receiver{
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

func (worker *Worker) sendFinishedTaskAsErr(logErrorMessage, masterErrorMessage string) {
	
	logger.LogError(logger.WORKER, logger.ESSENTIAL, logErrorMessage)

	finishedTaskArgs := &RPC.FinishedTaskArgs{Error: utils.Error{Err: true, ErrMsg: masterErrorMessage}}
	
	worker.sendFinishedTask(finishedTaskArgs)
}
