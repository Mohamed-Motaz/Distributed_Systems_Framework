package Database

import (
	"time"

	"gorm.io/gorm"
)

type JobInfo struct {
	Id           int       `gorm:"primaryKey; column:id"       json:"id"`
	ClientId     string    `gorm:"column:clientId"             json:"clientId"`
	MasterId     string    `gorm:"column:masterId"             json:"masterId"`
	JobId        string    `gorm:"column:jobId"                json:"jobId"`
	Content      string    `gorm:"column:content"              json:"content"`
	TimeAssigned time.Time `gorm:"column:timeAssigned"         json:"timeAssigned"`
	Status       JobStatus `gorm:"column:status"               json:"status"`
}

func (JobInfo) TableName() string {
	return "jobs.jobsinfo"
}

type JobStatus string

const IN_PROGESS JobStatus = "IN_PROGRESS"
const DONE JobStatus = "DONE"

func (dBWrapper *DBWrapper) GetAllJobsInfo(jobsInfo *[]JobInfo) *gorm.DB {
	return dBWrapper.Db.Raw("SELECT * FROM jobs.jobsinfo").Scan(jobsInfo)
}

func (dBWrapper *DBWrapper) CreateJobsInfo(jobInfo *JobInfo) *gorm.DB {
	return dBWrapper.Db.Create(jobInfo)
}
