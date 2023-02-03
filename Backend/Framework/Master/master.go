package main

import (
	common "Framework/Common"
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	utils "Framework/Utils"
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"Framework/RPC"
	"encoding/gob"
	"encoding/json"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// returns a pointer a master and runs it
func NewMaster() *Master {
	master := &Master{
		id:                uuid.NewString(), //random id
		q:                 mq.NewMQ(mq.CreateMQAddress(MqUsername, MqPassword, MqHost, MqPort)),
		maxHeartBeatTimer: 30 * time.Second, //each heartbeat should be every 10 seconds but we allow up to 2 failures
		mu:                sync.Mutex{},
	}
	master.resetStatus()
	master.addDumbJob()

	go master.server()
	go master.qConsumer()
	go master.sendPeriodicProgress()

	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Master is now alive")

	return master
}

func CreateMasterAddress() string {
	return MyHost + ":" + MyPort
}

func (master *Master) removeOldJobFiles() {
	myName := filepath.Base(os.Args[0])
	utils.RemoveFilesThatDontMatchPrefix(myName)
}

// this function expects to hold a lock
// it resets the currentJob of the master
func (master *Master) resetStatus() {
	//todo support other oss

	master.removeOldJobFiles()

	master.currentJob = CurrentJob{
		tasks:                  make([]Task, 0),
		finishedTasksFilePaths: make([]string, 0),
		workersTimers:          make([]WorkerAndHisTimer, 0),
		processBinary: utils.RunnableFile{
			File: utils.File{
				Content: make([]byte, 0),
			},
		},
		distributeBinary: utils.RunnableFile{
			File: utils.File{
				Content: make([]byte, 0),
			},
		},
		aggregateBinary: utils.RunnableFile{
			File: utils.File{
				Content: make([]byte, 0),
			},
		},
		optionalFilesZip: utils.File{
			Content: make([]byte, 0),
		},
	}
	master.isRunning = false

}

// this function expects to hold a lock
// this function is responsible for setting up the new job and running distribute

// this function doesn't log. The caller is responsible for logging
func (master *Master) setJobStatus(reply *RPC.GetJobReply) error {
	master.isRunning = true

	master.currentJob = CurrentJob{
		clientId:   reply.ClientId,
		jobContent: reply.JobContent,
		jobId:      reply.JobId,

		tasks:                  make([]Task, 0),
		finishedTasksFilePaths: make([]string, 0),
		workersTimers:          make([]WorkerAndHisTimer, 0),
		processBinary:          reply.ProcessBinary,
		distributeBinary:       reply.DistributeBinary,
		aggregateBinary:        reply.AggregateBinary,
	}

	//todo, think about supporting different os binaries
	//todo any errors here should be propagated to client

	//now write the distribute and aggregate folders to disk

	if err := utils.UnzipSource(master.currentJob.distributeBinary.Name, ""); err != nil {
		return fmt.Errorf("error while unzipping distribute zip %+v", err)
	}

	if err := utils.UnzipSource(master.currentJob.aggregateBinary.Name, ""); err != nil {
		return fmt.Errorf("error while unzipping aggregate zip %+v", err)
	}

	if err := utils.UnzipSource(master.currentJob.optionalFilesZip.Name, ""); err != nil {
		return fmt.Errorf("error while unzipping optional files zip %+v", err)
	}

	//now, need to run distribute
	data, err := common.ExecuteProcess(logger.MASTER, utils.DistributeBinary,
		utils.File{Name: "distribute.txt", Content: []byte(master.currentJob.jobContent)},
		master.currentJob.distributeBinary)
	if err != nil {
		return err
	}

	var tasks *[]string
	err = gob.NewDecoder(bytes.NewReader(data)).Decode(tasks)
	if err != nil {
		return fmt.Errorf("error while decoding the tasks array created by the distribute binary")
	}
	logger.LogInfo(logger.MASTER, logger.DEBUGGING, "These are the tasks for the job %+v\n: %+v", master.currentJob.jobId, tasks)

	//now that I have the tasks, set the appropriate fields in the master
	master.currentJob.tasks = make([]Task, len(*tasks))
	master.currentJob.finishedTasksFilePaths = make([]string, len(*tasks))
	master.currentJob.workersTimers = make([]WorkerAndHisTimer, len(*tasks))

	for i, task := range *tasks {
		master.currentJob.tasks[i] = Task{
			id:      uuid.NewString(),
			content: task,
			isDone:  false,
		}
	}
	return nil
}

// this function expects to hold a lock
func (master *Master) addDumbJob() {
	// process, _ := os.ReadFile(PROCESS_EXE + ".binary")
	// distribute, _ := os.ReadFile(DISTRIBUTE_EXE + ".binary")
	// aggregate, _ := os.ReadFile(AGGREGATE_EXE + ".binary")

	// reply := RPC.GetJobReply{
	// 	IsAccepted: true,
	// 	JobId:      "#1",
	// 	ClientId:   "clientId",
	// 	JobContent: "jobContent.txt",
	// 	ProcessBinary: utils.File{
	// 		Name:    PROCESS_EXE + ".binary",
	// 		Content: process,
	// 	},
	// 	DistributeBinary: utils.File{
	// 		Name:    DISTRIBUTE_EXE + ".binary",
	// 		Content: distribute,
	// 	},
	// 	AggregateBinary: utils.File{
	// 		Name:    AGGREGATE_EXE + ".binary",
	// 		Content: aggregate,
	// 	},
	// }
	// master.setJobStatus(&reply)
}

// start a thread that periodically sends my progress
func (master *Master) sendPeriodicProgress() {
	for {
		time.Sleep(time.Second * 1)

		master.mu.Lock()
		if !master.isRunning {
			master.mu.Unlock()
			continue
		}

		var progress float32 = 0
		for _, t := range master.currentJob.tasks {
			if t.isDone {
				progress++
			}
		}
		progress /= float32(len(master.currentJob.tasks))

		args := &RPC.CurrentJobProgressArgs{
			MasterId: master.id,
			JobId:    master.currentJob.jobId,
			ClientId: master.currentJob.clientId,
			Progress: progress,
			Status:   RPC.PROCESSING, //todo this will probably change in the future
		}

		master.mu.Unlock()
		reply := &RPC.CurrentJobProgressReply{}
		RPC.EstablishRpcConnection(&RPC.RpcConnection{
			Name:         "LockServer.HandleGetJob", //todo ask rawan for the name
			Args:         &args,
			Reply:        &reply,
			SenderLogger: logger.MASTER,
			Reciever: RPC.Reciever{
				Name: "Lockserver",
				Port: LockServerPort,
				Host: LockServerHost,
			},
		})

	}
}

// start a thread that waits on a job from the message queue
func (master *Master) qConsumer() {
	ch, err := master.q.Dequeue(mq.ASSIGNED_JOBS_QUEUE)
	time.Sleep(10 * time.Second) //sleep for 10 seconds to await lockServer waking up

	if err != nil {
		logger.FailOnError(logger.MASTER, logger.ESSENTIAL, "Master can't consume jobs with this error %v", err)
	}

	for {
		master.mu.Lock()

		if master.isRunning { //there is a current job, so dont try to pull a new one
			master.mu.Unlock()
			time.Sleep(time.Second * 5)
			continue
		} else {
			master.mu.Unlock()
		}

		//no lock
		select {
		case newJob := <-ch: //and I am available to get one
			//new job arrived
			body := newJob.Body
			data := &mq.AssignedJob{}

			err := json.Unmarshal(body, data)
			if err != nil {
				logger.LogError(logger.MASTER, logger.ESSENTIAL, "Unable to consume job with error %v\nWill discard it", err)
				newJob.Ack(false) //probably should just ack so it doesnt sit around in the queue forever

				//send err to the mq
				master.mu.Lock()
				master.publishErrAsFinJob(fmt.Sprintf("unable to martial received job %+v with err %+v", string(body), err))
				master.mu.Unlock()
				continue
			}

			//ask lockserver if i can get it
			//todo fill the mq properly
			args := &RPC.GetJobArgs{
				JobId:            data.JobId,
				ClientId:         data.ClientId,
				MasterId:         master.id,
				JobContent:       data.JobContent,
				MQJobFound:       true,
				ProcessBinary:    data.ProcessBinary,
				DistributeBinary: data.DistributeBinary,
				AggregateBinary:  data.AggregateBinary,
			}
			reply := &RPC.GetJobReply{}
			ok, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
				Name:         "LockServer.HandleGetJob",
				Args:         &args,
				Reply:        &reply,
				SenderLogger: logger.MASTER,
				Reciever: RPC.Reciever{
					Name: "Lockserver",
					Port: LockServerPort,
					Host: LockServerHost,
				},
			})
			if !ok {
				logger.LogError(logger.MASTER, logger.ESSENTIAL, "Unable to contact lockserver to ask about job with error %v\nWill discard it", err)
				newJob.Nack(false, true) //requeue job, so maybe another master can contact the lock server
				continue
			}

			if reply.IsAccepted {
				//use args that the lockserver has accepted
				logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer accepted job request %v for client %+v", args.JobId, args.ClientId)
				newJob.Ack(false)

				continue
			} else {
				//use alternative provided by lockserver
				logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer provided outstanding job %v for client %v instead of requested job %v", reply.JobId, reply.ClientId, args.JobId)
				newJob.Nack(false, true) //requeue job since lockserver provided another
			}

			if err := master.setJobStatus(reply); err != nil {
				logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while setting job status: %+v", err)
			}

		default: //I didn't find a job from the message queue

			//need to ask lockServer if there are any outstanding jobs
			reply := &RPC.GetJobReply{}
			args := &RPC.GetJobArgs{
				MasterId:   master.id,
				MQJobFound: false,
			}
			ok, _ := RPC.EstablishRpcConnection(&RPC.RpcConnection{
				Name:         "LockServer.HandleGetJob",
				Args:         args,
				Reply:        &reply,
				SenderLogger: logger.MASTER,
				Reciever: RPC.Reciever{
					Name: "Lockserver",
					Port: LockServerPort,
					Host: LockServerHost,
				},
			})
			if ok && reply.IsAccepted { //todo: make sure rawan and salma implemented it correctly
				//there is indeed an outstanding job
				logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer provided outstanding job %v", reply.JobId)
				if err := master.setJobStatus(reply); err != nil {
					logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while setting job status: %+v", err)
				}
				continue
			}

			logger.LogInfo(logger.MASTER, logger.DEBUGGING, "No jobs found, about to sleep")
			time.Sleep(time.Second * 5)
		}

	}

}

