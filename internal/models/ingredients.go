package models

type Ingredients struct {
	BaseModel
	Name     string `gorm:"type:varchar(256);not null" json:"name"`
	Supplier string `gorm:"type:varchar(256);not null" json:"supplier"`
}
