package main

import (
	database "Framework/Database"
	logger "Framework/Logger"
	"Framework/RPC"
	utils "Framework/Utils"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

//folder architecture
//-Binaries
//--Process
//---zipFile
//---zipFile
//---zipFile
//--Distribute
//--Aggregate
//-OptionalFiles
//--JobId
//---zipFile

func NewLockServer() *LockServer {

	lockServer := &LockServer{
		id:            uuid.NewString(), // random id
		db:            database.NewDbWrapper(database.CreateDBAddress(DbUser, DbPassword, DbProtocol, "", DbHost, DbPort, DbSettings)),
		mxLateJobTime: time.Duration(-60) * time.Second,
		mu:            sync.Mutex{},
		mastersState:  make(map[string]RPC.CurrentJobProgress),
	}
	go lockServer.server()
	go lockServer.checkIsMasterAlive()
	return lockServer
}

func CreateLockServerAddress() string {
	return MyHost + ":" + MyPort
}

func (lockServer *LockServer) server() error {
	rpc.Register(lockServer)
	rpc.HandleHTTP()

	addrToListen := CreateLockServerAddress()

	os.Remove(addrToListen)
	listener, err := net.Listen("tcp", addrToListen)

	if err != nil {
		logger.FailOnError(logger.MASTER, logger.ESSENTIAL, "Error while listening on socket: %v", err)
	} else {
		logger.LogInfo(logger.MASTER, logger.ESSENTIAL, "Listening on socket: %v", addrToListen)
	}

	go http.Serve(listener, nil)
	return nil
}

func (lockServer *LockServer) HandleGetJob(args *RPC.GetJobArgs, reply *RPC.GetJobReply) error {
	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "A master requests job %+v", args)
	reply.IsAccepted = false
	if lockServer.assignLateJob(args, reply) {
		// I found a late job, take it.
		logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Assigned late job to the master %+v", args)
		return nil
	}

	if !args.MQJobFound { //master doesn't have a job from the mq, so wants me to provide a late job
		logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to find late job to give to the master %+v", args)
		*reply = RPC.GetJobReply{} //not accepted
		return nil
	}

	//check the job isn't assigned to other master
	jobsInfo := &database.JobInfo{}
	err := lockServer.db.CheckIsJobAssigned(jobsInfo, args.JobId).Error
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Failed while checking if this job is assigned to another master %+v", err)
		return nil
	}
	if jobsInfo.Id == 0 {
		// the job isn't taken by any other master, send reply to the master
		reply.IsAccepted = true
		reply.JobId = args.JobId
		reply.ClientId = args.ClientId
		reply.JobContent = args.JobContent
		processBinary, distributeBinary, aggregateBinary, err := lockServer.setBinaryFiles(args.ProcessBinaryName, args.DistributeBinaryName, args.AggregateBinaryName)
		if err != nil {
			logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get binary files %+v", err)
			*reply = RPC.GetJobReply{} //not accepted
			return nil
		}
		reply.ProcessBinary = processBinary
		reply.DistributeBinary = distributeBinary
		reply.AggregateBinary = aggregateBinary
		optionalFiles, err := lockServer.getOptionalFiles(args.JobId)
		if err != nil {
			logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from optional files folder %+v", err)
			*reply = RPC.GetJobReply{} //not accepted
			return nil
		}
		// receive the optionalFiles from the ws server as a zip file
		reply.OptionalFilesZip = optionalFiles
		lockServer.addJobToDB(args)
		return nil
	}
	logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Job request rejected %+v", args.JobId)
	return nil
}

func (lockServer *LockServer) HandleDeleteBinaryFile(args *RPC.DeleteBinaryFileArgs, reply *RPC.DeleteBinaryFileReply) error {
	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Request to delete file %+v", args.FileName)
	reply.Err = false
	binaryFilePath := lockServer.getBinaryFilePath(
		lockServer.convertFileTypeToFolderType(args.FileType), args.FileName)

	err := os.Remove(binaryFilePath)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot delete file at this path %+v with err %+v", binaryFilePath, err)
		reply.Err = true
		reply.ErrMsg = err.Error()
		return nil
	}
	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Done deleting file %+v", args.FileName)
	return nil
}

