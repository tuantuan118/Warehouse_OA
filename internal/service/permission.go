package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetPermissionList(permission *models.Permission, pn, pSize int) ([]models.Permission, int64, error) {
	db := global.Db.Model(&models.Permission{})

	db = db.Where("enabled = ?", permission.Enabled)
	if permission.Name != "" {
		db = db.Where("name = ?", permission.Name)
	}
	if permission.Coding != "" {
		db = db.Where("coding = ?", permission.Coding)
	}
	if permission.Type != 0 {
		db = db.Where("type = ?", permission.Type)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db.Limit(pSize).Offset(offset)
	}

	var permissionList []models.Permission
	if err := db.Find(&permissionList).Error; err != nil {
		return nil, 0, err
	}
	return permissionList, total, nil
}

func SavePermission(permission *models.Permission) (*models.Permission, error) {
	var err error
	permission.Parent, err = getParent(permission.ParentID)
	if err != nil {
		return nil, err
	}

	err = global.Db.Model(&models.Permission{}).Create(permission).Error

	return permission, err
}

func UpdatePermission(permission *models.Permission) (*models.Permission, error) {
	if permission.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetPermissionById(permission.ID)
	if err != nil {
		return nil, err
	}

	permission.Parent, err = getParent(permission.ParentID)

	return permission, global.Db.Updates(&permission).Error
}

func DelPermission(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetPermissionById(id)
	if err != nil {
		return err
	}

	data.Operator = username
	data.IsDeleted = true
	err = global.Db.Updates(&data).Error
	if err != nil {
		return err
	}

	return global.Db.Delete(&data).Error
}

// GetPermissionFieldList 获取字段列表
func GetPermissionFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Permission{})
	switch field {
	case "name":
		db.Select("name")
	case "coding":
		db.Select("coding")
	case "type":
		db.Select("type")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}

func GetPermissionById(id int) (*models.Permission, error) {
	db := global.Db.Model(&models.Permission{})

	data := &models.Permission{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("permission does not exist")
	}

	return data, err
}

func GetPermissionByIdList(ids []int) ([]models.Permission, error) {
	db := global.Db.Model(&models.Permission{})

	var dataList []models.Permission
	err := db.Where("id IN ?", ids).Find(&dataList).Error

	return dataList, err
}

func getParent(id *int) (*models.Permission, error) {
	if id == nil {
		return nil, nil
	}

	if *id == 0 {
		return nil, nil
	}

	parent, err := GetPermissionById(*id)
	if err != nil {
		return nil, err
	}
	return parent, nil
}
