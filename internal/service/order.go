package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"os"
	"os/exec"
	"strings"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

func GetOrderList(order *models.Order, customerStr, begTime, endTime string, pn, pSize int, userId int) (interface{}, error) {
	db := global.Db.Model(&models.Order{})

	if order.OrderNumber != "" {
		slice := strings.Split(order.OrderNumber, ";")
		db = db.Where("order_number in ?", slice)
	}
	if order.Name != "" {
		slice := strings.Split(order.Name, ";")
		db = db.Where("name in ?", slice)
	}
	if order.Specification != "" {
		db = db.Where("specification = ?", order.Specification)
	}
	if order.Salesman != "" {
		slice := strings.Split(order.Salesman, ";")
		db = db.Where("salesman in ?", slice)
	}
	if customerStr != "" {
		slice := strings.Split(customerStr, ";")
		db = db.Where("customer_id in ?", slice)
	}
	if order.Status != 0 {
		db = db.Where("status = ?", order.Status)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}
	db = db.Preload("UserList")
	db = db.Preload("Customer")
	db = db.Preload("Ingredient")

	b, err := getAdmin(userId)
	if err != nil {
		return nil, err
	}
	if !b {
		db = db.Where(" id in (select order_id from tb_order_user where user_id = ?)", userId)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("id desc").Limit(pSize).Offset(offset)
	}

	data := make([]models.Order, 0)
	err = db.Find(&data).Error

	for n := range data {
		data[n].ImageList = make([]string, 0)
		if data[n].Images != "" {
			data[n].ImageList = strings.Split(data[n].Images, ";")
		}

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
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
	}, err
}

func GetOrderById(id int) (*models.Order, error) {
	db := global.Db.Model(&models.Order{})

	data := &models.Order{}
	db = db.Preload("UserList")
	db = db.Preload("Customer")
	db = db.Preload("Ingredient.IngredientInventory")
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveOrder(order *models.Order) (*models.Order, error) {
	var err error

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

	order.Images = strings.Join(order.ImageList, ";")

	for _, ingredient := range order.Ingredient {
		inventory := new(models.IngredientInventory)
		inventory, err = GetInventoryById(ingredient.IngredientID)
		if err != nil {
			return nil, err
		}
		ingredient.IngredientInventory = inventory
	}

	productData, err := GetProductById(order.ProductId)
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
	order.Name = productData.Name
	order.Specification = productData.Specification
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

	order.Images = strings.Join(order.ImageList, ";")

	for _, ingredient := range order.Ingredient {
		inventory := new(models.IngredientInventory)
		inventory, err = GetInventoryById(ingredient.IngredientID)
		if err != nil {
			return nil, err
		}
		ingredient.OrderID = order.ID
		ingredient.IngredientInventory = inventory
	}

	customer, err := GetCustomerById(order.CustomerId)
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

	// 清除 UserList 关联
	if err := global.Db.Model(&oldData).Association("UserList").Clear(); err != nil {
		return nil, err
	}

	order.UserList = userList
	order.Customer = customer
	order.OrderNumber = ""
	order.Name = ""
	order.Status = 0

	return order, global.Db.Updates(&order).Error
}

func FinishOrder(id int, totalPrice float64) (*models.Order, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}
	data, err := GetOrderById(id)
	if err != nil {
		return nil, err
	}

	if data.Status != 2 {
		return nil, errors.New("order has been finished, can not update")
	}

	data.UnFinishPrice = data.UnFinishPrice - totalPrice
	data.FinishPrice += totalPrice

	str := fmt.Sprintf("%s&%f;", time.Now().Format("2006-01-02 15:04:05"), totalPrice)
	data.FinishPriceStr += str

	if data.UnFinishPrice > 0 {
		data.Status = 2
	} else {
		data.Status = 3
	}

	return data, global.Db.Select("UnFinishPrice",
		"FinishPrice", "FinishPriceStr",
		"Status").Updates(&data).Error
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
	if err != nil {
		return err
	}

	if data.Status != 1 {
		return errors.New("order has been finished, can not out")
	}

	product, err := GetProductById(data.ProductId)
	if err != nil {
		return err
	}

	manageId, err := GetFinishedManageById(product.FinishedManageId)
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
	ft := time.Now()
	err = ProductSaveFinished(tx, &models.Finished{
		BaseModel: models.BaseModel{
			Operator: username,
		},
		Name:             product.Name,
		ActualAmount:     0 - product.Amount*data.Amount,
		Status:           2,
		FinishTime:       &ft,
		FinishedManageId: product.FinishedManageId,
		FinishedManage:   manageId,
		OperationType:    "出库",
		OperationDetails: fmt.Sprintf("【%s】销售出库", product.Name),
	})
	if err != nil {
		return err
	}

	for _, i := range data.Ingredient {
		logrus.Infoln("1111111111111111")
		logrus.Infoln(i)
		err = FinishedSaveInBound(tx, &models.IngredientInBound{
			BaseModel: models.BaseModel{
				Operator: username,
			},
			IngredientID:     i.IngredientInventory.IngredientID,
			StockNum:         float64(0 - i.Quantity),
			StockUnit:        i.IngredientInventory.StockUnit,
			StockUser:        username,
			StockTime:        time.Now(),
			OperationType:    "出库",
			OperationDetails: fmt.Sprintf("订单编号【%s】附加材料", data.OrderNumber),
		})
		if err != nil {
			return err
		}
	}

	data.Operator = username
	data.Status = 2

	return tx.Updates(&data).Error
}

