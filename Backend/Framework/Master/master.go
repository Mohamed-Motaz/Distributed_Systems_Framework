package main

import (
	common "Framework/Common"
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	utils "Framework/Utils"
	"bytes"
	"fmt"
	"os/exec"

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
		maxHeartBeatTimer: 30 * time.Second, //each heartbeat should be every 10 seconds but we allaow up to 2 failures
		publishCh:         make(chan mq.FinishedJob),
		publishChAck:      make(chan bool),

		mu: sync.Mutex{},
	}
	master.resetStatus()
	master.addDumbJob()

	go master.server()
	go master.qConsumer()

	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Master is now alive")

	return master
}

func CreateMasterAddress() string {
	return MyHost + ":" + MyPort
}

func (master *Master) removeOldJobFiles() {
	for _, f := range master.currentJob.optionalFiles {
		os.Remove(f.Name)
	}

	utils.RemoveFilesWithStartName(PROCESS_EXE)
	utils.RemoveFilesWithStartName(DISTRIBUTE_EXE)
	utils.RemoveFilesWithStartName(AGGREGATE_EXE)
}

// this function expects to hold a lock
// it resets the currentJob of the master
func (master *Master) resetStatus() {
	//todo support other oss

	master.removeOldJobFiles()

	master.currentJob = CurrentJob{
		tasks:         make([]Task, 0),
		finishedTasks: make([]string, 0),
		workersTimers: make([]WorkerAndHisTimer, 0),
		processExe: utils.File{
			Content: make([]byte, 0),
		},
		distributeExe: utils.File{
			Content: make([]byte, 0),
		},
		aggregateExe: utils.File{
			Content: make([]byte, 0),
		},
		optionalFiles: make([]utils.File, 0),
	}
	master.isRunning = false

}

// this function expects to hold a lock
// this function is responsible for setting up the new job and running distribute
func (master *Master) setJobStatus(reply *RPC.GetJobReply) error {
	master.isRunning = true

	master.currentJob = CurrentJob{
		clientId:   reply.ClientId,
		jobContent: reply.JobContent,
		jobId:      reply.JobId,

		tasks:         make([]Task, 0),
		finishedTasks: make([]string, 0),
		workersTimers: make([]WorkerAndHisTimer, 0),
		processExe:    reply.ProcessExe,
		distributeExe: reply.DistributeExe,
		aggregateExe:  reply.AggregateExe,
		optionalFiles: reply.OptionalFiles,
	}
	//add the exe identitfiers next to their names
	master.currentJob.processExe.Name = PROCESS_EXE + master.currentJob.processExe.Name
	master.currentJob.distributeExe.Name = DISTRIBUTE_EXE + master.currentJob.distributeExe.Name
	master.currentJob.aggregateExe.Name = AGGREGATE_EXE + master.currentJob.aggregateExe.Name

	//todo, think about supporting different os exes
	//todo any errors here should be propagated to client

	exeLogs := []string{"process", "distribute", "aggregate"}
	exeNames := []string{master.currentJob.processExe.Name, master.currentJob.distributeExe.Name, master.currentJob.aggregateExe.Name}
	exeContents := [][]byte{master.currentJob.processExe.Content, master.currentJob.distributeExe.Content, master.currentJob.aggregateExe.Content}

	for i := 0; i < len(exeNames); i++ {
		err := utils.CreateAndWriteToFile(exeNames[i], exeContents[i])
		if err != nil {
			logger.LogError(logger.MASTER, logger.ESSENTIAL, "error while creating the %+v exe file locally on the master %+v", exeLogs[i], err)
			return fmt.Errorf("error while creating the %+v exe file locally on the master", exeLogs[i])
		}
	}

	for _, f := range master.currentJob.optionalFiles {
		err := utils.CreateAndWriteToFile(f.Name, f.Content)
		if err != nil {
			logger.LogError(logger.MASTER, logger.ESSENTIAL, "error while creating the file %+v locally on the master %+v", f.Name, err)
			return fmt.Errorf("error while creating the file %+v locally on the master", f.Name)
		}
	}

	//now, need to run distribute
	data, err := common.ExecuteProcess(logger.MASTER, utils.DistributeExe,
		utils.File{Name: "distribute.txt", Content: []byte(master.currentJob.jobContent)},
		master.currentJob.distributeExe)
	if err != nil {
		return err
	}

	var tasks *[]string
	err = gob.NewDecoder(bytes.NewReader(data)).Decode(tasks)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "error while decoding the tasks array created by the distribute exe %+v", err)
		return fmt.Errorf("error while decoding the tasks array created by the distribute exe")
	}

	//now that I have the tasks, set the appropriate fields in the master
	master.currentJob.tasks = make([]Task, len(*tasks))
	master.currentJob.finishedTasks = make([]string, len(*tasks))
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
	process, _ := os.ReadFile(PROCESS_EXE + ".exe")
	distribute, _ := os.ReadFile(DISTRIBUTE_EXE + ".exe")
	aggregate, _ := os.ReadFile(AGGREGATE_EXE + ".exe")

	reply := RPC.GetJobReply{
		IsAccepted: true,
		JobId:      "#1",
		ClientId:   "clientId",
		JobContent: "jobContent.txt",
		ProcessExe: utils.File{
			Name:    PROCESS_EXE + ".exe",
			Content: process,
		},
		DistributeExe: utils.File{
			Name:    DISTRIBUTE_EXE + ".exe",
			Content: distribute,
		},
		AggregateExe: utils.File{
			Name:    AGGREGATE_EXE + ".exe",
			Content: aggregate,
		},
	}
	master.setJobStatus(&reply)
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

				//todo: should probably send an error in the finished jobs queue
				continue
			}

			//ask lockserver if i can get it

			args := &RPC.GetJobArgs{
				JobId:      data.JobId,
				ClientId:   data.ClientId,
				MasterId:   master.id,
				JobContent: data.JobContent,
				MQJobFound: true,
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
				newJob.Nack(false, true)
				continue
			}

			if reply.IsAccepted {
				//use args that the lockserver has accepted
				logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer accepted job request %v for client %+v", args.JobId, args.ClientId)
				newJob.Ack(false)
				master.setJobStatus(reply)
				continue
			} else {
				//use alternative provided by lockserver
				logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer provided outstanding job %v for client %v instead of requested job %v", reply.JobId, reply.ClientId, args.JobId)
				newJob.Nack(false, true)
				master.setJobStatus(reply)
				continue
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
			if ok && reply.IsAccepted {
				//there is indeed an outstanding job
				logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer provided outstanding job %v", reply.JobId)
				master.setJobStatus(reply)
				continue
			}

			logger.LogInfo(logger.MASTER, logger.DEBUGGING, "No jobs found, about to sleep")
			time.Sleep(time.Second * 5)
		}

	}

}

