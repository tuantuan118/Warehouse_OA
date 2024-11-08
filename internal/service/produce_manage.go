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

func GetProduceManageIngredients(id int) ([]map[string]interface{}, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}
	db := global.Db
	productIngredient := make([]models.ProductMaterial, 0)

	err := db.Where("produce_manage_id = ?", id).Preload(
		"IngredientInventory.Ingredient").Find(&productIngredient).Error
	if err != nil {
		return nil, err
	}

	requestData := make([]map[string]interface{}, 0)
	for _, v := range productIngredient {
		ingredient, err := GetIngredientsById(*v.IngredientInventory.IngredientID)
		if err != nil {
			return nil, err
		}
		requestData = append(requestData, map[string]interface{}{
			"inventory_id": v.IngredientInventory.ID,
			"name":         ingredient.Name,
			"quantity":     v.Quantity,
			"stockUnit":    v.IngredientInventory.StockUnit,
			"isCalculate":  v.IngredientInventory.Ingredient.IsCalculate,
		})
	}

	return requestData, err
}

func GetProduceManageById(id int) (*models.ProduceManage, error) {
	db := global.Db.Model(&models.ProduceManage{})

	data := &models.ProduceManage{}
	err := db.Where("id = ?", id).Preload("Material").First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveProduceManage(produceManage *models.ProduceManage) (*models.ProduceManage, error) {
	if produceManage.Material == nil || len(produceManage.Material) == 0 {
		return nil, errors.New("ingredients is empty")
	}
	var err error
	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	for _, material := range produceManage.Material {
		inventory := new(models.IngredientInventory)
		inventory, err = GetInventoryById(material.IngredientID)
		if err != nil {
			return nil, err
		}
		material.IngredientInventory = inventory
	}

	err = global.Db.Model(&models.ProduceManage{}).Create(&produceManage).Error
	if err != nil {
		return nil, err
	}

	return produceManage, err
}

func UpdateProduceManage(produceManage *models.ProduceManage) (*models.ProduceManage, error) {
	if produceManage.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetProduceManageById(produceManage.ID)
	if err != nil {
		return nil, err
	}

	total, err := GetProduceByStatus(1)
	if err != nil {
		return nil, err
	}
	if total > 0 {
		return nil, errors.New("exist produce, can not update")
	}

	if produceManage.Material == nil || len(produceManage.Material) == 0 {
		return nil, errors.New("ingredients is empty")
	}

	err = RemoveIngredients(produceManage.ID)
	if err != nil {
		return nil, err
	}

	for _, material := range produceManage.Material {
		inventory := new(models.IngredientInventory)
		inventory, err = GetInventoryById(material.IngredientID)
		if err != nil {
			return nil, err
		}
		material.IngredientInventory = inventory
	}

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

	total, err := GetProduceByStatus(1)
	if err != nil {
		return err
	}
	if total > 0 {
		return errors.New("exist produce, can not delete")
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

// UpdateIngredientStock 修改配料库存, b 时候为 true, 减少库存, b 时候为 false, 增加库存
func UpdateIngredientStock(db *gorm.DB, id, amount int, b bool) error {
	produceManage, err := GetProduceManageById(id)
	if err != nil {
		return err
	}

	for _, material := range produceManage.Material {
		if b {
			material.Quantity = 0 - material.Quantity
		}
		err = UpdateStockNum(db, material.IngredientID, amount*material.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveIngredients(manageId int) error {
	return global.Db.Model(&models.ProductMaterial{}).Where(
		"produce_manage_id = ?", manageId).Delete(&models.ProductMaterial{}).Error

}
