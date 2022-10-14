package Database

import (
	logger "Server/Logger"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBWrapper struct {
	Db *gorm.DB
}

//return a thread-safe *gorm.DB that can safely be used
//by multiple goroutines
func New() *DBWrapper {
	db := connect()
	logger.LogInfo(logger.DATABASE, logger.ESSENTIAL, "Db setup complete")
	wrapper := &DBWrapper{
		Db: db,
	}
	wrapper.ApplyMigrations()
	return wrapper
}

func connect() *gorm.DB {

	dsn := generateDSN(
		DbUser, DbPassword, DbProtocol,
		"", DbHost, DbPort, DbSettings)

	ctr := 0
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	for ctr < 5 {
		ctr++
		if err == nil {
			continue
		}
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		logger.FailOnError(logger.DATABASE, logger.ESSENTIAL, "Unable to connect to db with this error %v", err)
	}
	return db
}

func generateDSN(user, password, protocol, dbName, myHost, myPort, settings string) string {

	return fmt.Sprintf(
		"%v:%v@%v(%v:%v)/%v?%v",
		user, password, protocol, myHost, myPort, dbName, settings)
}
