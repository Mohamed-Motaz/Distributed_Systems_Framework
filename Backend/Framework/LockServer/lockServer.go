package main

import (
	database "Framework/Database"
	logger "Framework/Logger"
	"Framework/RPC"
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
		reply.JobId = args.JobId
		reply.ClientId = args.ClientId
		reply.JobContent = args.JobContent
		reply.IsAccepted = true
		// assign job to the master and update database
		jobInfo := database.JobInfo{}
		jobInfo.ClientId = args.ClientId
		jobInfo.MasterId = args.MasterId
		jobInfo.JobId = args.JobId
		jobInfo.Content = args.JobContent
		jobInfo.TimeAssigned = time.Now()
		jobInfo.Status = database.IN_PROGRESS
		// add job to database
		err := lockServer.databaseWrapper.CreateJobsInfo(&jobInfo).Error
		if err != nil {
			logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Failed while creating a job", err)
		}
		logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Job added successfully", args.JobId)
		return nil
	}
	logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Job request rejected %v", args.JobId)
	return nil
}

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
	reply.JobId = lateJob.JobId
	reply.JobContent = lateJob.Content
	reply.IsAccepted = false
	reply.ClientId = lateJob.ClientId
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

func (lockServer *LockServer) HandleAddExeFile(args *RPC.ProcessUploadArgs, reply *RPC.ProcessUploadReply) error {
	reply.Error = false
	path := filepath.Join("./ExeFiles", string(args.FileType))
	isFound := createFolderIfNotExist(path)
	if !isFound {
		reply.Error = true
		reply.ErrorMsg = "Cannot create folder with this exe file " + string(args.FileType)
	}
	fileOut, err := os.Create(filepath.Join(path, args.FileContent.Name))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to add this exe file, fileName: %v", string(args.FileType), err)
		reply.Error = true
		reply.ErrorMsg = "Cannot add this file" + args.FileContent.Name
	}
	defer fileOut.Close()
	return nil
}

func (lockServer *LockServer) HandleAddOptionalFiles(args *RPC.OptionalFilesUploadArgs, reply *RPC.OptionalFilesUploadReply) error {
	reply.Error = false
	path := filepath.Join("./OptionalFiles", args.JobId)
	isFound := createFolderIfNotExist(path)
	if !isFound {
		reply.Error = true
		reply.ErrorMsg = "Cannot create folder with this JobId " + args.JobId
	}
	for i := 0; i < len(args.FileContent); i++ {

		fileOut, err := os.Create(filepath.Join(path, args.FileContent[i].Name))
		if err != nil {
			logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Unable to create a file, fileName: %v", args.FileContent[i].Name, err)
			reply.Error = true
			reply.ErrorMsg = "Cannot create this file" + args.FileContent[i].Name
		}
		defer fileOut.Close()
	}
	return nil
}

func (lockServer *LockServer) HandleGetExeFiles(args *RPC.GetExeFilesArgs, reply *RPC.GetExeFilesReply) error {
	files, err := ioutil.ReadDir(filepath.Join("./ExeFiles", string(RPC.ProcessExe)))
	var foundError bool = false
	reply.Error = false
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from process exes folder %v", err)
		foundError = true
	}
	for _, file := range files {
		reply.ProcessExeFiles = append(reply.ProcessExeFiles, file.Name())
	}
	files, err = ioutil.ReadDir(filepath.Join("./ExeFiles", string(RPC.DistributeExe)))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from distribute exe folder %v", err)
		foundError = true
	}
	for _, file := range files {
		reply.DistributeExeFiles = append(reply.DistributeExeFiles, file.Name())
	}
	files, err = ioutil.ReadDir(filepath.Join("./ExeFiles", string(RPC.AggregateExe)))
	if err != nil {
		logger.LogError(logger.LOCK_SERVER, logger.ESSENTIAL, "Cannot get files from aggregate exe folder %v", err)
		foundError = true
	}
	for _, file := range files {
		reply.AggregateExeFiles = append(reply.AggregateExeFiles, file.Name())
	}
	if foundError {
		reply.Error = true
		reply.ErrorMsg = "There is an error while getting exe files."
	}
	return nil
}
