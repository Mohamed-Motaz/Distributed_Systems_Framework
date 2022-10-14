package Database

import "time"

type Common struct {
	Id         int       `gorm:"primaryKey; column:id"   json:"-"`
	Created_at time.Time `gorm:"column:created_at"       json:"-"`
	Updated_at time.Time `gorm:"column:updated_at"       json:"-"`
}

func MakeNewCommon() Common {
	now := time.Now()
	return Common{
		Id:         0,
		Created_at: now,
		Updated_at: now,
	}
}
