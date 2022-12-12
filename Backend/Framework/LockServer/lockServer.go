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

func NewLockServer() *LockServer {

	lockServer := &LockServer{
		id:              uuid.NewString(), // random id
		mu:              sync.Mutex{},
		databaseWrapper: database.NewDbWrapper(database.CreateDBAddress(DbUser, DbPassword, DbProtocol, "", DbHost, DbPort, DbSettings)),
	}
	go lockServer.server()
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

func getExeFileContent(exeFileName string, exeFolderName string) ([]byte, error) {
	files, err := ioutil.ReadDir(filepath.Join("./ExeFiles", exeFolderName))
	var fileContent []byte
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from %v exes folder %v", exeFolderName, err)
		return fileContent, err
	}
	for _, file := range files {
		if file.Name() == exeFileName {
			filePath := filepath.Join("./ExeFiles", exeFolderName, string(file.Name()))
			fileContent, err = os.ReadFile(filePath)
			if err != nil {
				logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get file %v from %v exe folder %v", exeFileName, exeFolderName, err)
				return fileContent, err
			}
		}
	}
	return fileContent, nil
}

func getExeFiles(processExeName string, distributeExeName string, aggregateExeName string, reply *RPC.GetJobReply) error {
	processFileContent, err := getExeFileContent(processExeName, string(utils.ProcessExe))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get exe file %v from process exes folder %v", processExeName, err)
		return err
	}
	reply.ProcessExe.Name = processExeName
	reply.ProcessExe.Content = processFileContent

	distributeFileContent, err := getExeFileContent(distributeExeName, string(utils.DistributeExe))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get exe file %v from distribute exes folder %v", distributeExeName, err)
		return err
	}
	reply.DistributeExe.Name = distributeExeName
	reply.DistributeExe.Content = distributeFileContent

	aggregateFileContent, err := getExeFileContent(aggregateExeName, string(utils.AggregateExe))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get exe file %v from aggregate exes folder %v", aggregateExeName, err)
		return err
	}
	reply.AggregateExe.Name = aggregateExeName
	reply.AggregateExe.Content = aggregateFileContent
	return nil
}

func getOptionalFiles(OptionalFilesNames []string) ([]utils.File, error) {
	files, err := ioutil.ReadDir("./OptionalFiles")
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from optional files folder %v", err)
	}
	var optionalFiles []utils.File
	var optionalFile utils.File
	for _, optionalFileName := range OptionalFilesNames {
		for _, file := range files {
			if file.Name() == optionalFileName {
				filePath := filepath.Join("./OptionalFiles", optionalFileName)
				fileContent, err := os.ReadFile(filePath)
				if err != nil {
					logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get file %v from optional files folder %v", optionalFileName, err)
					return optionalFiles, err
				}
				optionalFile.Name = optionalFileName
				optionalFile.Content = fileContent
				optionalFiles = append(optionalFiles, optionalFile)
			}

		}
	}
	return optionalFiles, nil
}
func (lockServer *LockServer) addJobToDB(args *RPC.GetJobArgs) {
	// assign job to the master and update database
	jobInfo := database.JobInfo{}
	jobInfo.ClientId = args.ClientId
	jobInfo.MasterId = args.MasterId
	jobInfo.JobId = args.JobId
	jobInfo.Content = args.JobContent
	jobInfo.TimeAssigned = time.Now()
	jobInfo.Status = database.IN_PROGRESS
	jobInfo.ProcessExeName = args.ProcessExeName
	jobInfo.DistributeExeName = args.DistributeExeName
	jobInfo.AggregateExeName = args.AggregateExeName
	jobInfo.OptionalFilesNames = args.OptionalFilesNames
	// add job to database
	err := lockServer.databaseWrapper.CreateJobsInfo(&jobInfo).Error
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Failed while adding a job in database", err)
	}
	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Job added successfully", args.JobId)

}
func (lockServer *LockServer) HandleGetJob(args *RPC.GetJobArgs, reply *RPC.GetJobReply) error {
	logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "A master request job", args)
	reply.IsAccepted = false
	if lockServer.getLateJob(args, reply) {
		// I found a late job, take it.
		logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Assigned late job to the master%v", args)
		return nil
	}
	//check the job isn't assigned to other master
	jobsInfo := []database.JobInfo{}
	err := lockServer.databaseWrapper.CheckIsJobAssigned(&jobsInfo, args.JobId).Error
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Failed while checking if this job is assigned in database", err)
	}
	if len(jobsInfo) == 0 {
		// the job isn't taken by any other master, send reply to the master
		reply.IsAccepted = true
		reply.JobId = args.JobId
		reply.ClientId = args.ClientId
		reply.JobContent = args.JobContent
		err = getExeFiles(args.ProcessExeName, args.DistributeExeName, args.AggregateExeName, reply)
		if err != nil {
			logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get exe files %v", err)
		}
		optionalFiles, err := getOptionalFiles(args.OptionalFilesNames)
		if err != nil {
			logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from optional files folder %v", err)
		}
		reply.OptionalFiles = optionalFiles
		// assign job to the master and update database
		// jobInfo := database.JobInfo{}
		// jobInfo.ClientId = args.ClientId
		// jobInfo.MasterId = args.MasterId
		// jobInfo.JobId = args.JobId
		// jobInfo.Content = args.JobContent
		// jobInfo.TimeAssigned = time.Now()
		// jobInfo.Status = database.IN_PROGRESS
		// jobInfo.ProcessExeName = args.ProcessExeName
		// jobInfo.DistributeExeName = args.DistributeExeName
		// jobInfo.AggregateExeName = args.AggregateExeName
		// jobInfo.OptionalFilesNames = args.OptionalFilesNames
		// // add job to database
		// err = lockServer.databaseWrapper.CreateJobsInfo(&jobInfo).Error
		// if err != nil {
		// 	logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Failed while adding a job in database", err)
		// }
		// logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Job added successfully", args.JobId)
		lockServer.addJobToDB(args)
		return nil
	}
	logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Job request rejected %v", args.JobId)
	return nil
}

