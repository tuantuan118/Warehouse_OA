package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

func GetUserList(user *models.User, pn, pSize int) ([]models.User, int64, error) {
	db := global.Db.Model(&models.User{})

	if user.Name != "" {
		db = db.Where("name = ?", user.Name)
	}
	if user.Email != "" {
		db = db.Where("email = ?", user.Email)
	}
	if user.Type != 0 {
		db = db.Where("type = ?", user.Type)
	}
	if user.Organize != "" {
		db = db.Where("organize = ?", user.Organize)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db.Limit(pSize).Offset(offset)
	}

	var userList []models.User
	if err := db.Preload("Roles").Find(&userList).Error; err != nil {
		return nil, 0, err
	}

	return userList, total, nil
}

func GetUserById(id int) (*models.User, error) {
	db := global.Db.Model(&models.User{})

	data := &models.User{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("role does not exist")
	}

	return data, err
}

func SaveUser(user *models.User) (*models.User, error) {
	_, err := GetUserByNameAndEmail(user.Name, user.Email)
	if err != nil {
		return nil, err
	}

	user.Password = utils.GenMd5(user.Password)
	err = global.Db.Model(&models.User{}).Create(user).Error

	return user, err
}

func UpdateUser(user *models.User) (*models.User, error) {
	if user.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetUserById(user.ID)
	if err != nil {
		return nil, err
	}

	user.Email = ""
	user.Name = ""
	user.Password = ""
	user.Roles = nil

	return user, global.Db.Updates(&user).Error
}

func DelUser(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetUserById(id)
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

func CheckPassword(email, password string) (*models.User, error) {
	user := &models.User{}

	db := global.Db.Model(&models.User{})
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}

	if user.Password == utils.GenMd5(password) {
		return user, nil
	}
	return nil, errors.New("wrong password")
}

// ChangePassword 修改密码
func ChangePassword(id int, oldPw, newPw, username string) error {
	user, err := GetUserById(id)
	if err != nil {
		return err
	}

	if user.Password != utils.GenMd5(oldPw) {
		return errors.New("wrong password")
	}
	user.Operator = username
	user.Password = utils.GenMd5(newPw)

	return global.Db.Updates(&user).Error
}

// GetUserFieldList 获取字段列表
func GetUserFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.User{})
	switch field {
	case "name":
		db.Select("name")
	case "email":
		db.Select("email")
	case "type":
		db.Select("type")
	case "organize":
		db.Select("organize")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}

// GetUserByNameAndEmail 判断邮件是否已存在
func GetUserByNameAndEmail(name, email string) (*models.User, error) {
	user := &models.User{}

	result := global.Db.First(&user, "name = ? and email = ?", name, email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user does not exist")
		} else {
			return nil, errors.New(fmt.Sprintf("error occurred: %s", result.Error.Error()))
		}
	}

	return user, nil
}

// SetRoles 分配角色
func SetRoles(id int, roleIds []int, operator string) error {
	roles, err := GetRoleByIdList(roleIds)
	if err != nil {
		return err
	}

	user, err := GetUserById(id)
	if err != nil {
		return err
	}

	err = global.Db.Model(&user).Association("Roles").Clear()
	if err != nil {
		return err
	}

	user.Roles = roles
	user.Operator = operator
	return global.Db.Save(&user).Error
}

func GetRolePermissions(id int) (interface{}, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}

	db := global.Db.Model(&models.User{})

	data := &models.User{}
	err := db.Where("id = ?", id).Preload("Roles").First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("role does not exist")
	}

	ids := make([]int, 0)
	for _, role := range data.Roles {
		ids = append(ids, role.ID)
	}
	permissions, err := GetPermissions(ids)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}
