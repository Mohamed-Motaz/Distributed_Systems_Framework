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
	TaskAvailable  bool
	TaskContent    string
	ProcessExe     []byte
	ProcessExeName string
	TaskId         string
	JobId          string
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
	IsAccepted        bool //lock server will answer whether it accepted my job request
	JobId             string
	ClientId          string
	JobContent        string
	ProcessExe        []byte
	ProcessExeName    string
	DistributeExe     []byte
	DistributeExeName string
	AggregateExe      []byte
	AggregateExeName  string
}

//websocketserver - lockserver communication --------------
type File struct {
	Name    string
	Content []byte
}

type FileType string

const ProcessExe FileType = "Process"
const DistributeExe FileType = "Distribute"
const AggregateExe FileType = "Aggregate"

type ProcessUploadArgs struct {
	FileType    FileType
	FileContent File
}

type ProcessUploadReply struct {
	Error    bool
	ErrorMsg string
}

type OptionalFilesUploadArgs struct {
	JobId       string
	FileContent []File
}

type OptionalFilesUploadReply struct {
	Error    bool
	ErrorMsg string
}

type FinishedJobArgs struct {
	JobId string
	MasterId string
}

type FinishedJobReply struct {
	Error bool
	ErrorMsg string
}
type GetExeFilesArgs struct {
}
type GetExeFilesReply struct {
	ProcessExeFiles    []string
	DistributeExeFiles []string
	AggregateExeFiles  []string
	Error              bool
	ErrorMsg           string
}