// helper function

func (lockServer *LockServer) getLateJob(args *RPC.GetJobArgs, reply *RPC.GetJobReply) bool {
	// return true if there is a late job else false

	// if there is late job it will be assigned to reply
	jobsInfo := []database.JobInfo{}
	// jobsInfo will be filled by all late jobs.
	// before 1 min
	err := lockServer.databaseWrapper.GetAllLateInProgressJobsInfo(&jobsInfo, time.Now().Add(time.Duration(-60)*time.Second)).Error

	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to connect to database", err)
	}
	if len(jobsInfo) == 0 {
		logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "There is no late job %+v", args)
		return false
	}
	// found job, let's assign it :')

	lateJob := &jobsInfo[0] // as jobsInfo is sorted
	logger.LogDelay(logger.LOCK_SERVER, logger.ESSENTIAL, "Found a late job that will be reassigned %+v", lateJob)
	// assign late job to the master
	reply.ClientId = lateJob.ClientId
	reply.JobId = lateJob.JobId
	reply.JobContent = lateJob.Content
	reply.IsAccepted = false
	getExeFiles(lateJob.ProcessExeName, lateJob.DistributeExeName, lateJob.AggregateExeName, reply)
	optionalFiles, err := getOptionalFiles(lateJob.OptionalFilesNames)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from optional files folder %v", err)
	}
	reply.OptionalFiles = optionalFiles
	return true
}

func createFolderIfNotExist(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot create folder %+v", err)
			return false
		}
	}
	return true
}

func deleteFolder(path string) error {
	err := os.Remove(path)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot delete folder at this path %+v", path, err)
	}
	return err
}
func deleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot delete file at this path %+v", path, err)
	}

	return err
}
func (lockServer *LockServer) HandleDeleteExeFile(args *RPC.DeleteExeFileArgs, reply *RPC.DeleteExeFileReply) error {
	reply.Err = false
	exeFolderName := string(args.FileType)
	filePath := filepath.Join("./ExeFiles", exeFolderName, args.FileName)
	err := deleteFile(filePath)
	if err != nil {
		reply.Err = true
		reply.ErrMsg = err.Error()
		return nil
	}
	return nil
}

func (lockServer *LockServer) HandleFinishedJob(args *RPC.FinishedJobArgs, reply *RPC.FinishedJobReply) error {
	reply.Err = false
	path := filepath.Join("./OptionalFiles", args.JobId)
	err := deleteFolder(path)
	if err != nil {
		reply.Err = true
		reply.ErrMsg = err.Error()
		return nil
	}

	err = lockServer.databaseWrapper.DeleteJobById(args.JobId).Error
	if err != nil {
		reply.Err = true
		reply.ErrMsg = err.Error()
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot delete job id from the database %+v", err)
		return nil
	}

	return nil
}

func (lockServer *LockServer) HandleAddExeFile(args *RPC.ExeUploadArgs, reply *RPC.FileUploadReply) error {
	reply.Err = false
	path := filepath.Join("./ExeFiles", string(args.FileType))
	isFound := createFolderIfNotExist(path)
	if !isFound {
		reply.Err = true
		reply.ErrMsg = "Cannot create folder with this exe file " + string(args.FileType)
		return nil
	}
	fileOut, err := os.Create(filepath.Join(path, args.File.Name))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to add this exe file, fileName: %v", string(args.FileType), err)
		reply.Err = true
		reply.ErrMsg = "Cannot add this file" + args.File.Name
		return nil
	}
	defer fileOut.Close()
	return nil
}

func (lockServer *LockServer) HandleAddOptionalFiles(args *RPC.OptionalFilesUploadArgs, reply *RPC.FileUploadReply) error {
	reply.Err = false
	path := filepath.Join("./OptionalFiles", args.JobId)
	isFound := createFolderIfNotExist(path)
	if !isFound {
		reply.Err = true
		reply.ErrMsg = "Cannot create folder with this JobId " + args.JobId
	}
	for i := 0; i < len(args.Files); i++ {

		fileOut, err := os.Create(filepath.Join(path, args.Files[i].Name))
		if err != nil {
			logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to create a file, fileName: %v", args.Files[i].Name, err)
			reply.Err = true
			reply.ErrMsg = "Cannot create this file" + args.Files[i].Name
		}
		defer fileOut.Close()
	}
	return nil
}

func (lockServer *LockServer) HandleGetExeFiles(reply *RPC.GetExeFilesReply) error {
	files, err := ioutil.ReadDir(filepath.Join("./ExeFiles", string(utils.ProcessExe)))
	var foundError bool = false
	reply.Err = false
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from process exes folder %v", err)
		foundError = true
	}
	for _, file := range files {
		reply.ProcessExeNames = append(reply.ProcessExeNames, file.Name())
	}
	files, err = ioutil.ReadDir(filepath.Join("./ExeFiles", string(utils.DistributeExe)))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from distribute exe folder %v", err)
		foundError = true
	}
	for _, file := range files {
		reply.DistributeExeNames = append(reply.DistributeExeNames, file.Name())
	}
	files, err = ioutil.ReadDir(filepath.Join("./ExeFiles", string(utils.AggregateExe)))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from aggregate exe folder %v", err)
		foundError = true
	}
	for _, file := range files {
		reply.AggregateExeNames = append(reply.AggregateExeNames, file.Name())
	}
	if foundError {
		reply.Err = true
		reply.ErrMsg = "There is an error while getting exe files."
	}
	return nil
}
