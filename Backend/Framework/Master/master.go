package main

import (
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
	go master.qConsumer()

	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Master is now alive")

	return master
}

func CreateMasterAddress() string {
	return MyHost + ":" + MyPort
}

// this function expects to hold a lock
//it resets the currentJob of the master
func (master *Master) resetStatus() {
	master.currentJob = CurrentJob{
		tasks:         make([]Task, 0),
		finishedTasks: make([]string, 0),
		workersTimers: make([]WorkerAndHisTimer, 0),
		processExe: Exe{
			exe: make([]byte, 0),
		},
		distributeExe: Exe{
			exe: make([]byte, 0),
		},
		aggregateExe: Exe{
			exe: make([]byte, 0),
		},
	}
	master.isRunning = false

	//todo support other oss
	utils.RemoveFilesWithStartName(PROCESS_EXE)
	utils.RemoveFilesWithStartName(DISTRIBUTE_EXE)
	utils.RemoveFilesWithStartName(AGGREGATE_EXE)

}

// this function expects to hold a lock
//this function is responsible for setting up the new job and running distribute
func (master *Master) setJobStatus(jobId string, jobContent string, clientId string,
	processExe []byte, processExeName string,
	distributeExe []byte, distributeExeName string,
	aggregateExe []byte, aggregateExeName string) error {

	master.currentJob = CurrentJob{
		clientId:   clientId,
		jobContent: jobContent,
		jobId:      jobId,

		tasks:         make([]Task, 0),
		finishedTasks: make([]string, 0),
		workersTimers: make([]WorkerAndHisTimer, 0),
		processExe: Exe{
			exe:  processExe,
			name: PROCESS_EXE + processExeName,
		},
		distributeExe: Exe{
			exe:  distributeExe,
			name: DISTRIBUTE_EXE + distributeExeName,
		},
		aggregateExe: Exe{
			exe:  aggregateExe,
			name: AGGREGATE_EXE + aggregateExeName,
		},
	}
	master.isRunning = true

	//todo, think about supporting different os exes
	//todo any errors here should be propagated to client

	err := utils.CreateAndWriteToFile(master.currentJob.processExe.name, master.currentJob.processExe.exe)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while creating the process exe file locally on the master %+v", err)
		return fmt.Errorf("Error while creating the process exe file locally on the master")
	}
	err = utils.CreateAndWriteToFile(master.currentJob.distributeExe.name, master.currentJob.distributeExe.exe)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while creating the distribute exe file locally on the master %+v", err)
		return fmt.Errorf("Error while creating the distribute exe file locally on the master")
	}
	err = utils.CreateAndWriteToFile(master.currentJob.aggregateExe.name, master.currentJob.aggregateExe.exe)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while creating the aggregate exe file locally on the master %+v", err)
		return fmt.Errorf("Error while creating the aggregate exe file locally on the master")
	}

	//now, need to run distribute
	fName := "distributeTaskContents.txt"
	fPath := "./" + fName
	err = utils.CreateAndWriteToFile(fName, []byte(master.currentJob.jobContent))
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while creating the temporary file that contains the job contents for distribute process locally on the master %+v", err)
		return fmt.Errorf("Error while creating the temporary file that contains the job contents for distribute process locally on the master")
	}

	_, err = exec.Command("./" + master.currentJob.distributeExe.name + " " + fPath).Output()
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while executing distribute process %+v", err)
		return fmt.Errorf("Error while executing distribute process")
	}

	//now need to read from this file the resulting data
	data, err := os.ReadFile(fPath)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while reading from the distribute process %+v", err)
		return fmt.Errorf("Error while reading from the distribute process")
	}

	var tasks *[]string
	err = gob.NewDecoder(bytes.NewReader(data)).Decode(tasks)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while decoding the tasks array created by the distribute exe %+v", err)
		return fmt.Errorf("Error while decoding the tasks array created by the distribute exe")
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

	master.currentJob = CurrentJob{
		clientId:      "id",
		jobContent:    "Do Stupid Things",
		jobId:         "#1",
		tasks:         make([]Task, 2),
		finishedTasks: make([]string, 2),
		workersTimers: make([]WorkerAndHisTimer, 2),
		processExe: Exe{
			exe:  make([]byte, 0),
			name: PROCESS_EXE + "process.exe",
		},
		distributeExe: Exe{
			exe:  make([]byte, 0),
			name: DISTRIBUTE_EXE + "distribute.exe",
		},
		aggregateExe: Exe{
			exe:  make([]byte, 0),
			name: AGGREGATE_EXE + "aggregate.exe",
		},
	}
	master.isRunning = true
	master.currentJob.tasks[0] = Task{id: "#1 task", content: "hello for 1", isDone: false}
	master.currentJob.tasks[1] = Task{id: "#2 task", content: "hello for 2", isDone: false}
}

