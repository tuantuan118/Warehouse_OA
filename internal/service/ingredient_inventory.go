package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetInventoryList(name, supplier, stockUser, begTime, endTime string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.IngredientInventory{})

	if name != "" {
		db = db.Where("name = ?", name)
	}
	if supplier != "" {
		db = db.Where("supplier = ?", supplier)
	}
	if stockUser != "" {
		db = db.Where("stock_user = ?", stockUser)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("created_at BETWEEN ? AND ?", begTime, endTime)
	}
	db = db.Preload("Ingredient")

	return Pagination(db, []models.IngredientInventory{}, pn, pSize)
}

func GetInventoryById(id int) (*models.IngredientInventory, error) {
	db := global.Db.Model(&models.IngredientInventory{})

	data := &models.IngredientInventory{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveInventoryByInBound(db *gorm.DB, inBound *models.IngredientInBound) error {
	data := &models.IngredientInventory{}
	var total int64

	db = db.Model(&models.IngredientInventory{})
	db = db.Where("ingredient_id = ?", *inBound.IngredientID)
	if inBound.Specification != "" {
		db = db.Where("specification = ?", inBound.Specification)
	}
	if inBound.StockUnit != "" {
		db = db.Where("stock_unit = ?", inBound.StockUnit)
	}

	var err error
	err = db.Count(&total).Error
	if err != nil {
		return err
	}
	if total == 0 {
		_, err = SaveInventory(&models.IngredientInventory{
			BaseModel: models.BaseModel{
				Operator: inBound.Operator,
			},
			IngredientID:  inBound.IngredientID,
			Ingredient:    inBound.Ingredient,
			Specification: inBound.Specification,
			TotalPrice:    inBound.TotalPrice,
			StockNum:      inBound.StockNum,
			StockUnit:     inBound.StockUnit,
		})
		return err
	}

	err = db.First(&data).Error
	if err != nil {
		return err
	}
	data.StockNum += inBound.StockNum
	data.TotalPrice += inBound.TotalPrice
	return global.Db.Updates(&data).Error
}

func SaveInventory(inventory *models.IngredientInventory) (*models.IngredientInventory, error) {
	ingredients, err := GetIngredientsById(*inventory.IngredientID)
	if err != nil {
		return nil, err
	}

	inventory.Ingredient = ingredients

	err = global.Db.Model(&models.IngredientInventory{}).Create(&inventory).Error

	return inventory, err
}

func UpdateInventoryByInBound(db *gorm.DB, oldInBound *models.IngredientInBound) error {
	data := &models.IngredientInventory{}
	var total int64

	db = db.Model(&models.IngredientInventory{})
	db = db.Where("ingredient_id = ?", *oldInBound.IngredientID)
	if oldInBound.Specification != "" {
		db = db.Where("specification = ?", oldInBound.Specification)
	}
	if oldInBound.StockUnit != "" {
		db = db.Where("stock_unit = ?", oldInBound.StockUnit)
	}

	var err error
	err = db.Count(&total).Error
	if err != nil {
		return err
	}
	if total == 0 {
		return errors.New("data does not exist")
	}
	err = db.First(&data).Error
	if err != nil {
		return err
	}
	data.StockNum -= oldInBound.StockNum
	data.TotalPrice -= oldInBound.TotalPrice
	return global.Db.Updates(&data).Error
}

func UpdateStockNum(db *gorm.DB, id int, total int) error {
	inventory, err := GetInventoryById(id)
	if err != nil {
		return err
	}
	if inventory.StockNum+total < 0 {
		return errors.New("stock not enough")
	}

	inventory.StockNum += total

	return db.Updates(&inventory).Error
}

// GetInventoryFieldList 获取字段列表
func GetInventoryFieldList(field string) (map[string]string, error) {
	db := global.Db.Model(&models.IngredientInventory{})
	db = db.Select("id")
	switch field {
	case "name":
		db = db.Select("name")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make(map[string]string)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}
