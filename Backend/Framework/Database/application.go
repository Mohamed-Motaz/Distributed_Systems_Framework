package Database

import (
	"gorm.io/gorm"
)

type Application struct {
	Common
	Token       string `gorm:"column:token"            json:"token"`
	Name        string `gorm:"column:name"             json:"name"`
	Chats_count int32  `gorm:"column:chats_count"      json:"chats_count"`
}

func (Application) TableName() string {
	return "instabug.applications"
}

func (db *DBWrapper) GetApplicationByToken(a *Application, token string) *gorm.DB {
	return db.Db.Raw("SELECT * FROM instabug.applications WHERE token = ? LIMIT 1", token).Scan(a)
}