func (lockServer *LockServer) HandleFinishedJob(args *RPC.FinishedJobArgs, reply *RPC.FinishedJobReply) error {
	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Request to submit finished job %+v", args)

	err := lockServer.db.DeleteJobById(args.JobId).Error //todo decide whether or not to delete jobs
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot delete job id from the database %+v", err)
		return nil
	}

	err = deleteFolder(lockServer.getOptionalFilesFolderPath(args.JobId))
	if err != nil {
		return nil
	}
	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Done handling finished job %+v", args)
	return nil
}

func (lockServer *LockServer) HandleAddBinaryFile(args *RPC.BinaryUploadArgs, reply *RPC.FileUploadReply) error {
	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Request to add binary file", args.File.Name)
	reply.Err = false
	binaryFilePath := lockServer.getBinaryFilePath(
		lockServer.convertFileTypeToFolderType(args.FileType), args.File.Name)

	err := utils.CreateAndWriteToFile(binaryFilePath, args.File.Content)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to add this binary file, fileName: %v with err %+v", args.File.Name, err)
		reply.Err = true
		reply.ErrMsg = "Cannot add this file " + args.File.Name
		return nil
	}
	runnableFile := &database.RunnableFiles{
		BinaryName:   args.File.Name,
		BinaryType:   string(args.FileType),
		BinaryRunCmd: args.File.RunCmd,
	}
	// add runnableFile to database
	err = lockServer.db.CreateRunnableFile(runnableFile).Error
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Failed while adding a runnableFile in the database %+v", err)
		return err
	}
	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "RunnableFile added successfully %+v", args.File.Name)

	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Done adding binary file %+v", args.File.Name)
	return nil
}

func (lockServer *LockServer) HandleAddOptionalFiles(args *RPC.OptionalFilesUploadArgs, reply *RPC.FileUploadReply) error {
	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Request to add optional files")

	reply.Err = false
	folderPath := lockServer.getOptionalFilesFolderPath(args.JobId)
	err := utils.CreateAndWriteToFile(filepath.Join(folderPath, args.FilesZip.Name), args.FilesZip.Content)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to create a file, fileName: %+v", args.FilesZip.Name, err)
		reply.Err = true
		reply.ErrMsg = "Cannot create this file" + args.FilesZip.Name
		return nil
	}

	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Done adding optional files")
	return nil
}

func (lockServer *LockServer) HandleGetBinaryFiles(args *RPC.GetBinaryFilesArgs, reply *RPC.GetBinaryFilesReply) error {
	var foundError bool = false
	reply.Err = false

	files, err := ioutil.ReadDir(filepath.Join(string(BINARY_FILES_FOLDER_NAME), string(PROCESS_BINARY_FOLDER_NAME)))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from process binaries folder %+v", err)
		foundError = true
	}
	for _, file := range files {
		reply.ProcessBinaryNames = append(reply.ProcessBinaryNames, file.Name())
	}

	files, err = ioutil.ReadDir(filepath.Join(string(BINARY_FILES_FOLDER_NAME), string(utils.DistributeBinary)))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from distribute binary folder %+v", err)
		foundError = true
	}
	for _, file := range files {
		reply.DistributeBinaryNames = append(reply.DistributeBinaryNames, file.Name())
	}

	files, err = ioutil.ReadDir(filepath.Join(string(BINARY_FILES_FOLDER_NAME), string(utils.AggregateBinary)))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from aggregate binary folder %+v", err)
		foundError = true
	}
	for _, file := range files {
		reply.AggregateBinaryNames = append(reply.AggregateBinaryNames, file.Name())
	}

	if foundError {
		reply.Err = true
		reply.ErrMsg = "There is an error while getting binary files."
	}
	return nil
}

func (lockServer *LockServer) HandleSetJobProgress(args *RPC.SetJobProgressArgs, reply *RPC.SetJobProgressReply) error {
	lockServer.mu.Lock()
	defer lockServer.mu.Unlock()
	lockServer.mastersState[args.MasterId] = privCJP{
		lastHeartBeat:      time.Now(),
		CurrentJobProgress: args.CurrentJobProgress,
	}
	return nil
}

