package main

import (
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	"Framework/RPC"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

//returns a pointer a master and runs it
func NewMaster() *Master {
	master := &Master{
		id:                uuid.NewString(), //random id
		q:                 mq.NewMQ(mq.CreateMQAddress(MqUsername, MqPassword, MqHost, MqPort)),
		maxHeartBeatTimer: 30 * time.Second, //each heartbeat should be every 10 seconds but we allaow up to 2 failures
		mu:                sync.Mutex{},
	}
	master.resetStatus()
	master.addDumbJob()

	go master.server()

	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Master is now alive")

	return master
}

func CreateMasterAddress() string {
	return MyHost + ":" + MyPort
}

// this function expects to hold a lock
func (master *Master) resetStatus() {
	master.currentJob = ""
	master.currentJobId = ""
	master.currentTasks = make([]Task, 0)
	master.finishedTasks = make([]string, 0)
	master.workersTimers = make([]WorkerAndHisTimer, 0)
	master.isRunning = false
}

// this function expects to hold a lock
func (master *Master) addDumbJob() {
	master.currentJob = "Do Stupid Things"
	master.currentJobId = "#1"
	master.currentTasks = make([]Task, 2)
	master.finishedTasks = make([]string, 2)
	master.workersTimers = make([]WorkerAndHisTimer, 2)
	master.isRunning = true

	master.currentTasks[0] = Task{id: "#1 task", content: "hello for 1", isDone: false}
	master.currentTasks[1] = Task{id: "#2 task", content: "hello for 2", isDone: false}
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

	for i := range master.currentTasks {
		currentTask := master.currentTasks[i]
		currentWorkerAndTimer := master.workersTimers[i]

		if currentTask.isDone {
			continue
		}
		if time.Since(currentWorkerAndTimer.lastHeartBeat) > master.maxHeartBeatTimer {
			//the other worker is probably dead, so give this worker this job
			reply.TaskAvailable = true
			reply.TaskContent = currentTask.content
			reply.TaskId = currentTask.id
			reply.JobId = master.currentJobId

			//now as a master, need to mark this job as given to a worker
			master.workersTimers[i] = WorkerAndHisTimer{
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

	if args.JobId != master.currentJobId {
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

	master.currentTasks[taskIndex].isDone = true
	master.finishedTasks[taskIndex] = args.TaskResult
	master.workersTimers[taskIndex].lastHeartBeat = time.Now()

	// TODO: create a thread that checks if all tasks are done and aggregates the results

	return nil
}

func (master *Master) HandleWorkerHeartBeats(args *RPC.WorkerHeartBeatArgs, reply *RPC.WorkerHeartBeatReply) error {
	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Worker called HandleWorkerHeartBeat with these args %+v", args)

	master.mu.Lock()
	defer master.mu.Unlock()

	if !master.isRunning {
		return nil
	}

	if args.JobId != master.currentJobId {
		return nil
	}

	taskIndex := master.getTaskIndexByTaskId(args.TaskId)
	if taskIndex == -1 {
		return nil
	}

	//now make sure that this worker was actually assigned this task
	if master.workersTimers[taskIndex].workerId != args.WorkerId {
		return nil
	}

	master.workersTimers[taskIndex].lastHeartBeat = time.Now()

	return nil
}

//
//main server loop
//
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

	for i := range master.currentTasks {
		if master.currentTasks[i].id == taskId {
			return i
		}
	}

	return -1
}

// func (master *Master) removeSliceElementByIndex (arr *[]Task, index int) int {

// 	// Shift a[i+1:] left one index.
// 	copy((*arr)[index:], (*arr)[index+1:])
// 	// Erase last element (write zero value).
// 	(*arr)[len((*arr))-1] = Task{}
// 	// Truncate slice.
// 	(*arr) = (*arr)[:len((*arr))-1]
// }
