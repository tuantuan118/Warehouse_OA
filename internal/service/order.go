package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"strings"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

func GetOrderList(order *models.Order, begTime, endTime string, pn, pSize int, userId int) (interface{}, error) {
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
	if order.CustomerId != 0 {
		db = db.Where("customer_id = ?", order.CustomerId)
	}
	if order.Status > 0 {
		db = db.Where("status = ?", order.Status)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}
	db = db.Preload("UserList")
	db = db.Preload("Customer")

	b, err := getAdmin(userId)
	if err != nil {
		return nil, err
	}
	if !b {
		db = db.Where(" id in (select order_id from tb_order_user where user_id = ?)", userId)
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

	finishedData, err := GetFinishedStockById(order.FinishedId)
	if err != nil {
		return nil, err
	}

	customer, err := GetCustomerById(order.CustomerId)
	if err != nil {
		return nil, err
	}
	order.Customer = customer

	today := time.Now().Format("20060102")
	total, err := getTodayOrderCount()
	if err != nil {
		return nil, err
	}

	order.OrderNumber = fmt.Sprintf("QY%s%d", today, total+10001)
	order.Name = finishedData.Name
	totalPrice := order.Price * float64(order.Amount)
	order.TotalPrice = totalPrice
	order.FinishPrice = 0
	order.UnFinishPrice = totalPrice
	order.Status = 1

	err = global.Db.Model(&models.Order{}).Create(order).Error

	return order, err
}

func UpdateOrder(order *models.Order) (*models.Order, error) {
	if order.ID == 0 {
		return nil, errors.New("id is 0")
	}
	oldData, err := GetOrderById(order.ID)
	if err != nil {
		return nil, err
	}

	if oldData.Status != 1 {
		return nil, errors.New("order has been finished, can not update")
	}

	if order.Price != oldData.Price || order.Amount != oldData.Amount {
		totalPrice := order.Price * float64(order.Amount)
		order.TotalPrice = totalPrice
		order.UnFinishPrice = totalPrice
	}

	customer, err := GetCustomerById(order.CustomerId)
	if err != nil {
		return nil, err
	}
	order.Customer = customer
	order.OrderNumber = ""
	order.Name = ""
	order.Status = 0

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

func FinishOrder(order *models.Order) (*models.Order, error) {
	if order.ID == 0 {
		return nil, errors.New("id is 0")
	}
	oldData, err := GetOrderById(order.ID)
	if err != nil {
		return nil, err
	}

	if oldData.Status != 2 {
		return nil, errors.New("order has been finished, can not update")
	}

	order.Amount = 0
	order.Price = 0
	order.OrderNumber = ""
	order.Name = ""
	order.UserList = nil

	order.UnFinishPrice = oldData.UnFinishPrice - order.FinishPrice
	order.FinishPrice += oldData.FinishPrice

	if oldData.TotalPrice <= order.FinishPrice {
		order.Status = 3
	}

	return order, global.Db.Select("UnFinishPrice",
		"FinishPrice",
		"Status").Updates(&order).Error
}
func VoidOrder(id int, username string) error {
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
	data.Status = 4

	return global.Db.Updates(&data).Error
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

// SaveOutBound 出库
func SaveOutBound(id int, username string) error {
	data, err := GetOrderById(id)

	data.Operator = username
	data.Status = 2

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = UpdateFinishedStockNum(tx, data.FinishedId, 0-data.Amount)
	if err != nil {
		return err
	}

	return tx.Updates(&data).Error
}

func ExportOrder(order *models.Order, begTime, endTime string) (*excelize.File, error) {
	db := global.Db.Model(&models.Order{})

	if order.ID != 0 {
		db = db.Where("id = ?", order.ID)
	}
	if order.OrderNumber != "" {
		db = db.Where("order_number = ?", order.OrderNumber)
	}
	if order.Name != "" {
		db = db.Where("name = ?", order.Name)
	}
	if order.Specification != "" {
		db = db.Where("specification = ?", order.Specification)
	}
	if order.CustomerId != 0 {
		db = db.Where("customer_name = ?", order.CustomerId)
	}
	if order.Status > 0 {
		db = db.Where("status = ?", order.Status)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("add_time BETWEEN ? AND ?", begTime, endTime)
	}

	data := make([]models.Order, 0)
	db = db.Preload("UserList")
	db = db.Preload("Customer")
	err := db.Find(&data).Error
	if err != nil {
		logrus.Infoln("导出订单错误: ", err.Error())
		return nil, err
	}

	keyList := []string{
		"订单编号",
		"产品名称",
		"单价（元）",
		"数量",
		"订单金额（元）",
		"已结金额（元）",
		"未结金额（元）",
		"订单状态",
		"客户名称",
		"订单分配",
		"销售人员",
		"创建时间",
		"备注",
	}

	valueList := make([]map[string]interface{}, 0)
	for _, v := range data {
		valueList = append(valueList, map[string]interface{}{
			"订单编号":    v.OrderNumber,
			"产品名称":    v.Name,
			"单价（元）":   v.Price,
			"数量":      v.Amount,
			"订单金额（元）": v.TotalPrice,
			"已结金额（元）": v.FinishPrice,
			"未结金额（元）": v.UnFinishPrice,
			"订单状态":    getOrderStatus(v.Status),
			"客户名称":    v.Customer.Name,
			"订单分配":    getOrderUser(v.UserList),
			"销售人员":    v.Salesman,
			"创建时间":    v.CreatedAt,
			"备注":      v.Remark,
		})
	}

	return utils.ExportExcel(keyList, valueList)
}

// GetOrderFieldList 获取字段列表
func GetOrderFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Order{})
	switch field {
	case "name":
		db.Distinct("name")
	case "orderNumber":
		db.Distinct("order_number")
	case "specification":
		db.Distinct("specification")
	case "salesman":
		db.Distinct("salesman")
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
	err := global.Db.Model(&models.Order{}).Where(
		"add_time >= ?", startOfDay).Count(&total).Error

	return total, err
}

func getOrderStatus(status int) string {
	switch status {
	case 1:
		return "待出库"
	case 2:
		return "未完成支付"
	case 3:
		return "已支付"
	case 4:
		return "作废"
	default:
		return "未知状态"
	}
}

func getOrderUser(userList []models.User) string {
	userStr := ""
	for _, user := range userList {
		userStr += user.Nickname + ", "
	}
	return strings.TrimRight(userStr, ", ")
}
