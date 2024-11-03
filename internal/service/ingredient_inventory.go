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

func SaveInventory(inventory *models.IngredientInventory) (*models.IngredientInventory, error) {
	ingredients, err := GetIngredientsById(*inventory.IngredientID)
	if err != nil {
		return nil, err
	}

	inventory.Ingredient = ingredients

	err = global.Db.Model(&models.IngredientInventory{}).Create(inventory).Error

	return inventory, err
}

func UpdateInventory(inventory *models.IngredientInventory) (*models.IngredientInventory, error) {
	if inventory.ID == 0 {
		return nil, errors.New("id is 0")
	}
	data, err := GetInventoryById(inventory.ID)
	if err != nil {
		return nil, err
	}

	if data.IngredientID != inventory.IngredientID {
		ingredients, err := GetIngredientsById(*inventory.IngredientID)
		if err != nil {
			return nil, err
		}

		inventory.Ingredient = ingredients
	}

	return inventory, global.Db.Updates(&inventory).Error
}

// GetInventoryFieldList 获取字段列表
func GetInventoryFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.IngredientInventory{})
	switch field {
	case "name":
		db.Select("name")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}
