package models

import "time"

type IngredientInBound struct {
	BaseModel
	IngredientID     *int         `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient       *Ingredients `gorm:"foreignKey:IngredientID" json:"ingredient"`
	Specification    string       `gorm:"type:varchar(256)" json:"specification"`
	Price            float64      `gorm:"type:decimal(12,2)" json:"price"`
	TotalPrice       float64      `gorm:"type:decimal(12,2)" json:"totalPrice"`
	StockNum         int          `gorm:"type:int(11)" json:"stockNum"`
	StockUnit        int          `gorm:"type:int(2)" json:"stockUnit"`
	StockUser        string       `gorm:"type:varchar(256)" json:"stockUser"`
	StockTime        time.Time    `gorm:"type:Time" json:"stockTime"`
	InAndOut         bool         `gorm:"type:tinyint(1)" json:"inAndOut"` // InAndOut True 入库 False 出库
	OperationType    string       `gorm:"type:varchar(256)" json:"operationType"`
	OperationDetails string       `gorm:"type:varchar(256)" json:"operationDetails"`
}
