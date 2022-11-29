package utils

type File struct {
	Name    string
	Content []byte
}

type FileType string

const ProcessExe FileType = "Process"
const DistributeExe FileType = "Distribute"
const AggregateExe FileType = "Aggregate"

type Error struct{
	IsFound bool
	Msg     string
}