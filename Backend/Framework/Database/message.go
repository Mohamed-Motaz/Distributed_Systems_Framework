package Database

import "gorm.io/gorm"

type Message struct {
	Common
	Chat_id int32  `gorm:"column:chat_id"       json:"chat_id"`
	Number  int32  `gorm:"column:number"        json:"number"`
	Body    string `gorm:"column:body"          json:"body"`
}

func (Message) TableName() string {
	return "instabug.messages"
}

func (db *DBWrapper) GetMessagedByChatIdAndNumber(id *int, chatId int, messageNum int) *gorm.DB {
	return db.Db.Raw(`SELECT id FROM instabug.messages 
					  WHERE chat_id = ?
					  AND number = ? LIMIT 1`, chatId, messageNum).Scan(id)
}
