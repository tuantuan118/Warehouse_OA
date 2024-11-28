package models

type Product struct {
	BaseModel
	OrderNumber      string          `gorm:"type:varchar(256);not null" json:"orderNumber"`
	Name             string          `gorm:"type:varchar(256)" json:"name"`
	FinishedManageId int             `gorm:"type:int(11)" json:"finishedManageId"`
	FinishedManage   *FinishedManage `gorm:"foreignKey:FinishedManageId;" json:"finishedManage"`
	Amount           int             `gorm:"type:int(11)" json:"amount"`
	Unit             int             `gorm:"type:int(2)" json:"unit"`
	Weight           int             `gorm:"type:int(11)" json:"weight"`
}
