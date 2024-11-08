package models

type ProduceStock struct {
	BaseModel
	Name               string         `gorm:"type:varchar(256);not null" json:"name"`
	Amount             float64        `gorm:"type:decimal(10,2);not null" json:"amount"`
	ProductIngredients string         `gorm:"type:Text;not null" json:"productIngredients"`
	ProduceManageId    int            `gorm:"type:int(11)" json:"produceManageId"`
	ProduceManage      *ProduceManage `gorm:"foreignKey:ProduceManageId;" json:"produceManage"`
}
