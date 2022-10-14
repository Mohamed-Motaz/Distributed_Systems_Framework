package Database

import "gorm.io/gorm"

type Chat struct {
	Common
	Application_token string `gorm:"column:application_token"  json:"application_token"`
	Number            int32  `gorm:"column:number"             json:"number"`
	Messages_count    int32  `gorm:"column:messages_count"     json:"messages_count"`
}

func (Chat) TableName() string {
	return "instabug.chats"
}

func (db *DBWrapper) GetChatIdByAppTokenAndNumber(id *int, token string, chatNum int) *gorm.DB {
	return db.Db.Raw(`SELECT id FROM instabug.chats 
					  WHERE application_token = ?
					  AND number = ? LIMIT 1`, token, chatNum).Scan(id)
}
