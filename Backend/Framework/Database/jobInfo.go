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
	CreatedAt          time.Time `gorm:"column:createdAt"            json:"createdAt"`
	TimeAssigned       time.Time `gorm:"column:timeAssigned"         json:"timeAssigned"`
	Status             JobStatus `gorm:"column:status"               json:"status"`
	ProcessBinaryId    string    `gorm:"column:processBinaryId"      json:"processBinaryId"`
	DistributeBinaryId string    `gorm:"column:distributeBinaryId"   json:"distributeBinaryId"`
	AggregateBinaryId  string    `gorm:"column:aggregateBinaryId"    json:"aggregateBinaryId"`
}

// todo: change binaries to be ids here and in the migration
func (JobInfo) TableName() string {
	return "jobs.jobsinfo"
}

type JobStatus string

const (
	IN_PROGRESS JobStatus = "IN_PROGRESS"
	DONE        JobStatus = "DONE"
)

func (dBWrapper *DBWrapper) GetAllJobsInfo(jobsInfo *[]JobInfo) *gorm.DB {
	return dBWrapper.Db.Raw("SELECT * FROM jobs.jobsinfo").Scan(jobsInfo)
}

func (dBWrapper *DBWrapper) GetLatestInProgressJobsInfo(jobsInfo *[]JobInfo, maxLateTime time.Time) *gorm.DB {
	return dBWrapper.Db.Raw(`
	SELECT * FROM jobs.jobsinfo 
	WHERE status = ? AND timeAssigned < ?
	ORDER BY timeAssigned ASC
	`, IN_PROGRESS, maxLateTime).Scan(jobsInfo)
}

// check if the job is assigned to another master
func (dBWrapper *DBWrapper) GetJobByJobId(jobsInfo *JobInfo, jobId string) *gorm.DB {
	return dBWrapper.Db.Raw(`
	SELECT * FROM jobs.jobsinfo 
	WHERE jobId = ?
	LIMIT 1
		`, jobId).Scan(jobsInfo)
}

// delete job by id
func (dBWrapper *DBWrapper) DeleteJobById(id int) *gorm.DB {
	return dBWrapper.Db.Delete(id)
}

func (db DBWrapper) DeleteJobByJobId(jobId string) *gorm.DB {
	return db.Db.Where("jobs.jobsinfo.jobId = ?", jobId).Delete(JobInfo{})
}

// assign job to master
func (dBWrapper *DBWrapper) CreateJobsInfo(jobInfo *JobInfo) *gorm.DB {
	return dBWrapper.Db.Create(jobInfo)
}