func ExportOrder(order *models.Order) ([]byte, error) {
	db := global.Db.Model(&models.Order{})

	if order.ID != 0 {
		db = db.Where("id = ?", order.ID)
	}

	data := &models.Order{}
	db = db.Preload("UserList")
	db = db.Preload("Customer")
	db = db.Preload("Ingredient")
	err := db.First(&data).Error
	if err != nil {
		logrus.Infoln("导出订单错误: ", err.Error())
		return nil, err
	}

	filePath := "./stencil.xlsx"
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer func(f *excelize.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	F6 := fmt.Sprintf("开单时间：%s", data.CreatedAt.Format("2006/01/02"))
	if err := f.SetCellValue("Sheet1", "F6", F6); err != nil {
		return nil, err
	}
	B7 := fmt.Sprintf("客户编号：%d", data.Customer.ID)
	if err := f.SetCellValue("Sheet1", "B7", B7); err != nil {
		return nil, err
	}
	D7 := fmt.Sprintf("客户名称：%s", data.Customer.Name)
	if err := f.SetCellValue("Sheet1", "D7", D7); err != nil {
		return nil, err
	}
	F7 := fmt.Sprintf("客户联系方式：%s", data.Customer.Phone)
	if err := f.SetCellValue("Sheet1", "F7", F7); err != nil {
		return nil, err
	}
	B8 := fmt.Sprintf("收货地址：%s", data.Customer.Address)
	if err := f.SetCellValue("Sheet1", "B8", B8); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", "B11", data.ProductId); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", "C11", data.Name); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", "D11", data.Specification); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", "E11", data.Amount); err != nil {
		return nil, err
	}
	F11 := fmt.Sprintf("¥%0.2f", data.Price)
	if err := f.SetCellValue("Sheet1", "F11", F11); err != nil {
		return nil, err
	}
	G11 := fmt.Sprintf("¥%0.2f", data.Price)
	if err := f.SetCellValue("Sheet1", "G11", G11); err != nil {
		return nil, err
	}
	totalPrice := utils.AmountConvert(data.TotalPrice, true)
	B13 := fmt.Sprintf("合计(大写): %s", totalPrice)
	if err := f.SetCellValue("Sheet1", "B13", B13); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", "E13", data.Amount); err != nil {
		return nil, err
	}
	G13 := fmt.Sprintf("¥%0.2f", data.TotalPrice)
	if err := f.SetCellValue("Sheet1", "G13", G13); err != nil {
		return nil, err
	}
	F15 := fmt.Sprintf("制单人：%s", data.Salesman)
	if err := f.SetCellValue("Sheet1", "F15", F15); err != nil {
		return nil, err
	}

	newName := fmt.Sprintf("./cos/execl/%d.xlsx", data.ID)
	if err := f.SaveAs(newName); err != nil {
		return nil, err
	} else {
		logrus.Infoln("文件已成功另存为", newName)
	}

	cmd := exec.Command("libreoffice",
		"--invisible",
		"--convert-to",
		"pdf",
		"--outdir",
		"./cos/pdf/",
		newName,
	)
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	pdfName := fmt.Sprintf("./cos/pdf/%d.pdf", data.ID)
	pdfData, err := os.ReadFile(pdfName)
	if err != nil {
		return nil, err
	}

	return pdfData, nil
}

// GetOrderFieldList 获取字段列表
func GetOrderFieldList(field string, userId int) ([]string, error) {
	db := global.Db.Model(&models.Order{})
	switch field {
	case "name":
		db = db.Distinct("name")
	case "orderNumber":
		db = db.Distinct("order_number")
	case "specification":
		db = db.Distinct("specification")
	case "salesman":
		db = db.Distinct("salesman")
	default:
		return nil, errors.New("field not exist")
	}
	b, err := getAdmin(userId)
	if err != nil {
		return nil, err
	}
	if !b {
		db = db.Where(" id in (select order_id from tb_order_user where user_id = ?)", userId)
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

func GetOrderByCustomer(customerId int) error {
	db := global.Db.Model(&models.Order{})

	data := &models.Order{}
	err := db.Where("customer_id = ?", customerId).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user does not exist")
	}

	return err
}