func (lockServer *LockServer) HandleGetSystemProgress(args *RPC.GetSystemProgressArgs, reply *RPC.GetSystemProgressReply) error {
	lockServer.mu.Lock()
	defer lockServer.mu.Unlock()

	progress := make([]RPC.CurrentJobProgress, 0)
	for _, v := range lockServer.mastersState {
		progress = append(progress, v.CurrentJobProgress)
	}
	reply.Progress = progress
	return nil
}

// helper functions
func (lockServer *LockServer) checkIsMasterAlive() {
	for {
		time.Sleep(time.Second * 5)

	}
}

// return true if there is a late job else false
func (lockServer *LockServer) assignLateJob(args *RPC.GetJobArgs, reply *RPC.GetJobReply) bool {
	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Attempting to assign late job to master %+v", args.MasterId)

	lateJob := &database.JobInfo{}

	err := lockServer.db.GetLatestInProgressJobsInfo(lateJob, time.Now().Add(lockServer.mxLateJobTime)).Error

	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to get in progress jobs %+v for master %+v", err, args.MasterId)
		*reply = RPC.GetJobReply{} //not accepted
		return false
	}

	if lateJob.Id == 0 {
		logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "There is no late job for master %+v", args.MasterId)
		*reply = RPC.GetJobReply{} //not accepted
		return false
	}

	// found job, let's assign it :')

	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Found a late job that will be reassigned %+v", lateJob)

	// assign late job to the master
	reply.IsAccepted = true
	reply.JobId = lateJob.JobId
	reply.ClientId = lateJob.ClientId
	reply.JobContent = lateJob.Content
	processBinary, distributeBinary, aggregateBinary, err := lockServer.setBinaryFiles(args.ProcessBinaryName, args.DistributeBinaryName, args.AggregateBinaryName)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from binary files folder %+v", err)
		*reply = RPC.GetJobReply{} //not accepted
		return false
	}
	reply.ProcessBinary = processBinary
	reply.DistributeBinary = distributeBinary
	reply.AggregateBinary = aggregateBinary
	optionalFiles, err := lockServer.getOptionalFiles(lateJob.JobId)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from optional files folder %+v", err)
		*reply = RPC.GetJobReply{} //not accepted
		return false
	}
	// receive the optionalFiles from the ws server as a zip file
	reply.OptionalFilesZip = optionalFiles
	return true
}

func deleteFolder(path string) error {
	err := os.Remove(path)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot delete folder at this path %+v", path, err)
	}
	return err
}
func (lockServer *LockServer) getBinaryRunnableFileFromDB(folderName FolderName, fileType utils.FileType, binaryName, binaryType string) (utils.RunnableFile, error) {
	binaryRunnableFile := utils.RunnableFile{
		File: utils.File{
			Content: make([]byte, 0),
		},
	}
	runnableFile := &database.RunnableFiles{}
	binaryFileContent, err := os.ReadFile(
		lockServer.getBinaryFilePath(folderName, binaryName))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get binary file %+v from %+v folder %+v", binaryName, binaryType, err)
		return binaryRunnableFile, err
	}
	err = lockServer.db.GetRunCmdOfBinary(runnableFile, binaryName, string(fileType)).Error
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to get in runCmd of %+v %+v", binaryType, err)
		return binaryRunnableFile, err
	}
	if runnableFile.Id == 0 {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "There is no %+v binary file with this name %+v in db %+v", binaryType, binaryName, err)
		return binaryRunnableFile, err
	}
	binaryRunnableFile.File = utils.File{
		Name:    binaryName,
		Content: binaryFileContent,
	}
	binaryRunnableFile.RunCmd = runnableFile.BinaryRunCmd
	return binaryRunnableFile, nil
}

