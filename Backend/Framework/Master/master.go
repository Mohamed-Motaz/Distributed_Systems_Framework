package main

import (
	logger "Framework/Logger"
	mq "Framework/MessageQueue"
	utils "Framework/Utils"
	"fmt"
	"math"
	"strings"

	"Framework/RPC"
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
	master.resetCurrentJob() //remove old files from previous jobs

	go master.rpcServer()
	go master.consumeJob()
	go master.sendPeriodicProgress()

	logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Master is now alive")

	return master
}

// this function expects to hold a lock
// it resets the currentJob of the master
func (master *Master) resetCurrentJob() {

	utils.KeepFilesThatMatch(FileNamesToIgnore)

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



// start a thread that waits on a job from the message queue
func (master *Master) consumeJob() {

	jobsChan, err := master.q.Dequeue(mq.ASSIGNED_JOBS_QUEUE)
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

		reply := &RPC.GetJobReply{}
		//no lock
		select {
		case newJob := <-jobsChan: //and I am available to get one
			//new job arrived
			body := newJob.Body
			data := &mq.AssignedJob{}

			err := json.Unmarshal(body, data)
			if err != nil {
				logger.LogError(logger.MASTER, logger.ESSENTIAL, "Unable to consume job with error %v\nWill discard it", err)
				newJob.Ack(false)

				continue
			}

			//ask lockserver if i can get it
			args := &RPC.GetJobArgs{
				JobId:              data.JobId,
				ClientId:           data.ClientId,
				MasterId:           master.id,
				JobContent:         data.JobContent,
				MQJobFound:         true,
				ProcessBinaryId:    data.ProcessBinaryId,
				DistributeBinaryId: data.DistributeBinaryId,
				AggregateBinaryId:  data.AggregateBinaryId,
				CreatedAt:          data.CreatedAt,
			}

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
				time.Sleep(1 * time.Minute) //sleep so maybe if the lockserver is asleep now, he can would have woken up by then
				newJob.Nack(false, true)    //requeue job, so maybe another master can contact the lock server
				continue
			}

			if reply.IsAccepted {
				//use args that the lockserver has accepted
				logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer accepted job request %v for client %+v", args.JobId, args.ClientId)
				newJob.Ack(false)
			} else {

				if reply.JobId != "" {
					//use alternative provided by lockserver
					logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer provided outstanding job %v for client %v instead of requested job %v", reply.JobId, reply.ClientId, args.JobId)
					newJob.Nack(false, true) //requeue job since lockserver provided another
				} else {
					//seems like there is an error with the lockserver, and I can't accept any jobs now
					logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "LockServer seems like it is unable to accept my job requests now")
					newJob.Nack(false, true)
					continue
				}
			}

		default: //I didn't find a job from the message queue

			args := &RPC.GetJobArgs{
				MasterId:   master.id,
				MQJobFound: false,
				JobId:      "",
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

			} else {
				logger.LogInfo(logger.MASTER, logger.DEBUGGING, "No jobs found, about to sleep")
				time.Sleep(time.Second * 5)
				continue
			}
		}

		master.mu.Lock()
		master.startWorkingOnJob(reply)
		master.mu.Unlock()

	}
}

func (master *Master) startWorkingOnJob(reply *RPC.GetJobReply) {

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
		optionalFilesZip:       reply.OptionalFilesZip,
		createdAt:              reply.CreatedAt,
		timeAssigned:           time.Now(),
	}

	if err := master.writeBinariesAndFilesOnDisk(); err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while setting job status: %+v", err)
		//DONE: send the lockserver an error alerting him that I finished this job (lie)
		master.attemptSendFinishedJobToLockServer()
		master.publishErrAsFinishedJob(err.Error(), master.currentJob.clientId, master.currentJob.jobId)
		return
	}

	if err := master.distributeJob(reply); err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while setting job status: %+v", err)
		//DONE: send the lockserver an error alerting him that I finished this job (lie)
		master.attemptSendFinishedJobToLockServer()
		master.publishErrAsFinishedJob(err.Error(), master.currentJob.clientId, master.currentJob.jobId)
	}
	
}


func (master *Master) distributeJob(reply *RPC.GetJobReply) error {

	//now, need to run distribute
	data, err := utils.ExecuteProcess(logger.MASTER, utils.DistributeBinary,
		utils.File{Name: "distribute.txt", Content: []byte(master.currentJob.jobContent)},
		master.currentJob.distributeBinary)
	if err != nil {
		return err
	}

	tasks := strings.Split(string(data), ",") //TODO: expect the tasks to be comma-separated

	if err != nil {
		return fmt.Errorf("error while decoding the tasks array created by the distribute binary")
	}
	logger.LogInfo(logger.MASTER, logger.DEBUGGING, "These are the tasks for the job %+v\n: %+v", master.currentJob.jobId, tasks)

	//now that I have the tasks, set the appropriate fields in the master
	master.currentJob.tasks = make([]Task, len(tasks))
	master.currentJob.finishedTasksFilePaths = make([]string, len(tasks))
	master.currentJob.workersTimers = make([]WorkerAndHisTimer, len(tasks))

	for i, task := range tasks {
		master.currentJob.tasks[i] = Task{
			id:      uuid.NewString(),
			content: task,
			isDone:  false,
		}

		master.currentJob.workersTimers[i] = WorkerAndHisTimer{
			lastHeartBeat: time.Time{},
		}
	}
	return nil
}