//
// publishes a finished job to the message queue
// this  doesn't have to hold a lock
func (master *Master) publishFinJob(finJob mq.FinishedJob) {

	res, err := json.Marshal(finJob)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Unable to convert finished job to string! Discarding...")
		master.publishErrAsFinJob("Unable to marshal finished job")
	} else {
		err = master.q.Enqueue(mq.FINISHED_JOBS_QUEUE, res)
		if err != nil {
			logger.LogError(logger.MASTER, logger.ESSENTIAL, "Finished job not published to queue with err %v", err)
		} else {
			logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Finished job %+v successfully published to finished jobs queue for client %+v", finJob.JobId, finJob.ClientId)
		}
	}

}

//this function expects to hold a lock because it calls master.resetStatus
func (master *Master) publishErrAsFinJob(err string) {
	fn := mq.FinishedJob{}
	fn.Err = true
	fn.ErrMsg = err
	master.publishFinJob(fn)
	master.resetStatus()
}

//
//RPC handlers
//

func (master *Master) HandleGetTasks(args *RPC.GetTaskArgs, reply *RPC.GetTaskReply) error {
	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Worker called HandleGetTasks with these args %+v", args)

	master.mu.Lock()
	defer master.mu.Unlock()
	defer logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Master replied with this reply: %+v", reply)

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
			reply.TaskId = currentTask.id
			reply.JobId = master.currentJob.jobId
			reply.ProcessBinary = master.currentJob.processBinary

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
		master.publishErrAsFinJob(fmt.Sprintf("Worker sent this error: %+v", args.ErrMsg))
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
		master.publishErrAsFinJob(fmt.Sprintf("Error while saving worker's task locally on the master: %+v", err))
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
	master.finishUpJob()
	return nil
}

