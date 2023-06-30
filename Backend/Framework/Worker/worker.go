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
		id: uuid.NewString(), //random id
		ProcessBinary: utils.RunnableFile{},
		OptionalFilesZip: utils.File{},
		JobId: "",
	}

	logger.LogInfo(logger.WORKER, logger.ESSENTIAL, "Worker is now alive")

	go worker.work()

	return worker
}

func (worker *Worker) work() {
	//endless for loop that keeps asking for tasks from the master
	for {
		//clean any old files
		utils.RemoveFilesThatDontMatchNames(FileNamesToIgnore)

		getTaskArgs := &RPC.GetTaskArgs{
			WorkerId: worker.id,
			ProcessBinaryId: worker.ProcessBinary.Id,
			JobId: worker.JobId,
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
			time.Sleep(5 * time.Second)
			continue
		}
		
		if getTaskReply.JobId == worker.JobId{
			getTaskReply.OptionalFilesZip = worker.OptionalFilesZip
		}

		if getTaskReply.ProcessBinary.Id == worker.ProcessBinary.Id{
			getTaskReply.ProcessBinary = worker.ProcessBinary
		}

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

		worker.JobId = getTaskReply.JobId
		worker.OptionalFilesZip = getTaskReply.OptionalFilesZip
		worker.ProcessBinary = getTaskReply.ProcessBinary

		//do this needs a goroutine
		go worker.handleTask(getTaskReply)

	}

}

func (worker *Worker) handleTask(getTaskReply *RPC.GetTaskReply) {
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

	//write the process to disk
	if err := utils.CreateAndWriteToFile(getTaskReply.ProcessBinary.Name, getTaskReply.ProcessBinary.Content); err != nil {
		worker.callMasterWithError(fmt.Sprintf("Error while creating the process binary zip file %+v", err), "Error while creating the process binary zip file")
		return
	}
	if err := utils.UnzipSource(getTaskReply.ProcessBinary.Name, ""); err != nil {
		worker.callMasterWithError(fmt.Sprintf("Unable to unzip the client process with err: %+v", err), "Error while unzipping process binary on the worker")
		return
	}

	//write the optional files to disk
	if err := utils.CreateAndWriteToFile(getTaskReply.OptionalFilesZip.Name, getTaskReply.OptionalFilesZip.Content); err != nil {
		worker.callMasterWithError(fmt.Sprintf("Error while creating the optional files zip file %+v", err), "Error while creating the optional files zip file")
		return
	}
	if err := utils.UnzipSource(getTaskReply.OptionalFilesZip.Name, ""); err != nil {
		worker.callMasterWithError(fmt.Sprintf("Unable to unzip the client optional files with err: %+v", err), "Error while unzipping the optional files on the worker")
		return
	}

	//now, need to run process
	data, err := utils.ExecuteProcess(logger.MASTER, utils.ProcessBinary,
		utils.File{Name: "process.txt", Content: []byte(getTaskReply.TaskContent)},
		getTaskReply.ProcessBinary)

	if err != nil {
		worker.callMasterWithError(fmt.Sprintf("Unable to excute the client process with err: %+v", err), "Error while execute process binary on the worker")
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

func (worker *Worker) callMasterWithError(logErrorMessage, masterErrorMessage string) {
	logger.LogError(logger.WORKER, logger.ESSENTIAL, logErrorMessage)

	finishedTaskArgs := &RPC.FinishedTaskArgs{Error: utils.Error{Err: true, ErrMsg: masterErrorMessage}}
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
