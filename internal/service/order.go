package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetOrderList(order *models.Order, begTime, endTime string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.Order{})

	if order.OrderNumber != "" {
		db = db.Where("order_number = ?", order.OrderNumber)
	}
	if order.Name != "" {
		db = db.Where("name = ?", order.Name)
	}
	if order.Specification != "" {
		db = db.Where("specification = ?", order.Specification)
	}
	if order.CustomerName != "" {
		db = db.Where("customer_name = ?", order.CustomerName)
	}
	if order.Status > 0 {
		db = db.Where("status = ?", order.Status)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("add_time BETWEEN ? AND ?", begTime, endTime)
	}

	return Pagination(db, []models.Order{}, pn, pSize)
}

func GetOrderById(id int) (*models.Order, error) {
	db := global.Db.Model(&models.Order{})

	data := &models.Order{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveOrder(order *models.Order) (*models.Order, error) {
	userList := make([]models.User, 0)
	if order.UserList != nil || len(order.UserList) > 0 {
		for _, v := range order.UserList {
			user, err := GetUserById(v.ID)
			if err != nil {
				return nil, err
			}
			userList = append(userList, *user)
		}
	}
	err := global.Db.Model(&models.Order{}).Create(order).Error

	return order, err
}

func UpdateOrder(order *models.Order) (*models.Order, error) {
	if order.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetOrderById(order.ID)
	if err != nil {
		return nil, err
	}

	userList := make([]models.User, 0)
	if order.UserList != nil || len(order.UserList) > 0 {
		for _, v := range order.UserList {
			user, err := GetUserById(v.ID)
			if err != nil {
				return nil, err
			}
			userList = append(userList, *user)
		}
	}

	return order, global.Db.Updates(&order).Error
}

func DelOrder(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetOrderById(id)
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

// GetOrderFieldList 获取字段列表
func GetOrderFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Order{})
	switch field {
	case "name":
		db.Select("name")
	case "orderNumber":
		db.Select("order_number")
	case "specification":
		db.Select("specification")
	case "customerName":
		db.Select("customer_name")
	case "salesman":
		db.Select("salesman")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}
