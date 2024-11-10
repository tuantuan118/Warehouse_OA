package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
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

	produceManage, err := GetProduceManageById(produce.ProduceManageId)
	if err != nil {
		return nil, err
	}

	if produce.FinishHour <= 0 {
		return nil, errors.New("finish hour is invalid")
	}

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	today := time.Now().Format("20060102")
	total, err := getTodayProduceCount()
	if err != nil {
		return nil, err
	}

	produce.OrderNumber = fmt.Sprintf("QY%s%d", today, total+10001)
	produce.ProduceManage = produceManage
	produce.Name = produceManage.Name
	produce.Status = 1
	produce.FinishTime = time.Now().Add(time.Duration(produce.FinishHour) * time.Hour)

	err = tx.Model(&models.Produce{}).Create(produce).Error
	if err != nil {
		return nil, err
	}
	// 扣除配料库存
	err = UpdateIngredientStock(tx, produce.ProduceManageId, produce.Amount, true)

	return produce, err
}

func UpdateProduce(produce *models.Produce) (*models.Produce, error) {
	if produce.ID == 0 {
		return nil, errors.New("id is 0")
	}
	oldData, err := GetProduceById(produce.ID)
	if err != nil {
		return nil, err
	}

	if oldData.Status == 2 || oldData.Status == 3 {
		return nil, errors.New("produce has been finished, can not update")
	}

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if oldData.Amount != produce.Amount {
		margin := oldData.Amount - produce.Amount

		logrus.Infoln(produce)
		err = UpdateIngredientStock(tx, oldData.ProduceManageId, margin, false)
		if err != nil {
			return nil, err
		}
	}

	produce.OrderNumber = ""
	produce.Name = ""
	produce.ProduceManage = nil

	err = tx.Updates(&produce).Error
	if err != nil {
		return nil, err
	}

	return produce, err
}

func VoidProduce(id int, username string) error {
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

	if data.Status == 2 || data.Status == 3 {
		return errors.New("produce has been finished, can not update")
	}

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = UpdateIngredientStock(tx, data.ProduceManageId, data.Amount, false)
	if err != nil {
		return err
	}

	data.Operator = username
	data.Status = 3

	return tx.Updates(&data).Error
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

	if data.Status == 2 || data.Status == 3 {
		return errors.New("produce has been finished, can not update")
	}

	data.Operator = username
	data.IsDeleted = true
	err = global.Db.Updates(&data).Error
	if err != nil {
		return err
	}

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = UpdateIngredientStock(tx, data.ProduceManageId, data.Amount, false)
	if err != nil {
		return err
	}

	return tx.Delete(&data).Error
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

func getTodayProduceCount() (int64, error) {
	today := time.Now().Format("2006-01-02")
	startOfDay, _ := time.Parse("2006-01-02", today)

	var total int64
	err := global.Db.Model(&models.Produce{}).Where(
		"add_time >= ?", startOfDay).Count(&total).Error

	return total, err
}

func GetProduceByStatus(id, status int) (int64, error) {
	var total int64
	db := global.Db.Model(&models.Produce{})
	db = db.Where("produce_manage_id = ?", id)
	db = db.Where("status = ?", status)
	err := db.Count(&total).Error

	return total, err
}
