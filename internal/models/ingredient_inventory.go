package models

type IngredientInventory struct {
	BaseModel
	IngredientID  *int         `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient    *Ingredients `gorm:"foreignKey:IngredientID" json:"ingredient"`
	Specification string       `gorm:"type:varchar(256)" json:"specification"`
	Price         float64      `gorm:"type:decimal(12,2)" json:"price"`
	StockNum      int          `gorm:"type:int(11)" json:"stockNum"`
	StockUnit     string       `gorm:"type:varchar(256)" json:"stockUnit"`
}
