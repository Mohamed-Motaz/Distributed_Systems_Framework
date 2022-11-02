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
		maxHeartBeatTimer: 10 * time.Second,
		mu:                sync.Mutex{},
	}
	master.resetStatus()

	go master.server()

	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Master is now alive")

	return master
}

//
// start a thread that listens for RPCs
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

func CreateMasterAddress() string {
	return MyHost + ":" + MyPort
}

func (master *Master) resetStatus() {
	master.currentJob = ""
	master.currentJobId = ""
	master.currentTasks = make([]Task, 0)
	master.workersTimers = make([]WorkerAndHisTimer, 0)
	master.isRunning = false
}

//
//RPC handlers
//

func (master *Master) HandleGetTasks(args *RPC.GetTaskArgs, reply *RPC.GetTaskReply) error {
	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Worker called HandleGetTasks with these args %+v", args)

	master.mu.Lock()
	defer master.mu.Unlock()

	if !master.isRunning {
		reply.TaskAvailable = false
		return nil
	}

	//now need to check the available tasks

	for i := range master.currentTasks {
		currentTask := master.currentTasks[i]
		currentWorkerAndTimer := master.workersTimers[i]

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
