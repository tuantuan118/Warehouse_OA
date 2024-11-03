package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetProduceManageList(name string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.ProduceManage{})

	if name != "" {
		db = db.Where("name = ?", name)
	}

	return Pagination(db, []models.ProduceManage{}, pn, pSize)
}

func GetProduceManageById(id int) (*models.ProduceManage, error) {
	db := global.Db.Model(&models.ProduceManage{})

	data := &models.ProduceManage{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveProduceManage(produceManage *models.ProduceManage) (*models.ProduceManage, error) {
	if produceManage.Ingredients == nil || len(produceManage.Ingredients) == 0 {
		return nil, errors.New("ingredients is empty")
	}

	ingredientList := make([]models.IngredientInventory, 0)
	for _, ingredients := range produceManage.Ingredients {
		inventory, err := GetInventoryById(ingredients.ID)
		if err != nil {
			return nil, err
		}
		ingredientList = append(ingredientList, *inventory)
	}

	produceManage.Ingredients = ingredientList

	return produceManage, global.Db.Model(&models.ProduceManage{}).Create(produceManage).Error
}

func UpdateProduceManage(produceManage *models.ProduceManage) (*models.ProduceManage, error) {
	if produceManage.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetProduceManageById(produceManage.ID)
	if err != nil {
		return nil, err
	}

	if produceManage.Ingredients == nil || len(produceManage.Ingredients) == 0 {
		return nil, errors.New("ingredients is empty")
	}

	ingredientList := make([]models.IngredientInventory, 0)
	for _, ingredients := range produceManage.Ingredients {
		inventory, err := GetInventoryById(ingredients.ID)
		if err != nil {
			return nil, err
		}
		ingredientList = append(ingredientList, *inventory)
	}

	produceManage.Ingredients = ingredientList

	return produceManage, global.Db.Updates(&produceManage).Error
}

func DelProduceManage(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetProduceManageById(id)
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

// GetProduceManageFieldList 获取字段列表
func GetProduceManageFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.ProduceManage{})
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
