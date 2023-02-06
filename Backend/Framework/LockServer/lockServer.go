package main

import (
	database "Framework/Database"
	logger "Framework/Logger"
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
		mxLateJobTime: time.Duration(-31) * time.Minute,
		mu:            sync.Mutex{},
		mastersState:  make(map[string]privCJP),
	}
	if err := lockServer.initDir(); err != nil {
		logger.FailOnError(logger.LOCK_SERVER, logger.ESSENTIAL, "Error while initializing Files: %v", err)
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
		logger.FailOnError(logger.LOCK_SERVER, logger.ESSENTIAL, "Error while listening on socket: %v", err)
	} else {
		logger.LogInfo(logger.LOCK_SERVER, logger.ESSENTIAL, "Listening on socket: %v", addrToListen)
	}

	go http.Serve(listener, nil)
	return nil
}

func (lockServer *LockServer) initDir() error {
	if err := os.MkdirAll("./"+filepath.Join(string(BINARY_FILES_FOLDER_NAME), string(PROCESS_BINARY_FOLDER_NAME)), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll("./"+filepath.Join(string(BINARY_FILES_FOLDER_NAME), string(AGGREGATE_BINARY_FOLDER_NAME)), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll("./"+filepath.Join(string(BINARY_FILES_FOLDER_NAME), string(DISTRIBUTE_BINARY_FOLDER_NAME)), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll("./"+string(OPTIONAL_FILES_FOLDER_NAME), os.ModePerm); err != nil {
		return err
	}
	return nil
}

// helper functions
