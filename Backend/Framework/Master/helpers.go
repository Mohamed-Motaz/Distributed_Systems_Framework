package main

import (
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	utils "Framework/Utils"
	"fmt"
	"sort"

	"Framework/RPC"
)


func CreateMasterAddress() string {
	return MyHost + ":" + MyPort
}


// this function expects to hold a lock because master.publishFinishedJob needs to hold a lock
// it resets the job status completely
func (master *Master) publishErrAsFinishedJob(err, clientId, jobId string) { //DONE fix
	fn := mq.FinishedJob{}
	fn.ClientId = clientId
	fn.JobId = jobId
	fn.Err = true
	fn.ErrMsg = err
	master.publishFinishedJob(fn, true)
}

// this function expects to hold a lock
// returns -1 if task not found
func (master *Master) getTaskIndexByTaskId(taskId string) int {

	for i := range master.currentJob.tasks {
		if master.currentJob.tasks[i].id == taskId {
			return i
		}
	}

	return -1
}

// this function expects to hold a lock
func (master *Master) allTasksDone() bool {
	if !master.isRunning {
		return false //no tasks in the first place
	}

	for _, t := range master.currentJob.tasks {
		if !t.isDone {
			return false
		}
	}
	return true
}

func (master *Master) writeBinariesAndFilesOnDisk() error {

	//now write the distribute, aggregate, and optionalFilesZip  to disk
	if err := utils.CreateAndWriteToFile(master.currentJob.distributeBinary.Name, master.currentJob.distributeBinary.Content); err != nil {
		return fmt.Errorf("error while creating the distribute binary zip file %+v", err)
	}
	if err := utils.UnzipSource(master.currentJob.distributeBinary.Name, ""); err != nil {
		return fmt.Errorf("error while unzipping distribute zip %+v", err)
	}

	if err := utils.CreateAndWriteToFile(master.currentJob.aggregateBinary.Name, master.currentJob.aggregateBinary.Content); err != nil {
		return fmt.Errorf("error while creating the aggregate binary zip file %+v", err)
	}
	if err := utils.UnzipSource(master.currentJob.aggregateBinary.Name, ""); err != nil {
		return fmt.Errorf("error while unzipping aggregate zip %+v", err)
	}

	if err := utils.CreateAndWriteToFile(master.currentJob.optionalFilesZip.Name, master.currentJob.optionalFilesZip.Content); err != nil {
		return fmt.Errorf("error while creating the optional files zip file %+v", err)
	}
	if err := utils.UnzipSource(master.currentJob.optionalFilesZip.Name, ""); err != nil {
		return fmt.Errorf("error while unzipping optional files zip %+v", err)
	}

	return nil
}

// this function generates the WorkersTasks field
// this function expects to hold a lock
func (master *Master) generateWorkersTasks() []RPC.WorkerTask {
	workersTasks := make([]RPC.WorkerTask, 0)
	workersMp := make(map[string]*RPC.WorkerTask)
	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "These are the workers data: %+v", master.currentJob.workersTimers)

	for i, workerTimer := range master.currentJob.workersTimers {
		logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "These are the workers data: %+v ---\n---\n%+v", master.currentJob.workersTimers[i], master.currentJob.tasks[i])

		if workerTimer.workerId == "" { //task isn't yet assigned
			continue
		}
		//init a new key for the worker if it doesn't exist in the map
		if _, ok := workersMp[workerTimer.workerId]; !ok {
			workersMp[workerTimer.workerId] = &RPC.WorkerTask{
				FinishedTasksContent: make([]string, 0),
			}
		}

		task := master.currentJob.tasks[i]

		if task.isDone {
			workersMp[workerTimer.workerId].FinishedTasksContent =
				append(workersMp[workerTimer.workerId].FinishedTasksContent, task.content)
		} else if workerTimer.lastHeartBeat.UnixMilli() > 0 {
			//check if the worker is currently working
			workersMp[workerTimer.workerId].CurrentTaskContent = task.content
		}
	}

	for k, v := range workersMp {
		workerT := v
		workerT.WorkerId = k
		workersTasks = append(workersTasks, *workerT)
	}

	//sort the workers
	sort.Slice(workersTasks, func(i, j int) bool {
		return workersTasks[i].WorkerId < workersTasks[j].WorkerId
	})
	return workersTasks
}

// func (master *Master) removeSliceElementByIndex (arr *[]Task, index int) int {

// 	// Shift a[i+1:] left one index.
// 	copy((*arr)[index:], (*arr)[index+1:])
// 	// Erase last element (write zero value).
// 	(*arr)[len((*arr))-1] = Task{}
// 	// Truncate slice.
// 	(*arr) = (*arr)[:len((*arr))-1]
// }
