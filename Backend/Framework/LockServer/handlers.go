package main

import (
	database "Framework/Database"
	logger "Framework/Logger"
	"Framework/RPC"
	utils "Framework/Utils"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func (lockServer *LockServer) HandleGetJob(args *RPC.GetJobArgs, reply *RPC.GetJobReply) error {
	lockServer.getJobMu.Lock()
	defer lockServer.getJobMu.Unlock()

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
	err := lockServer.db.GetJobByJobId(jobsInfo, args.JobId).Error
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
		if err = lockServer.addJobToDB(args); err != nil {
			logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Job request rejected %+v because of db err %+v", args.JobId, err)
			*reply = RPC.GetJobReply{} //not accepted
			return nil
		}

		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Job request accepted %+v", reply)
		return nil
	}
	logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Job request rejected %+v", args.JobId)
	return nil
}

func (lockServer *LockServer) HandleDeleteBinaryFile(args *RPC.DeleteBinaryFileArgs, reply *RPC.DeleteBinaryFileReply) error {

	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Request to delete file %+v", args.FileName)
	// remove the record from the runnablefiles table
	if err := lockServer.db.DeleteRunnableFile(args.FileName, string(args.FileType)).Error; err != nil {
		logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to delete binary file %+v from runnableFiles table with err %+v", args.FileName, err)
		reply.Err = true
		reply.ErrMsg = fmt.Sprintf("Unable to delete binary file %+v from runnableFiles", args.FileName)
		return nil
	}
	reply.Err = false

	binaryFilePath := lockServer.getBinaryFilePath(
		lockServer.convertFileTypeToFolderType(args.FileType), args.FileName)

	err := os.Remove(binaryFilePath)

	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot delete file at this path %+v with err %+v", binaryFilePath, err)
		reply.Err = true
		reply.ErrMsg = fmt.Sprintf("Cannot delete file at this path %+v", binaryFilePath)
		return nil
	}
	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Done deleting file %+v", args.FileName)
	return nil
}

func (lockServer *LockServer) HandleDeleteOptionalFiles(args *RPC.DeleteOptionalFilesArgs, reply *RPC.DeleteOptionalFilesReply) error {

	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Request to delete optional file %+v", args.JobId)

	reply.Err = false

	optionalFilesFolderPath := lockServer.getOptionalFilesFolderPath(args.JobId)

	if _, err := os.Stat(optionalFilesFolderPath); errors.Is(err, os.ErrNotExist) {
		logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Done deleting optional file %+v", args.JobId)
		return nil
	}

	err := os.Remove(optionalFilesFolderPath)

	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot delete optional files at this path %+v with err %+v", optionalFilesFolderPath, err)
		reply.Err = true
		reply.ErrMsg = fmt.Sprintf("Cannot delete optional files at this path %+v", optionalFilesFolderPath)
		return nil
	}

	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Done deleting optional file %+v", args.JobId)

	return nil
}

func (lockServer *LockServer) HandleFinishedJob(args *RPC.FinishedJobArgs, reply *RPC.FinishedJobReply) error {

	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Request to submit finished job %+v", args)

	err := lockServer.db.DeleteJobByJobId(args.JobId).Error //todo decide whether or not to delete jobs
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot delete job id from the database %+v", err)
		return nil
	}

	err = os.Remove(lockServer.getOptionalFilesFolderPath(args.JobId))
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

	file, err := os.OpenFile(binaryFilePath, os.O_RDONLY, os.ModePerm)
	if !errors.Is(err, os.ErrNotExist) {
		// handle the case where the file already exists
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "The binary file name %+v already exists", args.File.Name)
		reply.Err = true
		reply.ErrMsg = fmt.Sprintf("The binary file name %+v already exists", args.File.Name) //todo error here is not 500, it is a bad request
		return nil
	}
	file.Close()

	err = utils.CreateAndWriteToFile(binaryFilePath, args.File.Content)

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
		reply.Err = true
		reply.ErrMsg = fmt.Sprintf("Failed while adding a runnableFile in the database %+v", err)
		return nil
	}
	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Done adding binary file %+v", args.File.Name)
	return nil
}

func (lockServer *LockServer) HandleAddOptionalFiles(args *RPC.OptionalFilesUploadArgs, reply *RPC.FileUploadReply) error {

	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Request to add optional files")

	reply.Err = false
	folderPath := lockServer.getOptionalFilesFolderPath(args.JobId)
	err := utils.CreateAndWriteToFile(filepath.Join(folderPath, args.FilesZip.Name), args.FilesZip.Content)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to create a file, fileName: %+v with error -> %+v", args.FilesZip.Name, err)
		reply.Err = true
		reply.ErrMsg = "Cannot create this file" + args.FilesZip.Name
		return nil
	}

	logger.LogInfo(logger.LOCK_SERVER, logger.DEBUGGING, "Done adding optional files")

	return nil
}

func (lockServer *LockServer) HandleGetBinaryFiles(args *RPC.GetBinaryFilesArgs, reply *RPC.GetBinaryFilesReply) error {

	var foundFile bool = false
	reply.Err = false
	reply.AggregateBinaryNames = make([]string, 0)
	reply.DistributeBinaryNames = make([]string, 0)
	reply.ProcessBinaryNames = make([]string, 0)

	files, err := ioutil.ReadDir(filepath.Join(string(BINARY_FILES_FOLDER_NAME), string(PROCESS_BINARY_FOLDER_NAME)))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from process binaries folder %+v", err)
	} else {
		foundFile = true
	}
	for _, file := range files {
		reply.ProcessBinaryNames = append(reply.ProcessBinaryNames, file.Name())
	}

	files, err = ioutil.ReadDir(filepath.Join(string(BINARY_FILES_FOLDER_NAME), string(utils.DistributeBinary)))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from distribute binary folder %+v", err)
	} else {
		foundFile = true
	}

	for _, file := range files {
		reply.DistributeBinaryNames = append(reply.DistributeBinaryNames, file.Name())
	}

	files, err = ioutil.ReadDir(filepath.Join(string(BINARY_FILES_FOLDER_NAME), string(utils.AggregateBinary)))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from aggregate binary folder %+v", err)
	} else {
		foundFile = true
	}

	for _, file := range files {
		reply.AggregateBinaryNames = append(reply.AggregateBinaryNames, file.Name())
	}

	if !foundFile {
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
	for _, masterState := range lockServer.mastersState {
		progress = append(progress, masterState.CurrentJobProgress)
	}
	reply.Progress = progress
	return nil
}