//
// start a thread that waits on a job from the message queue
//
func (master *Master) qConsumer() {
	ch, err := master.q.Dequeue(mq.ASSIGNED_JOBS_QUEUE)
	time.Sleep(10 * time.Second) //sleep for 10 seconds to await lockServer waking up

	if err != nil {
		logger.FailOnError(logger.MASTER, logger.ESSENTIAL, "Master can't consume jobs because with this error %v", err)
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
				JobContent: data.Content,
				MQJobFound: true,
			}
			reply := &RPC.GetJobReply{}
			ok := master.callLockServer("LockServer.HandleGetJob", args, reply)
			if !ok {
				logger.LogError(logger.MASTER, logger.ESSENTIAL, "Unable to contact lockserver to ask about job with error %v\nWill discard it", err)
				newJob.Nack(false, true)
				continue
			}

			if reply.IsAccepted {
				//use args that the lockserver has accepted
				logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer accepted job request %v for client %+v", args.JobId, args.ClientId)
				newJob.Ack(false)
				master.setJobStatus(args.JobId, args.JobContent, args.ClientId)
				continue
			} else {
				//use alternative provided by lockserver
				logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer provided outstanding job %v for client %v instead of requested job %v", reply.JobId, reply.ClientId, args.JobId)
				newJob.Nack(false, true)
				master.setJobStatus(reply.JobId, reply.JobContent, reply.ClientId)
				continue
			}

		default: //I didn't find a job from the message queue

			//need to ask lockServer if there are any outstanding jobs

			args := &RPC.GetJobArgs{
				MasterId:   master.id,
				MQJobFound: false,
			}
			reply := &RPC.GetJobReply{}
			ok := master.callLockServer("LockServer.HandleGetJob", args, reply)
			if ok && reply.IsAccepted {
				//there is indeed an outstanding job
				logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer provided outstanding job %v", reply.JobId)
				master.setJobStatus(reply.JobId, reply.JobContent, reply.ClientId)
				continue
			}

			logger.LogInfo(logger.MASTER, logger.DEBUGGING, "No jobs found, about to sleep")
			time.Sleep(time.Second * 5)
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

	master.currentJob.tasks[taskIndex].isDone = true
	master.currentJob.finishedTasks[taskIndex] = args.TaskResult
	master.currentJob.workersTimers[taskIndex].lastHeartBeat = time.Now()

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

//blocking
func (Master *Master) callLockServer(rpcName string, args interface{}, reply interface{}) bool {
	ctr := 1
	successfullConnection := false
	var client *rpc.Client
	var err error

	//attempt to conncect to master
	for ctr <= 3 && !successfullConnection {
		client, err = rpc.DialHTTP("tcp", LockServerHost+":"+LockServerPort) //blocking
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

// func (master *Master) removeSliceElementByIndex (arr *[]Task, index int) int {

// 	// Shift a[i+1:] left one index.
// 	copy((*arr)[index:], (*arr)[index+1:])
// 	// Erase last element (write zero value).
// 	(*arr)[len((*arr))-1] = Task{}
// 	// Truncate slice.
// 	(*arr) = (*arr)[:len((*arr))-1]
// }
