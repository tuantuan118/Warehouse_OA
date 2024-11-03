package models

type ProduceStock struct {
	BaseModel
	Name               string `gorm:"type:varchar(256);not null" json:"name"`
	Ratio              string `gorm:"type:varchar(256);not null" json:"ratio"`
	Amount             int    `gorm:"type:int(11);not null" json:"amount"`
	ProductIngredients string `gorm:"type:Text;not null" json:"productIngredients"`
}
