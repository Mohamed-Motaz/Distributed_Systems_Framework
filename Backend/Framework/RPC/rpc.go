package RPC

/*
This package contains all the RPC definitions
for any inter-servers communication
*/

//master-worker communication ---------
type GetTaskArgs struct {
}

type GetTaskReply struct {
	TaskContent   string
	TaskId        string
	JobId         string
	TaskAvailable bool
}

//master-lockserver communication ---------
