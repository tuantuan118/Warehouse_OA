package models

type ProduceManage struct {
	BaseModel
	Name        string                `gorm:"type:varchar(256);not null" json:"name"`
	Ingredients []IngredientInventory `gorm:"many2many:manage_ingredient;" json:"roles"`
}
