package RPC

import (
	logger "Framework/Logger"
	utils "Framework/Utils"
	"net/rpc"
	"time"
)

/*
This package contains all the RPC definitions
for any inter-server communication
*/

// master-worker communication ---------
type GetTaskArgs struct {
	WorkerId        string
	ProcessBinaryId int
	JobId           string
}
type GetTaskReply struct {
	TaskAvailable    bool
	TaskContent      string
	ProcessBinary    utils.RunnableFile
	OptionalFilesZip utils.File
	TaskId           string
	JobId            string
}

type FinishedTaskArgs struct {
	TaskId     string
	JobId      string
	TaskResult string
	utils.Error
}
type FinishedTaskReply struct {
}

type WorkerHeartBeatArgs struct {
	WorkerId string
	TaskId   string
	JobId    string
}
type WorkerHeartBeatReply struct {
}

//master-lockserver communication ---------

type GetJobArgs struct {
	JobId              string
	ClientId           string
	MasterId           string
	JobContent         string
	MQJobFound         bool
	ProcessBinaryId    string //required only when the lock server adds this job to the db
	DistributeBinaryId string //required only when the lock server adds this job to the db
	AggregateBinaryId  string //required only when the lock server adds this job to the db
	CreatedAt          time.Time
}

type GetJobReply struct {
	IsAccepted       bool //lock server will answer whether it accepted my job request
	JobId            string
	ClientId         string
	JobContent       string
	ProcessBinary    utils.RunnableFile
	DistributeBinary utils.RunnableFile
	AggregateBinary  utils.RunnableFile
	OptionalFilesZip utils.File
	CreatedAt        time.Time
}

type FinishedJobArgs struct {
	JobId    string
	MasterId string
	ClientId string
}

type FinishedJobReply struct {
}

//internal map for the lock server
// {masterId, value is this struct}
//flow -- Master continuously (every x sec) calls the lockServer (if the master has a current job)
//ws server will continuously call the lockserver to ask for results -- [of the structs]

type JobProgress string

const (
	DISTRIBUTING JobProgress = "Distributing"
	PROCESSING   JobProgress = "Processing"
	AGGREGATING  JobProgress = "Aggregating"
	UNRESPONSIVE JobProgress = "Unresponsive"
	FREE         JobProgress = "Free"
)

type WorkerTask struct {
	WorkerId             string
	CurrentTaskContent   string //may be empty
	FinishedTasksContent []string
}

// DONE: add the current binary names used ex processbinaryname ta7t
type CurrentJobProgress struct {
	MasterId             string
	JobId                string
	ClientId             string
	DistributeBinaryName string
	ProcessBinaryName    string
	AggregateBinaryName  string
	Progress             float32
	Status               JobProgress
	WorkersTasks         []WorkerTask
	CreatedAt            time.Time
	TimeAssigned         time.Time
}
type SetJobProgressArgs struct {
	CurrentJobProgress
}

type SetJobProgressReply struct {
}

//websocketserver - lockserver communication --------------

type BinaryUploadArgs struct {
	FileType utils.FileType
	File     utils.RunnableFile
}

type OptionalFilesUploadArgs struct {
	JobId    string
	FilesZip utils.File
}

type FileUploadReply struct {
	utils.Error
}

type GetBinaryFilesArgs struct {
}

type GetBinaryFilesReply struct {
	ProcessBinaryNames    []BinaryFileNameAndId
	DistributeBinaryNames []BinaryFileNameAndId
	AggregateBinaryNames  []BinaryFileNameAndId
	utils.Error
}

type BinaryFileNameAndId struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type DeleteBinaryFileArgs struct {
	FileType utils.FileType
	FileName string
}

type DeleteBinaryFileReply struct {
	utils.Error
}

type GetSystemProgressArgs struct {
}

type GetSystemProgressReply struct {
	Progress []CurrentJobProgress
	utils.Error
}
type DeleteOptionalFilesArgs struct {
	JobId string
}

type DeleteOptionalFilesReply struct {
	utils.Error
}

// actual helper functions ----------------------------------------------------------------------------------------------------------------------------------------
func EstablishRpcConnection(rpcConn *RpcConnection) (bool, error) {
	successfullConnection := false
	var client *rpc.Client
	var err error

	for i := 1; i <= 3 && !successfullConnection; i++ {
		client, err = rpc.DialHTTP("tcp", rpcConn.Receiver.Host+":"+rpcConn.Receiver.Port)
		if err != nil {
			logger.LogError(
				rpcConn.SenderLogger,
				logger.ESSENTIAL,
				"Attempt number %v of dialing %v failed with error: %v\n",
				i, rpcConn.Receiver.Name, err,
			)
			time.Sleep(5 * time.Second)
		} else {
			successfullConnection = true
		}
	}

	if !successfullConnection {
		logger.FailOnError(
			rpcConn.SenderLogger,
			logger.ESSENTIAL,
			"Error dialing http: %v\nFatal Error: Can't establish connection with %v. Exiting now",
			rpcConn.Receiver.Name, err,
		)
	}

	defer client.Close()

	err = client.Call(rpcConn.Name, rpcConn.Args, rpcConn.Reply)

	if err != nil {
		logger.LogError(
			rpcConn.SenderLogger,
			logger.ESSENTIAL,
			"Unable to call %v with RPC with error: %v",
			rpcConn.Receiver.Name, err,
		)
		return false, err
	}

	return true, nil
}

type Receiver struct {
	Name string
	Port string
	Host string
}

type RpcConnection struct {
	Name         string
	Args         interface{}
	Reply        interface{}
	SenderLogger int
	Receiver     Receiver
}