//
// start a thread that listens for a finished job
// and then publishes it to the message queue
func (master *Master) qPublisher() {

	for {
		select {
		case finishedJob := <-master.publishCh:

			res, err := json.Marshal(finishedJob)
			if err != nil {
				logger.LogError(logger.MASTER, logger.ESSENTIAL, "Unable to convert finished job to string! Discarding...")
				//todo publish an error to the queue
			} else {
				err = master.q.Enqueue(mq.FINISHED_JOBS_QUEUE, res)
				if err != nil {
					logger.LogError(logger.MASTER, logger.ESSENTIAL, "Finished job not published to queue with err %v", err)
				} else {
					logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Finished job %+v successfully published to finished jobs queue for client %+v", fin.JobId, fin.ClientId)
				}
			}

			master.publishChAck <- err == nil

		default:
			time.Sleep(time.Second)
		}

	}

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
			reply.ProcessExe = master.currentJob.processExe

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

	if !args.IsSuccess {
		//TODO: need to handle this failure
		return nil
	}

	taskIndex := master.getTaskIndexByTaskId(args.TaskId)
	if taskIndex == -1 {
		return nil
	}

	//now need to write the results to a file, and save this files location
	filePath := "./" + args.TaskId + ".txt"
	err := utils.CreateAndWriteToFile(filePath, []byte(args.TaskResult))
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "error while creating the task file locally on the master %+v", err)
		return nil
		//todo how should i handle this error
	}

	master.currentJob.tasks[taskIndex].isDone = true
	master.currentJob.finishedTasks[taskIndex] = filePath
	master.currentJob.workersTimers[taskIndex].lastHeartBeat = time.Now()

	//check if all tasks are done and aggregate the results
	jobDone := master.allTasksDone()
	if !jobDone {
		return nil
	}

	//all tasks have been finished!
	var finishedTasksBytes *bytes.Buffer
	err = gob.NewEncoder(finishedTasksBytes).Encode(master.currentJob.finishedTasks)
	if err != nil {
		//TODO: send this in the mq somehow
		return nil
	}

	fPath := "./aggregate.txt"
	err = utils.CreateAndWriteToFile(fPath, finishedTasksBytes.Bytes())
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "error while creating the temporary file that contains the aggregate tasks locally on the master %+v", err)
		//TODO: send this in the mq somehow
	}

	_, err = exec.Command("./" + master.currentJob.aggregateExe.Name).Output()
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "error while executing aggregate process %+v", err)
		//TODO: send this in the mq somehow
	}

	//now need to read from this file the resulting data
	finalResult, err := os.ReadFile(fPath)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "error while reading from the aggregate process %+v", err)
		//TODO: send this in the mq somehow
	}

	logger.LogMilestone(logger.MASTER, logger.ESSENTIAL, "Finished job %+v for client %+v with result %+v",
		master.currentJob.jobId, master.currentJob.clientId, string(finalResult))

	//now need to push this to the mq
	master.publishCh <- mq.FinishedJob{
		ClientId: master.currentJob.clientId,
		JobId:    master.currentJob.jobId,
		Content:  master.currentJob.jobContent,
		Result:   string(finalResult),
	}

	//todo then need to send this to the lockserver
	success := <-master.publishChAck //block till Publisher sends the message to the queue or fails to do so

	if success {
		master.attemptSendFinishedJobToLockServer()
	}

	master.resetStatus()
	return nil
}

func (master *Master) attemptSendFinishedJobToLockServer() bool {
	ok := false
	ctr := 1
	mxRetries := 3

	for !ok && ctr < mxRetries {

		master.mu.Lock()
		if !master.isRunning {
			master.mu.Unlock()
			break
		}

		args := &RPC.FinishedJobArgs{
			MasterId: master.id,
			JobId:    master.jobId,
			ClientId: master.clientId,
			URL:      master.currentURL,
		}
		master.mu.Unlock()

		ok := master.callLockServer("LockServer.HandleFinishedJobs", args, &RPC.FinishedJobReply{}) //send data to lockServer

		if !ok {
			logger.LogError(logger.MASTER, logger.ESSENTIAL, "Attempt number %v to send finished job to lockServer unsuccessfull", ctr)
		} else {
			logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Attempt number %v to send finished job to lockServer successfull", ctr)
			return true
		}
		ctr++
		time.Sleep(10 * time.Second)
	}
	return ok
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
