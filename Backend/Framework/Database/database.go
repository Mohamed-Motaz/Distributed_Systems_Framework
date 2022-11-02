package Database

import (
	logger "Framework/Logger"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//return a thread-safe *gorm.DB that can safely be used
//by multiple goroutines
func NewDbWrapper(address string) *DBWrapper {
	db := connect(address)
	logger.LogInfo(logger.DATABASE, logger.ESSENTIAL, "Db setup complete")
	wrapper := &DBWrapper{
		Db: db,
	}
	return wrapper
}

func connect(dsn string) *gorm.DB {
	//try to reconnect and sleep 10 seconds on failure
	var err error = fmt.Errorf("error")
	var db *gorm.DB
	ctr := 0

	for ctr < 3 && err != nil {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		ctr++
		if err != nil {
			time.Sleep(time.Second * 10)
		}
	}

	applyMigrations(db)

	if err != nil {
		logger.FailOnError(logger.DATABASE, logger.ESSENTIAL, "Unable to automigrate the db with this error %v", err)
	}
	return db
}

func CreateDBAddress(user, password, protocol, dbName, myHost, myPort, settings string) string {

	return fmt.Sprintf(
		"%v:%v@%v(%v:%v)/%v?%v",
		user, password, protocol, myHost, myPort, dbName, settings)
}

//fine, I'll do it myself
func applyMigrations(db *gorm.DB) {

	//create jobs database if it doesnt exist
	err := db.Exec("CREATE DATABASE IF NOT EXISTS `jobs`").Error
	if err != nil {
		logger.FailOnError(logger.DATABASE, logger.ESSENTIAL, "Unable to create database jobs with this error %v", err)
	}

	err = db.Exec(`
	CREATE TABLE IF NOT EXISTS jobs.jobsinfo (id bigint AUTO_INCREMENT,clientId longtext,masterId longtext,jobId longtext,content longtext,timeAssigned datetime(3) NULL,status longtext,PRIMARY KEY (id))
	`).Error

	if err != nil {
		logger.FailOnError(logger.DATABASE, logger.ESSENTIAL, "Unable to create the db with this error %v", err)
	}
}
