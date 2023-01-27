package Database

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type JobInfo struct {
	Id                     int       `gorm:"primaryKey; column:id"       json:"id"`
	ClientId               string    `gorm:"column:clientId"             json:"clientId"`
	MasterId               string    `gorm:"column:masterId"             json:"masterId"`
	JobId                  string    `gorm:"column:jobId"                json:"jobId"`
	Content                string    `gorm:"column:content"              json:"content"`
	TimeAssigned           time.Time `gorm:"column:timeAssigned"         json:"timeAssigned"`
	Status                 JobStatus `gorm:"column:status"               json:"status"`
	ProcessBinaryName      string    `gorm:"column:processBinaryName"       json:"processBinaryName"`
	DistributeBinaryName   string    `gorm:"column:distributeBinaryName"    json:"distributeBinaryName"`
	AggregateBinaryName    string    `gorm:"column:aggregateBinaryName"     json:"aggregateBinaryName"`
	ProcessBinaryRunCmd    string    `gorm:"column:processBinaryRunCmd"       json:"processBinaryRunCmd"`
	DistributeBinaryRunCmd string    `gorm:"column:distributeBinaryRunCmd"    json:"distributeBinaryRunCmd"`
	AggregateBinaryRunCmd  string    `gorm:"column:aggregateBinaryRunCmd"     json:"aggregateBinaryRunCmd"`
	OptionalFilesNames     string    `gorm:"column:optionalFilesNames"   json:"-"`
}

func (JobInfo) TableName() string {
	return "jobs.jobsinfo"
}

//optionalFilesNames is now a string, but it represents an array
func (ji *JobInfo) OptionalFilesStrToArr() []string {
	return strings.Split(ji.OptionalFilesNames, ",")
}

func OptionalFilesArrToStr(optionalFiles []string) string {
	return strings.Join(optionalFiles, ",")
}

type JobStatus string

const (
	IN_PROGRESS JobStatus = "IN_PROGRESS"
	DONE        JobStatus = "DONE"
)

func (dBWrapper *DBWrapper) GetAllJobsInfo(jobsInfo *[]JobInfo) *gorm.DB {
	return dBWrapper.Db.Raw("SELECT * FROM jobs.jobsinfo").Scan(jobsInfo)
}

func (dBWrapper *DBWrapper) GetLatestInProgressJobsInfo(jobsInfo *JobInfo, maxLateTime time.Time) *gorm.DB {
	return dBWrapper.Db.Raw(`
	SELECT * FROM jobs.jobsinfo 
	WHERE status = ? AND timeAssigned < ?
	ORDER BY timeAssigned ASC
	LIMIT 1
	`, IN_PROGRESS, maxLateTime).Scan(jobsInfo)
}

// check if the job is assigned to another master
func (dBWrapper *DBWrapper) CheckIsJobAssigned(jobsInfo *JobInfo, jobId string) *gorm.DB {
	return dBWrapper.Db.Raw(`
	SELECT * FROM jobs.jobsinfo 
	WHERE jobId = ?
	LIMIT 1
		`, jobId).Scan(jobsInfo)
}

// delete job by id
func (dBWrapper *DBWrapper) DeleteJobById(jobId string) *gorm.DB {
	return dBWrapper.Db.Delete(jobId)
}

// assign job to master
func (dBWrapper *DBWrapper) CreateJobsInfo(jobInfo *JobInfo, optionalFiles []string) *gorm.DB {
	jobInfo.OptionalFilesNames = OptionalFilesArrToStr(optionalFiles)
	return dBWrapper.Db.Create(jobInfo)
}
