package main

import (
	database "Framework/Database"
	"sync"

	"github.com/google/uuid"
)

func NewLockServer() *LockServer {

	lockServer := &LockServer{
		id:             uuid.NewString(), // random id
		mu:             sync.Mutex{},
		databaseWrapper: database.NewDbWrapper(database.CreateDBAddress(DbUser, DbPassword, DbProtocol, "", DbHost, DbPort, DbSettings)),
	}
	
}
