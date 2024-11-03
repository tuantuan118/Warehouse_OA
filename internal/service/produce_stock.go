package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetProduceStockList(produce *models.ProduceStock,
	begReportingTime, endReportingTime string,
	pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.ProduceStock{})

	if produce.Name != "" {
		db = db.Where("name = ?", produce.Name)
	}
	if begReportingTime != "" && endReportingTime != "" {
		db = db.Where("add_time BETWEEN ? AND ?", begReportingTime, endReportingTime)
	}

	return Pagination(db, []models.ProduceStock{}, pn, pSize)
}

func GetProduceStockById(id int) (*models.ProduceStock, error) {
	db := global.Db.Model(&models.ProduceStock{})

	data := &models.ProduceStock{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveProduceStock(produce *models.ProduceStock) (*models.ProduceStock, error) {
	err := global.Db.Model(&models.ProduceStock{}).Create(produce).Error

	return produce, err
}

func UpdateProduceStock(produce *models.ProduceStock) (*models.ProduceStock, error) {
	if produce.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetProduceStockById(produce.ID)
	if err != nil {
		return nil, err
	}

	return produce, global.Db.Updates(&produce).Error
}

// GetProduceStockFieldList 获取字段列表
func GetProduceStockFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.ProduceStock{})
	switch field {
	case "name":
		db.Select("name")
	case "orderNumber":
		db.Select("order_number")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}
