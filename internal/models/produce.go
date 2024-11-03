package models

import "time"

type Produce struct {
	BaseModel
	Name            string         `gorm:"type:varchar(256);not null" json:"name"`
	OrderNumber     string         `gorm:"type:varchar(256);not null" json:"orderNumber"`
	Ratio           string         `gorm:"type:varchar(256);not null" json:"ratio"`
	Amount          int            `gorm:"type:int(11);not null" json:"amount"`
	Status          int            `gorm:"type:int(11);not null" json:"status"`
	FinishTime      time.Time      `gorm:"type:Time;not null" json:"finishTime"`
	ProduceManageId *int           `gorm:"type:int(11)" json:"produceManageId"`
	ProduceManage   *ProduceManage `gorm:"foreignKey:ProduceManageId;" json:"produceManage"`
}
