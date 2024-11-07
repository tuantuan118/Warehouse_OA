package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetProduceList(produce *models.Produce,
	begReportingTime, endReportingTime string,
	begFinishTime, endFinishTime string,
	pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.Produce{})

	if produce.Name != "" {
		db = db.Where("name = ?", produce.Name)
	}
	if produce.Status > 0 {
		db = db.Where("status = ?", produce.Status)
	}
	if begReportingTime != "" && endReportingTime != "" {
		db = db.Where("add_time BETWEEN ? AND ?", begReportingTime, endReportingTime)
	}
	if begFinishTime != "" && endFinishTime != "" {
		db = db.Where("finish_time BETWEEN ? AND ?", begFinishTime, endFinishTime)
	}

	return Pagination(db, []models.Produce{}, pn, pSize)
}

func GetProduceById(id int) (*models.Produce, error) {
	db := global.Db.Model(&models.Produce{})

	data := &models.Produce{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveProduce(produce *models.Produce) (*models.Produce, error) {
	err := IfIngredientsByName(produce.Name)
	if err != nil {
		return nil, err
	}

	produceManage, err := GetProduceManageById(*produce.ProduceManageId)
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("20060102")
	total, err := getTodayOrderCount()
	if err != nil {
		return nil, err
	}
	produce.OrderNumber = fmt.Sprintf("SC%s%d", today, total+10000)
	produce.ProduceManage = produceManage
	produce.Name = produceManage.Name
	produce.Status = 1

	err = global.Db.Model(&models.Produce{}).Create(produce).Error

	return produce, err
}

func UpdateProduce(produce *models.Produce) (*models.Produce, error) {
	if produce.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetProduceById(produce.ID)
	if err != nil {
		return nil, err
	}

	produce.Name = ""
	produce.ProduceManageId = nil
	produce.ProduceManage = nil

	return produce, global.Db.Updates(&produce).Error
}

func DelProduce(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetProduceById(id)
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

// GetProduceFieldList 获取字段列表
func GetProduceFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Produce{})
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

func getTodayOrderCount() (int64, error) {
	today := time.Now().Format("2006-01-02")
	startOfDay, _ := time.Parse("2006-01-02", today)

	var total int64
	err := global.Db.Model(&models.Produce{}).Where(
		"add_time >= ?", startOfDay).Count(&total).Error

	return total, err
}
