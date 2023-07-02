package main

import (
	logger "Framework/Logger"
	utils "Framework/Utils"
	"fmt"

	"Framework/RPC"
	"time"
)


//
//RPC handlers
//

func (master *Master) HandleGetTasks(args *RPC.GetTaskArgs, reply *RPC.GetTaskReply) error {
	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Worker called HandleGetTasks with these args %+v", args)

	master.mu.Lock()
	defer master.mu.Unlock()
	defer logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Master replied with this reply: %+v", struct {
		TaskAvailable        bool
		TaskContent          string
		ProcessBinaryName    string
		OptionalFilesZipName string
		TaskId               string
		JobId                string
	}{TaskAvailable: reply.TaskAvailable, TaskContent: reply.TaskContent,
		ProcessBinaryName: reply.ProcessBinary.Name, OptionalFilesZipName: reply.OptionalFilesZip.Name,
		TaskId: reply.TaskId, JobId: reply.JobId})

	if !master.isRunning {
		reply.TaskAvailable = false
		return nil
	}

	//now need to check the available tasks

	for i := range master.currentJob.tasks {
		currentTask := master.currentJob.tasks[i]
		currentWorkerAndTimer := master.currentJob.workersTimers[i]

		if currentTask.isDone {
			continue
		}
		if time.Since(currentWorkerAndTimer.lastHeartBeat) > master.maxHeartBeatTimer {

			//the other worker is probably dead, so give this worker this job
			reply.TaskAvailable = true
			reply.TaskContent = currentTask.content
			reply.ProcessBinary = master.currentJob.processBinary
			reply.OptionalFilesZip = master.currentJob.optionalFilesZip
			reply.TaskId = currentTask.id
			reply.JobId = master.currentJob.jobId

			if args.ProcessBinaryId == master.currentJob.processBinary.Id {
				reply.ProcessBinary = utils.RunnableFile{Id: master.currentJob.processBinary.Id}
			}

			if args.JobId == master.currentJob.jobId {
				reply.OptionalFilesZip = utils.File{}
			}

			//now as a master, need to mark this job as given to a worker
			master.currentJob.workersTimers[i] = WorkerAndHisTimer{
				lastHeartBeat: time.Now(),
				workerId:      args.WorkerId,
			}

			return nil
		}

	}

	reply.TaskAvailable = false
	return nil
}

func (master *Master) HandleFinishedTasks(args *RPC.FinishedTaskArgs, reply *RPC.FinishedTaskReply) error {
	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Worker called HandleFinishedTasks with these args %+v", args)
	master.mu.Lock()
	defer master.mu.Unlock()

	if !master.isRunning {
		return nil
	}

	if args.JobId != master.currentJob.jobId {
		return nil
	}

	if args.Err {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Worker sent this error %+v", args.ErrMsg)
		master.publishErrAsfinishedJob(fmt.Sprintf("Worker sent this error: %+v", args.ErrMsg), master.currentJob.clientId, master.currentJob.jobId)
		return nil
	}

	taskIndex := master.getTaskIndexByTaskId(args.TaskId)
	if taskIndex == -1 {
		return nil
	}

	//now need to write the results to a file, and save this files location
	filePath := args.TaskId + ".txt"
	err := utils.CreateAndWriteToFile(filePath, []byte(args.TaskResult))
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "error while creating the task file %+v", err)
		master.publishErrAsfinishedJob(
			fmt.Sprintf("Error while saving worker's task locally on the master: %+v", err),
			master.currentJob.clientId,
			master.currentJob.jobId)
		return nil
	}

	master.currentJob.tasks[taskIndex].isDone = true
	master.currentJob.finishedTasksFilePaths[taskIndex] = filePath
	master.currentJob.workersTimers[taskIndex].lastHeartBeat = time.Now()

	//check if all tasks are done and aggregate the results
	jobDone := master.allTasksDone()
	if !jobDone {
		return nil
	}

	//all tasks have been finished!
	master.aggregateTasks()
	return nil
}


func (master *Master) HandleWorkerHeartBeats(args *RPC.WorkerHeartBeatArgs, reply *RPC.WorkerHeartBeatReply) error {
	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Worker called HandleWorkerHeartBeat with these args %+v", args)

	master.mu.Lock()
	defer master.mu.Unlock()

	if !master.isRunning {
		return nil
	}

	if args.JobId != master.currentJob.jobId {
		return nil
	}

	taskIndex := master.getTaskIndexByTaskId(args.TaskId)
	if taskIndex == -1 {
		return nil
	}

	//now make sure that this worker was actually assigned this task
	if master.currentJob.workersTimers[taskIndex].workerId != args.WorkerId {
		return nil
	}

	master.currentJob.workersTimers[taskIndex].lastHeartBeat = time.Now()

	return nil
}