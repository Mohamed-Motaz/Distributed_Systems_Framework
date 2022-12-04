package utils

type File struct {
	Name    string
	Content []byte
}

type Folder struct {
	Type  FolderType
	Name  string
	Files []File
}

type FolderType string

const ProcessExe FolderType = "Process"
const DistributeExe FolderType = "Distribute"
const AggregateExe FolderType = "Aggregate"

type Error struct {
	IsFound bool
	Msg     string
}
