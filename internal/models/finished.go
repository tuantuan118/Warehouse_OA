package models

import "time"

type Finished struct {
	BaseModel
	Name             string          `gorm:"type:varchar(256);not null" json:"name"`
	Ratio            float64         `gorm:"type:decimal(10,2);not null" json:"ratio"`
	ExpectAmount     int             `gorm:"type:int(11);not null" json:"expectAmount"`
	ActualAmount     int             `gorm:"type:int(11);not null" json:"actualAmount"`
	Status           int             `gorm:"type:int(11);not null" json:"status"`
	FinishHour       int             `gorm:"-" json:"finishHour"`
	FinishTime       time.Time       `gorm:"type:Time" json:"finishTime"`
	FinishedManageId int             `gorm:"type:int(11)" json:"finishedManageId"`
	FinishedManage   *FinishedManage `gorm:"foreignKey:FinishedManageId;" json:"finishedManage"`
}
