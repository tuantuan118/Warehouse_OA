package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetInBoundList(name, supplier, stockUser, begTime, endTime string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.IngredientInBound{})

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

	return Pagination(db, []models.IngredientInBound{}, pn, pSize)
}

func GetInBoundById(id int) (*models.IngredientInBound, error) {
	db := global.Db.Model(&models.IngredientInBound{})

	data := &models.IngredientInBound{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveInBound(inBound *models.IngredientInBound) (*models.IngredientInBound, error) {
	ingredients, err := GetIngredientsById(*inBound.IngredientID)
	if err != nil {
		return nil, err
	}

	inBound.Ingredient = ingredients

	err = global.Db.Model(&models.IngredientInBound{}).Create(inBound).Error

	return inBound, err
}

func UpdateInBound(inBound *models.IngredientInBound) (*models.IngredientInBound, error) {
	if inBound.ID == 0 {
		return nil, errors.New("id is 0")
	}
	data, err := GetInBoundById(inBound.ID)
	if err != nil {
		return nil, err
	}

	if data.IngredientID != inBound.IngredientID {
		ingredients, err := GetIngredientsById(*inBound.IngredientID)
		if err != nil {
			return nil, err
		}

		inBound.Ingredient = ingredients
	}

	return inBound, global.Db.Updates(&inBound).Error
}

func DelInBound(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetInBoundById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	data.Operator = username
	data.IsDeleted = true
	err = global.Db.Updates(&data).Error
	if err != nil {
		return err
	}

	return global.Db.Delete(&data).Error
}

// GetInBoundFieldList 获取字段列表
func GetInBoundFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.IngredientInBound{})
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