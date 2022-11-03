package RPC

/*
This package contains all the RPC definitions
for any inter-servers communication
*/

//master-worker communication ---------
type GetTaskArgs struct {
	WorkerId string
}
type GetTaskReply struct {
	TaskAvailable bool
	TaskContent   string
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
	JobId      string
	ClientId   string
	MasterId   string
	JobContent string
	MQJobFound bool
}

type GetJobReply struct {
	IsAccepted bool //lock server will answer whether it accepted my job request
	JobId      string
	ClientId   string
	JobContent string
}
