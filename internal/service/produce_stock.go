package service

import (
	"errors"
	"github.com/sirupsen/logrus"
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

func SaveProduceStockByInBound(produce *models.Produce) error {
	var total int64
	db := global.Db.Model(&models.ProduceStock{})

	db = db.Where("produce_manage_id = ?", produce.ProduceManageId)
	db.Count(&total)

	quantity, err := GetProduceQuantity(produce.ProduceManageId)
	if err != nil {
		return err
	}
	amount := float64(quantity) * float64(produce.Amount) * produce.Ratio / 100

	logrus.Infoln(*produce)

	if total == 0 {
		_, err = SaveProduceStock(&models.ProduceStock{
			BaseModel: models.BaseModel{
				Operator: produce.Operator,
				Remark:   produce.Remark,
			},
			Name:            produce.Name,
			Amount:          amount,
			ProduceManageId: produce.ProduceManageId,
		})
		return err
	}

	data := &models.ProduceStock{}
	err = db.First(&data).Error
	if err != nil {
		return err
	}
	data.Amount += amount

	return db.Updates(&data).Error
}

func GetProduceQuantity(produceManageId int) (int, error) {
	dataCount, err := GetProduceManageIngredients(produceManageId)
	if err != nil {
		return 0, err
	}
	var quantity int
	for _, v := range dataCount {
		if v["isCalculate"].(bool) {
			quantity += v["quantity"].(int)
		}
	}
	return quantity, nil
}

func SaveProduceStock(produce *models.ProduceStock) (*models.ProduceStock, error) {
	err := global.Db.Model(&models.ProduceStock{}).Create(&produce).Error

	return produce, err
}
func UpdateProduceStockNum(db *gorm.DB, id int, total int) error {
	if id == 0 {
		return errors.New("id is 0")
	}
	produce, err := GetProduceStockById(id)
	if err != nil {
		return err
	}
	if produce.Amount+float64(total) < 0 {
		return errors.New("stock not enough")
	}

	produce.Amount += float64(total)

	return db.Updates(&produce).Error
}

// GetProduceStockFieldList 获取字段列表
func GetProduceStockFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.ProduceStock{})
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
