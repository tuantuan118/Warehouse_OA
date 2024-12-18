package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"math/big"
	"strings"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

func GetInBoundList(name, supplier, stockUser, stockUnit, begTime, endTime string,
	pn, pSize int, b bool) (interface{}, error) {
	db := global.Db.Model(&models.IngredientInBound{})
	totalDb := global.Db.Model(&models.IngredientInBound{})

	if name != "" {
		idList, err := GetIngredientsByName(name)
		if err != nil {
			return nil, err
		}
		db = db.Where("ingredient_id in ?", idList)
		totalDb = totalDb.Where("ingredient_id in ?", idList)
	}
	if supplier != "" {
		idList, err := GetIngredientsBySupplier(supplier)
		if err != nil {
			return nil, err
		}
		db = db.Where("ingredient_id in ?", idList)
		totalDb = totalDb.Where("ingredient_id in ?", idList)
	}
	if stockUser != "" {
		db = db.Where("stock_user = ?", stockUser)
		totalDb = totalDb.Where("stock_user = ?", stockUser)
	}
	if stockUnit != "" {
		db = db.Where("stock_unit = ?", stockUnit)
		totalDb = totalDb.Where("stock_unit = ?", stockUnit)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	var totalPrice float64
	err := totalDb.Select("COALESCE(SUM(total_price), 0)").Scan(&totalPrice).Error
	if err != nil {
		return nil, err
	}

	if b {
		db = db.Where("in_and_out = ?", b)
	}
	db = db.Preload("Ingredient")

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("id desc").Limit(pSize).Offset(offset)
	}

	data := make([]models.IngredientInBound, 0)
	err = db.Find(&data).Error
	if err != nil {
		return nil, err
	}

	for n := range data {
		if data[n].FinishPriceStr == "" {
			continue
		}
		data[n].FinishPriceList = make([]map[string]string, 0)
		fpl := strings.Split(data[n].FinishPriceStr, ";")
		for _, f := range fpl {
			fp := strings.Split(f, "&")
			if len(fp) != 2 {
				continue
			}
			data[n].FinishPriceList = append(data[n].FinishPriceList, map[string]string{
				"time":  fp[0],
				"price": fp[1],
			})
		}
	}

	return map[string]interface{}{
		"data":            data,
		"pageNo":          pn,
		"pageSize":        pSize,
		"totalCount":      total,
		"sum_total_price": totalPrice,
	}, err

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

	price := big.NewFloat(inBound.Price)
	stockNum := big.NewFloat(inBound.StockNum)
	floatResult := new(big.Float).Mul(price, stockNum)
	inBound.TotalPrice, _ = floatResult.Float64()
	inBound.Ingredient = ingredients
	inBound.InAndOut = true
	inBound.OperationType = "入库"
	inBound.OperationDetails = fmt.Sprintf("配料入库")
	inBound.UnFinishPrice, _ = floatResult.Float64()

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = SaveInventoryByInBound(tx, inBound)
	if err != nil {
		return nil, err
	}
	err = tx.Model(&models.IngredientInBound{}).Create(inBound).Error

	return inBound, err
}

