package service

import (
	"errors"
	"gorm.io/gorm"
	"math/big"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetInBoundList(name, supplier, stockUser, begTime, endTime string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.IngredientInBound{})
	totalDb := global.Db.Model(&models.IngredientInBound{})

	if name != "" {
		db = db.Where("name = ?", name)
		totalDb = totalDb.Where("name = ?", name)
	}
	if supplier != "" {
		db = db.Where("supplier = ?", supplier)
		totalDb = totalDb.Where("supplier = ?", supplier)
	}
	if stockUser != "" {
		db = db.Where("stock_user = ?", stockUser)
		totalDb = totalDb.Where("stock_user = ?", stockUser)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("created_at BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("created_at BETWEEN ? AND ?", begTime, endTime)
	}

	var totalPrice float64
	err := totalDb.Select("SUM(total_price)").Scan(&totalPrice).Error
	if err != nil {
		return nil, err
	}

	db = db.Preload("Ingredient")
	m, err := Pagination(db, []models.IngredientInBound{}, pn, pSize)
	if err != nil {
		return nil, err
	}

	m["sum_total_price"] = totalPrice

	return m, nil
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
	stockNum := big.NewFloat(float64(inBound.StockNum))
	floatResult := new(big.Float).Mul(price, stockNum)
	inBound.TotalPrice, _ = floatResult.Float64()
	inBound.Ingredient = ingredients

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

	inBound.TotalPrice = inBound.Price * float64(inBound.StockNum)
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