// this function expects to hold a lock because it calls publishErrAsFinishedJob & publishFinishedJob
// it runs aggregate binary and pushes the job
// to the finishedJobs queue and notifies the lockserver
func (master *Master) aggregateTasks() {
	//aggregate.txt contains the paths of the finished tasks, each path in a newline
	finishedTasks := strings.Join(master.currentJob.finishedTasksFilePaths, ",")

	//now, need to run aggregate
	finalResult, err := utils.ExecuteProcess(logger.MASTER, utils.AggregateBinary,
		utils.File{Name: "aggregate.txt", Content: []byte(finishedTasks)},
		master.currentJob.aggregateBinary)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Error while running aggregate process: %+v", err)
		master.publishErrAsFinishedJob(fmt.Sprintf("Error while running aggregate process: %+v", err), master.currentJob.clientId, master.currentJob.jobId)
		return
	}

	logger.LogMilestone(logger.MASTER, logger.ESSENTIAL, "Finished job %+v for client %+v with result %+v",
		master.currentJob.jobId, master.currentJob.clientId, string(finalResult))

	master.publishFinishedJob(mq.FinishedJob{
		ClientId:     master.currentJob.clientId,
		JobId:        master.currentJob.jobId,
		Content:      master.currentJob.jobContent,
		Result:       string(finalResult),
		CreatedAt:    master.currentJob.createdAt,
		TimeAssigned: master.currentJob.timeAssigned,
	}, false)

	master.attemptSendFinishedJobToLockServer()

	master.resetCurrentJob()
}

// publishes a finished job to the message queue
// this has to hold a lock
func (master *Master) publishFinishedJob(finishedJob mq.FinishedJob, resetCurrentJob bool) {

	res, err := json.Marshal(finishedJob)
	if err != nil {
		logger.LogError(logger.MASTER, logger.ESSENTIAL, "Unable to convert finished job to string! Discarding...")
		fn := mq.FinishedJob{
			Error: utils.Error{
				Err:    true,
				ErrMsg: "Unable to marshal finished job",
			},
		}
		res, _ = json.Marshal(fn)
		err = master.q.Enqueue(mq.FINISHED_JOBS_QUEUE, res)
		if err != nil {
			logger.LogError(logger.MASTER, logger.ESSENTIAL, "Finished job not published to queue with err %v", err)
		} else {
			logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Finished job %+v successfully published to finished jobs queue for client %+v", finishedJob.JobId, finishedJob.ClientId)
		}
	} else {
		err = master.q.Enqueue(mq.FINISHED_JOBS_QUEUE, res)
		if err != nil {
			logger.LogError(logger.MASTER, logger.ESSENTIAL, "Finished job not published to queue with err %v", err)
		} else {
			logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Finished job %+v successfully published to finished jobs queue for client %+v", finishedJob.JobId, finishedJob.ClientId)
		}
	}
	if resetCurrentJob {
		master.resetCurrentJob()
	}
}

// start a thread that periodically sends my progress
func (master *Master) sendPeriodicProgress() {
	for {
		time.Sleep(time.Second * 1)
		master.mu.Lock()
		if !master.isRunning {
			args := &RPC.SetJobProgressArgs{
				CurrentJobProgress: RPC.CurrentJobProgress{
					MasterId: master.id,
					JobId:    "",
					ClientId: "",
					Progress: 0,
					Status:   RPC.FREE,
				},
			}
			master.mu.Unlock()
			logger.LogInfo(logger.MASTER, logger.DEBUGGING, "About to send periodic progress %+v", args)
			reply := &RPC.SetJobProgressReply{}
			if _, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
				Name:         "LockServer.HandleSetJobProgress",
				Args:         &args,
				Reply:        &reply,
				SenderLogger: logger.MASTER,
				Reciever: RPC.Reciever{
					Name: "Lockserver",
					Port: LockServerPort,
					Host: LockServerHost,
				},
			}); err != nil {
				logger.LogError(logger.MASTER, logger.DEBUGGING, "Error while sending periodic progress err: %+v", err)
			}
			continue
		}
		var progress float32 = 0
		for _, t := range master.currentJob.tasks {
			if t.isDone {
				progress++
			}
		}
		progress /= float32(len(master.currentJob.tasks))
		progress *= 100
		progress = float32(math.Round((float64(progress*100) / 100)))
		//DONE: set progress to 1 decimal point
		args := &RPC.SetJobProgressArgs{
			CurrentJobProgress: RPC.CurrentJobProgress{
				MasterId:             master.id,
				JobId:                master.currentJob.jobId,
				ClientId:             master.currentJob.clientId,
				Progress:             progress,
				Status:               RPC.PROCESSING, //todo this will probably change in the future
				CreatedAt:            master.currentJob.createdAt,
				TimeAssigned:         master.currentJob.timeAssigned,
				WorkersTasks:         master.generateWorkersTasks(),
				DistributeBinaryName: master.currentJob.distributeBinary.Name,
				ProcessBinaryName:    master.currentJob.processBinary.Name,
				AggregateBinaryName:  master.currentJob.aggregateBinary.Name,
			},
		}
		master.mu.Unlock()
		logger.LogInfo(logger.MASTER, logger.DEBUGGING, "About to send periodic progress %+v", args)
		reply := &RPC.SetJobProgressReply{}
		if _, err := RPC.EstablishRpcConnection(&RPC.RpcConnection{
			Name:         "LockServer.HandleSetJobProgress",
			Args:         &args,
			Reply:        &reply,
			SenderLogger: logger.MASTER,
			Reciever: RPC.Reciever{
				Name: "Lockserver",
				Port: LockServerPort,
				Host: LockServerHost,
			},
		}); err != nil {
			logger.LogError(logger.MASTER, logger.DEBUGGING, "Error while sending periodic progress err: %+v", err)
		}
	}
}

// this function expects to hold a lock
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

// main server loop
func (master *Master) rpcServer() error {

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