func (lockServer *LockServer) setBinaryFiles(processBinaryName, distributeBinaryName, aggregateBinaryName string) (utils.RunnableFile, utils.RunnableFile, utils.RunnableFile, error) {
	ProcessBinary := utils.RunnableFile{}
	DistributeBinary := utils.RunnableFile{}
	AggregateBinary := utils.RunnableFile{}
	ProcessBinary, err := lockServer.getBinaryRunnableFileFromDB(PROCESS_BINARY_FOLDER_NAME, utils.ProcessBinary, processBinaryName, "Process")
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "cannot get runCmd of processBinaryName %+v from the db %+v", processBinaryName, err)
		return ProcessBinary, DistributeBinary, AggregateBinary, err
	}
	DistributeBinary, err = lockServer.getBinaryRunnableFileFromDB(DISTRIBUTE_BINARY_FOLDER_NAME, utils.DistributeBinary, distributeBinaryName, "Distribute")
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "cannot get runCmd of distributeBinaryName %+v from the db %+v", distributeBinaryName, err)
		return ProcessBinary, DistributeBinary, AggregateBinary, err
	}
	AggregateBinary, err = lockServer.getBinaryRunnableFileFromDB(AGGREGATE_BINARY_FOLDER_NAME, utils.AggregateBinary, aggregateBinaryName, "Aggregate")
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "cannot get runCmd of aggregateBinaryName %+v from the db %+v", aggregateBinaryName, err)
		return ProcessBinary, DistributeBinary, AggregateBinary, err
	}
	return ProcessBinary, DistributeBinary, AggregateBinary, nil
}

func (lockServer *LockServer) getOptionalFiles(jobId string) (utils.File, error) {
	//-OptionalFiles
	//--JobId
	//---actualFiles
	//---actualFiles
	//---actualFiles
	optionalFilesZip := utils.File{
		Content: make([]byte, 0),
	}
	optionalFilesFolderPath := lockServer.getOptionalFilesFolderPath(jobId)
	if _, err := os.Stat(optionalFilesFolderPath); errors.Is(err, os.ErrNotExist) {
		return optionalFilesZip, err
	}

	file, err := ioutil.ReadDir(optionalFilesFolderPath)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get optional zip file %+v", err)
		return optionalFilesZip, err
	}

	optionalFiles, err := os.ReadFile(filepath.Join(optionalFilesFolderPath, file[0].Name()))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL,
			"Cannot get file %+v from optional files folder with err %+v", filepath.Join(optionalFilesFolderPath, file[0].Name()), err)
		return optionalFilesZip, err //todo should I actually return the remaining files or not
	}
	optionalFilesZip = utils.File{
		Name:    file[0].Name(),
		Content: optionalFiles,
	}
	return optionalFilesZip, nil

}

func (lockServer *LockServer) addJobToDB(args *RPC.GetJobArgs) {
	// assign job to the master and update database
	jobInfo := &database.JobInfo{
		ClientId:             args.ClientId,
		MasterId:             args.MasterId,
		JobId:                args.JobId,
		Content:              args.JobContent,
		TimeAssigned:         time.Now(),
		Status:               database.IN_PROGRESS,
		ProcessBinaryName:    args.ProcessBinaryName,
		DistributeBinaryName: args.DistributeBinaryName,
		AggregateBinaryName:  args.AggregateBinaryName,
		//You need to create a new table, that maps each process name, to its run cmd.
		//You will receive the run cmd from the ws server when uploading a binary
	}

	// add job to database
	err := lockServer.db.CreateJobsInfo(jobInfo).Error
	if err != nil {
		//todo should there be an error returned to the master?
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Failed while adding a job in database %+v", err)
		return
	}
	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Job added successfully %+v", args.JobId)

}

func (lockServer *LockServer) getBinaryFilePath(folderName FolderName, fileName string) string {
	//-Binaries
	//--Process
	//---zipFile
	return filepath.Join(string(BINARY_FILES_FOLDER_NAME), string(folderName), fileName)
}

func (lockServer *LockServer) getOptionalFilesFolderPath(jobId string) string {
	//-OptionalFiles
	//--JobId
	//---actualFiles
	return filepath.Join(string(OPTIONAL_FILES_FOLDER_NAME), jobId)
}

func (lockServer *LockServer) convertFileTypeToFolderType(fileType utils.FileType) FolderName {
	switch fileType {
	case utils.ProcessBinary:
		return PROCESS_BINARY_FOLDER_NAME
	case utils.DistributeBinary:
		return DISTRIBUTE_BINARY_FOLDER_NAME
	case utils.AggregateBinary:
		return AGGREGATE_BINARY_FOLDER_NAME
	default:
		return ""
	}
}
