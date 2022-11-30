package Database

import (
	"time"

	"gorm.io/gorm"
)

type JobInfo struct {
	Id                 int       `gorm:"primaryKey; column:id"       json:"id"`
	ClientId           string    `gorm:"column:clientId"             json:"clientId"`
	MasterId           string    `gorm:"column:masterId"             json:"masterId"`
	JobId              string    `gorm:"column:jobId"                json:"jobId"`
	Content            string    `gorm:"column:content"              json:"content"`
	TimeAssigned       time.Time `gorm:"column:timeAssigned"         json:"timeAssigned"`
	Status             JobStatus `gorm:"column:status"               json:"status"`
	ProcessExeName     string    `gorm:"column:processExeName"       json:"processExeName"`
	DistributeExeName  string    `gorm:"column:distributeExeName"    json:"distributeExeName"`
	AggregateExeName   string    `gorm:"column:aggregateExeName"     json:"aggregateExeName"`
	OptionalFilesNames string    `gorm:"column:optionalFilesNames"   json:"optionalFilesNames"`
}

func (JobInfo) TableName() string {
	return "jobs.jobsinfo"
}

type JobStatus string

const IN_PROGRESS JobStatus = "IN_PROGRESS"
const DONE JobStatus = "DONE"

func (dBWrapper *DBWrapper) GetAllJobsInfo(jobsInfo *[]JobInfo) *gorm.DB {
	return dBWrapper.Db.Raw("SELECT * FROM jobs.jobsinfo").Scan(jobsInfo)
}

func (dBWrapper *DBWrapper) GetAllLateInProgressJobsInfo(jobsInfo *[]JobInfo, maxTime time.Time) *gorm.DB {
	return dBWrapper.Db.Raw(`
	SELECT * FROM jobs.jobsinfo 
	WHERE status = ? AND timeAssigned < ?
	ORDER BY timeAssigned ASC
	`, IN_PROGRESS, maxTime).Scan(jobsInfo)
}

// check if the job is assigned to another master

func (dBWrapper *DBWrapper) CheckIsJobAssigned(jobsInfo *[]JobInfo, jobId string) *gorm.DB {
	return dBWrapper.Db.Raw(`
	SELECT * FROM jobs.jobsinfo 
	WHERE jobId = ?
		`, jobId).Scan(jobsInfo)
}

// delete job by id

func (dBWrapper *DBWrapper) DeleteJobById(jobId string) *gorm.DB {
	return dBWrapper.Db.Delete(jobId)
}

// assign job to master

func (dBWrapper *DBWrapper) CreateJobsInfo(jobInfo *JobInfo) *gorm.DB {
	return dBWrapper.Db.Create(jobInfo)
}
