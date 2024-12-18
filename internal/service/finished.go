package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetFinishedList(finished *models.Finished,
	begTime, endTime string,
	pn, pSize int, b bool) (interface{}, error) {
	db := global.Db.Model(&models.Finished{})
	db.Preload("FinishedManage")

	if finished.FinishedManageId > 0 {
		db = db.Where("finished_manage_id = ?", finished.FinishedManageId)
	}
	if finished.Name != "" {
		slice := strings.Split(finished.Name, ";")
		db = db.Where("name in ?", slice)
	}
	if finished.Status > 0 {
		db = db.Where("status = ?", finished.Status)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(finish_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}
	if b {
		db = db.Where("in_and_out = ?", b)
	} else {
		db = db.Where("status != ?", 1)
	}

	return Pagination(db, []models.Finished{}, pn, pSize)
}

func GetOutFinishedList(finished *models.Finished,
	begTime, endTime string,
	pn, pSize int) (interface{}, error) {

	var finishedManageId int
	if finished.ID != 0 {
		stock, err := GetFinishedStockById(finished.ID)
		if err != nil {
			return nil, err
		}
		finishedManageId = stock.FinishedManageId
	}

	return GetFinishedList(&models.Finished{
		Name:             finished.Name,
		Status:           finished.Status,
		FinishedManageId: finishedManageId,
	}, begTime, endTime, pn, pSize, false)
}

func GetFinishedById(id int) (*models.Finished, error) {
	db := global.Db.Model(&models.Finished{})

	data := &models.Finished{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveFinished(finished *models.Finished) (*models.Finished, error) {
	err := IfIngredientsByName(finished.Name)
	if err != nil {
		return nil, err
	}

	finishedManage, err := GetFinishedManageById(finished.FinishedManageId)
	if err != nil {
		return nil, err
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

	finished.FinishedManage = finishedManage
	finished.Name = finishedManage.Name
	finished.Status = 1
	finished.FinishTime = nil
	finished.InAndOut = true
	finished.OperationType = "入库"

	if finished.FinishHour > 0 {
		et := time.Now().Add(time.Duration(finished.FinishHour) * time.Hour)
		finished.EstimatedTime = &et

	} else {
		finished.Status = 4
		et := time.Now()
		finished.EstimatedTime = &et
	}

	err = tx.Model(&models.Finished{}).Create(finished).Error
	if err != nil {
		return nil, err
	}

	// 扣除配料库存
	for _, material := range finishedManage.Material {
		err = FinishedSaveInBound(tx, &models.IngredientInBound{
			BaseModel: models.BaseModel{
				Operator: finished.Operator,
			},
			IngredientID:     material.IngredientInventory.IngredientID,
			StockNum:         0 - (material.Quantity * float64(finished.ExpectAmount)),
			StockUnit:        material.IngredientInventory.StockUnit,
			StockUser:        finished.Operator,
			StockTime:        time.Now(),
			OperationType:    "出库",
			OperationDetails: fmt.Sprintf("报工生产【%s】", finished.Name),
		})
		if err != nil {
			return nil, err
		}
	}

	return finished, err
}

func UpdateFinished(finished *models.Finished) (*models.Finished, error) {
	if finished.ID == 0 {
		return nil, errors.New("id is 0")
	}
	oldData, err := GetFinishedById(finished.ID)
	if err != nil {
		return nil, err
	}

	if oldData.Status == 2 || oldData.Status == 3 {
		return nil, errors.New("finished has been finished, can not update")
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

	finished.ExpectAmount = 0
	finished.Name = ""
	finished.FinishedManage = nil

	err = tx.Updates(&finished).Error
	if err != nil {
		return nil, err
	}

	return finished, err
}

func VoidFinished(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetFinishedById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	if data.Status == 2 || data.Status == 3 {
		return errors.New("finished has been finished, can not update")
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

	// 扣除配料库存
	finishedManage, err := GetFinishedManageById(data.FinishedManageId)
	if err != nil {
		return err
	}
	for _, material := range finishedManage.Material {
		err = FinishedSaveInBound(tx, &models.IngredientInBound{
			BaseModel: models.BaseModel{
				Operator: username,
			},
			IngredientID:     material.IngredientInventory.IngredientID,
			StockNum:         material.Quantity * float64(data.ExpectAmount),
			StockUnit:        material.IngredientInventory.StockUnit,
			StockUser:        username,
			StockTime:        time.Now(),
			OperationType:    "入库",
			OperationDetails: fmt.Sprintf("报工生产【%s】作废重新入库", data.Name),
		})
		if err != nil {
			return err
		}
	}

	data.Operator = username
	data.Status = 3

	return tx.Updates(&data).Error
}

func FinishFinished(id, amount int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetFinishedById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	if data.Status == 2 || data.Status == 3 {
		return errors.New("finished has been finished, can not update")
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

	data.Operator = username
	data.Status = 2
	data.ActualAmount = amount
	data.Ratio = (float64(data.ActualAmount) / float64(data.ExpectAmount)) * float64(100)
	ft := time.Now()
	data.FinishTime = &ft
	data.OperationDetails = fmt.Sprintf("生产完工")

	err = SaveFinishedStockByInBound(tx, data)
	if err != nil {
		return err
	}

	return tx.Updates(&data).Error
}

func DelFinished(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetFinishedById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	if data.Status == 2 || data.Status == 3 {
		return errors.New("finished has been finished, can not update")
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

	// 扣除配料库存
	finishedManage, err := GetFinishedManageById(data.FinishedManageId)
	if err != nil {
		return err
	}
	for _, material := range finishedManage.Material {
		err = FinishedSaveInBound(tx, &models.IngredientInBound{
			BaseModel: models.BaseModel{
				Operator: username,
			},
			IngredientID:     material.IngredientInventory.IngredientID,
			StockNum:         material.Quantity * float64(data.ExpectAmount),
			StockUnit:        material.IngredientInventory.StockUnit,
			StockUser:        username,
			StockTime:        time.Now(),
			OperationType:    "入库",
			OperationDetails: fmt.Sprintf("报工生产【%s】删除重新入库", data.Name),
		})
		if err != nil {
			return err
		}
	}

	return tx.Delete(&data).Error
}

// GetFinishedFieldList 获取字段列表
func GetFinishedFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Finished{})
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

func GetFinishedByStatus(id, status int) (int64, error) {
	var total int64
	db := global.Db.Model(&models.Finished{})
	db = db.Where("finished_manage_id = ?", id)
	db = db.Where("status = ?", status)
	err := db.Count(&total).Error

	return total, err
}

func ProductSaveFinished(tx *gorm.DB, finished *models.Finished) error {
	manage, err := GetFinishedManageById(finished.FinishedManageId)
	if err != nil {
		return err
	}

	finished.FinishedManage = manage

	err = SaveFinishedStockByInBound(tx, finished)
	if err != nil {
		return err
	}
	err = tx.Model(&models.Finished{}).Create(finished).Error

	return err
}
