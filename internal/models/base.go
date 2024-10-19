package models

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        int       `gorm:"primaryKey"`
	Operator  string    `gorm:"type:varchar(100)"`
	Remark    string    `gorm:"type:varchar(256)"`
	CreatedAt time.Time `gorm:"column:add_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt
	IsDeleted bool `gorm:"column:is_delete"`
}
