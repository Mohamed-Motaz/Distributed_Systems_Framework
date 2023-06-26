package Database

import (
	"gorm.io/gorm"
)

type RunnableFiles struct {
	Id           int    `gorm:"primaryKey; column:id"       json:"id"`
	BinaryName   string `gorm:"column:binaryName"           json:"binaryName"`
	BinaryType   string `gorm:"column:binaryType"           json:"binaryType"`
	BinaryRunCmd string `gorm:"column:binaryRunCmd"         json:"binaryRunCmd"`
}

func (RunnableFiles) TableName() string {
	return "jobs.runnableFiles"
}

func (dBWrapper *DBWrapper) CreateRunnableFile(runnableFile *RunnableFiles) *gorm.DB {
	return dBWrapper.Db.Create(runnableFile)
}

func (dBWrapper *DBWrapper) GetBinaryByNameAndType(runnableFile *RunnableFiles, binaryName string, binaryType string) *gorm.DB {
	return dBWrapper.Db.Raw(`
	SELECT * FROM jobs.runnableFiles
	WHERE binaryName = ? AND binaryType = ?
	`, binaryName, binaryType).Scan(runnableFile)
}

func (dBWrapper *DBWrapper) GetRunnableFiles(runnableFile *RunnableFiles) *gorm.DB {
	return dBWrapper.Db.Raw(`
	SELECT * FROM jobs.runnableFiles
	`).Scan(runnableFile)
}
func (dBWrapper DBWrapper) DeleteRunnableFile(fileName, fileType string) *gorm.DB {
	return dBWrapper.Db.Where("jobs.runnableFiles.binaryName = ? AND jobs.runnableFiles.binaryType = ?", fileName, fileType).Delete(RunnableFiles{})
}
