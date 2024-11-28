package models

type FinishedManage struct {
	BaseModel
	Name     string            `gorm:"type:varchar(256);not null" json:"name"`
	Material []ProductMaterial `gorm:"foreignKey:FinishedManageID;references:ID" json:"material"`
}

type ProductMaterial struct {
	FinishedManageID    int                  `gorm:"primaryKey;index" json:"finished_manage_id"`
	IngredientID        int                  `gorm:"primaryKey;" json:"ingredient_id"`
	IngredientInventory *IngredientInventory `gorm:"foreignKey:IngredientID" json:"ingredient_inventory"`
	Quantity            int                  `gorm:"type:int(11);not null" json:"quantity"` // 用量
}
