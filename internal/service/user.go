package service

import (
	"errors"
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

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db.Limit(pSize).Offset(offset)
	}

	var userList []models.User
	if err := db.Find(&userList).Error; err != nil {
		return nil, 0, err
	}
	return userList, total, nil
}

func GetUserById(id int) (*models.User, error) {
	db := global.Db.Model(&models.User{})

	data := &models.User{}
	err := db.Where("id = ?", id).First(&models.User{}).Error

	return data, err
}

func SaveUser(user *models.User) (*models.User, error) {
	_, total, err := GetUserList(user, 0, 0)
	if err != nil {
		return nil, err
	}
	if total > 0 {
		return nil, errors.New("user is exist")
	}
	user.Password = utils.GenMd5(user.Password)

	err = global.Db.Model(&models.User{}).Create(user).Error

	return user, err
}

func UpdateUser(user *models.User) (*models.User, error) {
	old, err := GetUserById(user.ID)
	if err != nil {
		return nil, err
	}
	if old == nil {
		return nil, errors.New("user does not exist")
	}
	if user.Password != old.Password {
		user.Password = utils.GenMd5(user.Password)
	}
	user.Roles = nil

	return user, global.Db.Updates(&user).Error
}

func DelUser(id int) error {
	data, err := GetUserById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}
	err = global.Db.Model(&data).Update("IsDeleted", true).Error
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
