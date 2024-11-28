package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetFinishedStockList(finished *models.FinishedStock,
	begReportingTime, endReportingTime string,
	pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.FinishedStock{})

	if finished.Name != "" {
		db = db.Where("name = ?", finished.Name)
	}
	if begReportingTime != "" && endReportingTime != "" {
		db = db.Where("add_time BETWEEN ? AND ?", begReportingTime, endReportingTime)
	}

	return Pagination(db, []models.FinishedStock{}, pn, pSize)
}

func GetFinishedStockById(id int) (*models.FinishedStock, error) {
	db := global.Db.Model(&models.FinishedStock{})

	data := &models.FinishedStock{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveFinishedStockByInBound(tx *gorm.DB, finished *models.Finished) error {
	var err error
	var total int64
	db := global.Db.Model(&models.FinishedStock{})

	db = db.Where("finished_manage_id = ?", finished.FinishedManageId)
	db.Count(&total)

	if total == 0 {
		_, err := SaveFinishedStock(&models.FinishedStock{
			BaseModel: models.BaseModel{
				Operator: finished.Operator,
				Remark:   finished.Remark,
			},
			Name:             finished.Name,
			Amount:           float64(finished.ActualAmount),
			FinishedManageId: finished.FinishedManageId,
		})
		return err
	}

	data := &models.FinishedStock{}
	err = db.First(&data).Error
	if err != nil {
		return err
	}
	data.Amount += float64(finished.ActualAmount)

	return tx.Updates(&data).Error
}

func SaveFinishedStock(finished *models.FinishedStock) (*models.FinishedStock, error) {
	err := global.Db.Model(&models.FinishedStock{}).Create(&finished).Error

	return finished, err
}
func UpdateFinishedStockNum(db *gorm.DB, id int, total int) error {
	if id == 0 {
		return errors.New("id is 0")
	}
	finished, err := GetFinishedStockById(id)
	if err != nil {
		return err
	}
	if finished.Amount+float64(total) < 0 {
		return errors.New("stock not enough")
	}

	finished.Amount += float64(total)

	return db.Updates(&finished).Error
}

// GetFinishedStockFieldList 获取字段列表
func GetFinishedStockFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.FinishedStock{})
	switch field {
	case "name":
		db = db.Select("name")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}
