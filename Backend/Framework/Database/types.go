package Database

import (
	"gorm.io/gorm"
)

type DBWrapper struct {
	Db *gorm.DB
}
