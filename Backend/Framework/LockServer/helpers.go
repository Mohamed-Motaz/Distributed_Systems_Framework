package main

import (
	database "Framework/Database"
	logger "Framework/Logger"
	"Framework/RPC"
	utils "Framework/Utils"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)   
    
func (lockServer *LockServer) checkIsMasterAlive() {
	for {
		time.Sleep(time.Second * 5)
		lockServer.mu.Lock()
		for k, v := range lockServer.mastersState {
			if time.Since(v.lastHeartBeat) > 30*time.Minute {
				delete(lockServer.mastersState, k)
			} else if time.Since(v.lastHeartBeat) > 5*time.Second {
				v.Status = RPC.UNRESPONSIVE
			}
		}
		lockServer.mu.Unlock()
	}
}   
    
//  return true if there is a late job else false
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
	case utils.OptionalFiles:
		return OPTIONAL_FILES_FOLDER_NAME
	default:
		return ""
	}
    
}   
    