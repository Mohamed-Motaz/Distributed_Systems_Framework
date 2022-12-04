package RPC

import (
	logger "Framework/Logger"
	utils "Framework/Utils"
	"net/rpc"
	"time"
)

/*
This package contains all the RPC definitions
for any inter-servers communication
*/
func EstablishRpcConnection(rpcConn *RpcConnection) (bool, error) {
	successfullConnection := false
	var client *rpc.Client
	var err error

	for i := 1; i <= 3 && !successfullConnection; i++ {
		client, err = rpc.DialHTTP("tcp", rpcConn.Reciever.Host+":"+rpcConn.Reciever.Port)
		if err != nil {
			logger.LogError(
				rpcConn.SenderLogger,
				logger.ESSENTIAL,
				"Attempt number %v of dialing %v failed with error: %v\n",
				i, rpcConn.Reciever.Name, err,
			)
			time.Sleep(10 * time.Second)
		} else {
			successfullConnection = true
		}
	}

	if !successfullConnection {
		logger.FailOnError(
			rpcConn.SenderLogger,
			logger.ESSENTIAL,
			"Error dialing http: %v\nFatal Error: Can't establish connection with %v. Exiting now",
			rpcConn.Reciever.Name, err,
		)
	}

	defer client.Close()

	err = client.Call(rpcConn.Name, rpcConn.Args, rpcConn.Reply)

	if err != nil {
		logger.LogError(
			rpcConn.SenderLogger,
			logger.ESSENTIAL,
			"Unable to call %v with RPC with error: %v",
			rpcConn.Reciever.Name, err,
		)
		return false, err
	}

	return true, nil
}

type Reciever struct {
	Name string
	Port string
	Host string
}

type RpcConnection struct {
	Name         string
	Args         interface{}
	Reply        interface{}
	SenderLogger int
	Reciever     Reciever
}

// master-worker communication ---------
type GetTaskArgs struct {
	WorkerId string
}
type GetTaskReply struct {
	TaskAvailable bool
	TaskContent   string
	ProcessFolder utils.Folder
	OptionalFiles []utils.File
	TaskId        string
	JobId         string
}

type FinishedTaskArgs struct {
	TaskId     string
	JobId      string
	TaskResult string
	IsSuccess  bool
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
	JobId                string
	ClientId             string
	MasterId             string
	JobContent           string
	MQJobFound           bool
	ProcessFolderName    string
	DistributeFolderName string
	AggregateFolderName  string
	OptionalFilesNames   []string
}

type GetJobReply struct {
	IsAccepted       bool //lock server will answer whether it accepted my job request
	JobId            string
	ClientId         string
	JobContent       string
	ProcessFolder    utils.Folder
	DistributeFolder utils.Folder
	AggregateFolder  utils.Folder
	OptionalFiles    []utils.File
}

type FinishedJobArgs struct {
	JobId    string
	ClientId string
	MasterId string
}

type FinishedJobReply struct {
	Error utils.Error
}

//websocketserver - lockserver communication --------------

type ExeFolderUploadArgs struct {
	FolderType utils.FolderType
	Folder     utils.Folder
}

type OptionalFilesUploadArgs struct {
	JobId string
	Files []utils.File
}

type FileUploadReply struct {
	Error utils.Error
}

type GetExeFoldersReply struct {
	ProcessFolderNames    []string
	DistributeFolderNames []string
	AggregateFolderNames  []string
	Error                 utils.Error
}

type DeleteExeFolderArgs struct {
	FolderType utils.FolderType
	FolderName string
}

type DeleteExeFolderReply struct {
	Error utils.Error
}
