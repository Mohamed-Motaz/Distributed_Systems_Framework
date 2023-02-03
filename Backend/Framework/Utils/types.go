package utils

type File struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}

type RunnableFile struct {
	File
	RunCmd string //cmd to run the file
}

type FileType string

const ProcessBinary FileType = "Process"
const DistributeBinary FileType = "Distribute"
const AggregateBinary FileType = "Aggregate"

type Error struct {
	Err    bool   `json:"error"`
	ErrMsg string `json:"errorMsg"`
}
