package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetRoleList(role *models.Role, pn, pSize int) ([]models.Role, int64, error) {
	db := global.Db.Model(&models.Role{})

	db = db.Where("enabled = ?", role.Enabled)
	if role.Name != "" {
		db = db.Where("name = ?", role.Name)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db.Limit(pSize).Offset(offset)
	}

	var roleList []models.Role
	if err := db.Find(&roleList).Error; err != nil {
		return nil, 0, err
	}
	return roleList, total, nil
}

func SaveRole(role *models.Role) (*models.Role, error) {
	_, err := GetRoleByNameEn(role.NameEn)
	if err != nil {
		return nil, err
	}

	err = global.Db.Model(&models.Role{}).Create(role).Error

	return role, err
}

func UpdateRole(role *models.Role) (*models.Role, error) {
	if role.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetRoleById(role.ID)
	if err != nil {
		return nil, err
	}

	role.Permissions = nil

	return role, global.Db.Updates(&role).Error
}

func DelRole(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetRoleById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("role does not exist")
	}

	data.Operator = username
	data.IsDeleted = true
	err = global.Db.Updates(&data).Error
	if err != nil {
		return err
	}

	return global.Db.Delete(&data).Error
}

// GetRoleFieldList 获取字段列表
func GetRoleFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Role{})
	switch field {
	case "name":
		db.Select("name")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}

// GetRoleByNameEn 判断角色英文名是否已存在
func GetRoleByNameEn(nameEn string) (*models.Role, error) {
	role := &models.Role{}

	result := global.Db.First(&role, "name_en = ?", nameEn)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("role does not exist")
		} else {
			return nil, errors.New(fmt.Sprintf("error occurred: %s", result.Error.Error()))
		}
	}

	return role, nil
}

// SetPermissions 分配角色
func SetPermissions(id int, roleIds []int, operator string) error {
	permissions, err := GetPermissionByIdList(roleIds)
	if err != nil {
		return err
	}

	role, err := GetRoleById(id)
	if err != nil {
		return err
	}

	role.Permissions = permissions
	role.Operator = operator
	return global.Db.Updates(&role).Error
}

func GetRoleById(id int) (*models.Role, error) {
	db := global.Db.Model(&models.Role{})

	data := &models.Role{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("role does not exist")
	}

	return data, err
}

func GetRoleByIdList(ids []int) ([]models.Role, error) {
	db := global.Db.Model(&models.Role{})

	var dataList []models.Role
	err := db.Where("id IN ?", ids).Find(&dataList).Error

	return dataList, err
}

func GetPermissions(ids []int) (interface{}, error) {
	db := global.Db.Model(&models.Role{})

	data := &models.Role{}
	err := db.Distinct().Where("id IN ?", ids).Preload("Permissions").First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("role does not exist")
	}
	if err != nil {
		return nil, err
	}
	permissions := data.Permissions
	logrus.Infoln(data)

	type request struct {
		*models.Permission
		Children []*request `json:"children"`
	}

	itemMap := make(map[int]*request)
	var roots []*request

	// Populate the map
	for i := range permissions {
		itemMap[permissions[i].ID] = &request{
			Permission: &permissions[i],
			Children:   make([]*request, 0),
		}
	}

	// Build the tree structure
	for i := range permissions {
		item := itemMap[permissions[i].ID]
		if item.ParentID == nil {
			// If ParentID is nil, this is a root item
			roots = append(roots, item)
		} else {
			// Otherwise, add it to its parent's Children slice
			parent := itemMap[*item.ParentID]
			parent.Children = append(parent.Children, item)
		}
	}

	return roots, nil
}