func UpdateInBound(inBound *models.IngredientInBound) (*models.IngredientInBound, error) {
	if inBound.ID == 0 {
		return nil, errors.New("id is 0")
	}
	var err error
	oldData := new(models.IngredientInBound)
	oldData, err = GetInBoundById(inBound.ID)
	if err != nil {
		return nil, err
	}

	if oldData.IngredientID != inBound.IngredientID {
		ingredients := new(models.Ingredients)
		ingredients, err = GetIngredientsById(*inBound.IngredientID)
		if err != nil {
			return nil, err
		}

		inBound.Ingredient = ingredients
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

	inBound.TotalPrice = inBound.Price * inBound.StockNum
	inBound.UnFinishPrice = inBound.TotalPrice - oldData.FinishPrice
	err = tx.Updates(&inBound).Error
	if err != nil {
		return nil, err
	}

	inBound.StockNum = inBound.StockNum - oldData.StockNum

	return inBound, SaveInventoryByInBound(tx, inBound)
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

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = tx.Updates(&data).Error
	if err != nil {
		return err
	}
	err = tx.Delete(&data).Error
	if err != nil {
		return err
	}

	data.StockNum = 0 - data.StockNum
	return UpdateInventoryByInBound(tx, data)
}

func ExportIngredients(name, supplier, stockUser, begTime, endTime string) (*excelize.File, error) {
	db := global.Db.Model(&models.IngredientInBound{})
	totalDb := global.Db.Model(&models.IngredientInBound{})

	if name != "" {
		idList, err := GetIngredientsByName(name)
		if err != nil {
			return nil, err
		}
		db = db.Where("ingredient_id in ?", idList)
		totalDb = totalDb.Where("ingredient_id in ?", idList)
	}
	if supplier != "" {
		idList, err := GetIngredientsBySupplier(supplier)
		if err != nil {
			return nil, err
		}
		db = db.Where("ingredient_id in ?", idList)
		totalDb = totalDb.Where("ingredient_id in ?", idList)
	}
	if stockUser != "" {
		db = db.Where("stock_user = ?", stockUser)
		totalDb = totalDb.Where("stock_user = ?", stockUser)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	var totalPrice float64
	err := totalDb.Select("COALESCE(SUM(total_price), 0)").Scan(&totalPrice).Error
	if err != nil {
		return nil, err
	}

	data := make([]models.IngredientInBound, 0)
	err = db.Preload("Ingredient").Find(&data).Error
	if err != nil {
		logrus.Infoln("导出订单错误: ", err.Error())
	}

	keyList := []string{
		"配料名称",
		"配料供应商",
		"配料规格",
		"单价（元）",
		"金额（元）",
		"采购金额",
		"入库数量",
		"入库人员",
		"入库时间",
		"备注",
	}

	valueList := make([]map[string]interface{}, 0)
	for _, v := range data {
		valueList = append(valueList, map[string]interface{}{
			"配料名称":  v.Ingredient.Name,
			"配料供应商": v.Ingredient.Supplier,
			"配料规格":  v.Specification,
			"单价（元）": v.Price,
			"金额（元）": v.TotalPrice,
			"采购金额":  v.Price * v.StockNum,
			"入库数量":  fmt.Sprintf("%f%s", v.StockNum, returnUnit(v.StockUnit)),
			"入库人员":  v.StockUser,
			"入库时间":  v.StockTime,
			"备注":    v.Remark,
		})
	}
	valueList = append(valueList, map[string]interface{}{
		"金额（元）": totalPrice,
	})

	return utils.ExportExcel(keyList, valueList)
}

func returnUnit(i int) string {
	switch i {
	case 1:
		return "斤"
	case 2:
		return "克"
	case 3:
		return "件"
	case 4:
		return "个"
	case 5:
		return "张"
	case 6:
		return "盆"
	case 7:
		return "桶"
	case 8:
		return "包"
	case 9:
		return "箱"
	}
	return ""
}

func FinishedSaveInBound(tx *gorm.DB, inBound *models.IngredientInBound) error {
	ingredients, err := GetIngredientsById(*inBound.IngredientID)
	if err != nil {
		return err
	}

	inBound.Ingredient = ingredients

	err = SaveInventoryByInBound(tx, inBound)
	if err != nil {
		return err
	}
	err = tx.Model(&models.IngredientInBound{}).Create(inBound).Error

	return nil
}

func GetOutInBoundList(id int, supplier, stockUser, begTime, endTime string,
	pn, pSize int) (interface{}, error) {

	var name string
	var stockUnit string
	if id != 0 {
		inventory, err := GetInventoryById(id)
		if err != nil {
			return nil, err
		}
		name = inventory.Ingredient.Name
		stockUnit = fmt.Sprintf("%d", inventory.StockUnit)
	}

	return GetInBoundList(
		name,
		supplier,
		stockUser,
		stockUnit,
		begTime,
		endTime,
		pn,
		pSize,
		false,
	)
}

func FinishInBound(id int, totalPrice float64) (*models.IngredientInBound, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}
	data, err := GetInBoundById(id)
	if err != nil {
		return nil, err
	}

	if data.Status == 1 {
		return nil, errors.New("ingredient has been finished, can not update")
	}

	data.UnFinishPrice = data.UnFinishPrice - totalPrice
	data.FinishPrice += totalPrice

	str := fmt.Sprintf("%s&%f;", time.Now().Format("2006-01-02 15:04:05"), totalPrice)
	data.FinishPriceStr += str

	if data.UnFinishPrice > 0 {
		data.Status = 0
	} else {
		data.Status = 1
	}

	return data, global.Db.Select("UnFinishPrice",
		"FinishPrice", "FinishPriceStr",
		"Status").Updates(&data).Error
}