//this function expects to hold a lock
//it runs aggregate binary and pushes the job
//to the finJobs queue and notifies the lockserver
func (master *Master) finishUpJob() {
	//todo: aggregate.txt contains the paths of the finished tasks, each path in a newline

	finishedTasks := strings.Join(master.currentJob.finishedTasksFilePaths, "\n")

	//now, need to run aggregate

	finalResult, err := common.ExecuteProcess(logger.MASTER, utils.AggregateBinary,
		utils.File{Name: "aggregate.txt", Content: []byte(finishedTasks)},
		master.currentJob.aggregateBinary)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while running aggregate process: %+v", err)
		master.publishErrAsFinJob(fmt.Sprintf("Error while running aggregate process: %+v", err))
		return
	}

	logger.LogMilestone(logger.MASTER, logger.ESSENTIAL, "Finished job %+v for client %+v with result %+v",
		master.currentJob.jobId, master.currentJob.clientId, string(finalResult))

	master.publishFinJob(mq.FinishedJob{
		ClientId: master.currentJob.clientId,
		JobId:    master.currentJob.jobId,
		Content:  master.currentJob.jobContent,
		Result:   string(finalResult),
	})

	master.attemptSendFinishedJobToLockServer()

	master.resetStatus()
}

//this function expects to hold a lock
func (master *Master) attemptSendFinishedJobToLockServer() {
	ok := false
	ctr := 1
	mxRetries := 3

	for !ok && ctr < mxRetries {

		if !master.isRunning {
			break
		}

		args := &RPC.FinishedJobArgs{
			MasterId: master.id,
			JobId:    master.currentJob.jobId,
			ClientId: master.currentJob.clientId,
		}
		reply := &RPC.FinishedJobReply{}
		ok, _ := RPC.EstablishRpcConnection(&RPC.RpcConnection{
			Name:         "LockServer.HandleFinishedJob",
			Args:         &args,
			Reply:        &reply,
			SenderLogger: logger.MASTER,
			Reciever: RPC.Reciever{
				Name: "Lockserver",
				Port: LockServerPort,
				Host: LockServerHost,
			},
		})

		if !ok {
			logger.LogError(logger.MASTER, logger.ESSENTIAL, "Attempt number %v to send finished job to lockServer unsuccessfull", ctr)
		} else {
			logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Attempt number %v to send finished job to lockServer successfull", ctr)
			return
		}
		ctr++
		time.Sleep(10 * time.Second)
	}
	return
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

// main server loop
func (master *Master) server() error {
	rpc.Register(master)
	rpc.HandleHTTP()

	addrToListen := CreateMasterAddress()

	os.Remove(addrToListen)
	listener, err := net.Listen("tcp", addrToListen)

	if err != nil {
		logger.FailOnError(logger.MASTER, logger.ESSENTIAL, "Error while listening on socket: %v", err)
	} else {
		logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Listening on socket: %v", addrToListen)
	}

	go http.Serve(listener, nil)
	return nil
}

//
//helpers
//

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
		return false //no tasks in the first plac
	}

	for _, t := range master.currentJob.tasks {
		if !t.isDone {
			return false
		}
	}
	return true
}

// func (master *Master) removeSliceElementByIndex (arr *[]Task, index int) int {

// 	// Shift a[i+1:] left one index.
// 	copy((*arr)[index:], (*arr)[index+1:])
// 	// Erase last element (write zero value).
// 	(*arr)[len((*arr))-1] = Task{}
// 	// Truncate slice.
// 	(*arr) = (*arr)[:len((*arr))-1]
// }
