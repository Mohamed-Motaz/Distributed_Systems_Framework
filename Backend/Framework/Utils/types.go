package utils

type File struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}

type RunnableFile struct {
	Id int
	File
	RunCmd string //cmd to run the file
}

type FileType string

const ProcessBinary FileType = "Process"
const DistributeBinary FileType = "Distribute"
const AggregateBinary FileType = "Aggregate"
const OptionalFiles FileType = "OptionalFiles"

type Error struct {
	Err    bool   `json:"error"`
	ErrMsg string `json:"errorMsg"`
}

type HttpResponse struct {
	Success  bool        `json:"success"`
	Response interface{} `json:"response"`
}
