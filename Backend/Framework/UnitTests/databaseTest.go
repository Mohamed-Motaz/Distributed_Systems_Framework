package main

import (
	db "Framework/Database"
	logger "Framework/Logger"

	"time"
)

func main() {

	dbObj := db.NewDbWrapper()
	// Id           int       `gorm:"primaryKey; column:id"       json:"id"`
	// ClientId     string    `gorm:"column:clientId"             json:"clientId"`
	// MasterId     string    `gorm:"column:masterId"             json:"masterId"`
	// JobId        string    `gorm:"column:jobId"                json:"jobId"`
	// Content      string    `gorm:"column:content"              json:"content"`
	// TimeAssigned time.Time `gorm:"column:timeAssigned"         json:"timeAssigned"`
	// Status       JobStatus `gorm:"column:status"               json:"status"`

	jobInfo := &db.JobInfo{
		ClientId:     "Agina2",
		MasterId:     "Bedo2",
		JobId:        "Rawan2",
		Content:      "Salma2",
		TimeAssigned: time.Now(),
		Status:       db.IN_PROGRESS,
	}

	err := dbObj.CreateJobsInfo(jobInfo).Error
	if err != nil {
		logger.LogError(logger.DATABASE, logger.ESSENTIAL, "Unable to create job %+v with err %+v", jobInfo, err)
	}

	logger.LogInfo(logger.DATABASE, logger.ESSENTIAL, "%+v", jobInfo)

	jobsInfo := &[]db.JobInfo{}

	err = dbObj.GetAllJobsInfo(jobsInfo).Error
	if err != nil {
		logger.LogError(logger.DATABASE, logger.ESSENTIAL, "Unable to get all jobsinfo with err %+v", err)
	}
	logger.LogInfo(logger.DATABASE, logger.ESSENTIAL, "%+v", jobsInfo)

	lateJobsInfo := &[]db.JobInfo{}

	err = dbObj.GetAllLateInProgressJobsInfo(lateJobsInfo, time.Now().Add(time.Duration(-10)*time.Second)).Error
	if err != nil {
		logger.LogError(logger.DATABASE, logger.ESSENTIAL, "Unable to get all lateJobsInfo with err %+v", err)
	}
	logger.LogInfo(logger.DATABASE, logger.ESSENTIAL, "late jobs info\n%+v", lateJobsInfo)

}
