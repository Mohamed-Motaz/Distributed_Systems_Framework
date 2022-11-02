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

//master-lockserver communication ---------
