package Database

import (
	"gorm.io/gorm"
)

type RunnableFiles struct {
	Id           int    `gorm:"primaryKey; column:id"                   json:"id"`
	BinaryName   string `gorm:"primaryKey; column:binaryName"           json:"binaryName"`
	BinaryType   string `gorm:"column:binaryType"                       json:"binaryType"`
	BinaryRunCmd string `gorm:"primaryKey; column:binaryRunCmd"         json:"binaryRunCmd"`
}

func (RunnableFiles) TableName() string {
	return "jobs.runnableFiles"
}

func (dBWrapper *DBWrapper) CreateRunnableFile(runnableFile *RunnableFiles) *gorm.DB {
	return dBWrapper.Db.Create(runnableFile)
}

func (dBWrapper *DBWrapper) GetRunCmdOfBinary(runnableFile *RunnableFiles, binaryName string, binaryType string) *gorm.DB {
	return dBWrapper.Db.Raw(`
	SELECT * FROM jobs.runnableFiles
	WHERE BinaryName = ? AND BinaryType = ?
	`, binaryName, binaryType).Scan(runnableFile)
}
