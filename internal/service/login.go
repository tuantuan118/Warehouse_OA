package service

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

func Login(email, password string) (map[string]interface{}, error) {
	if email == "" || password == "" {
		return nil, errors.New("email or password is empty")
	}

	user, err := CheckPassword(email, password)
	if err != nil {
		return nil, err
	}

	jwtUser := utils.NewJWT()
	token, err := jwtUser.CreateToken(utils.CustomClaims{
		Id:   user.ID,
		Name: user.Name,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    "jia_hua",
		},
	})

	return map[string]interface{}{
		"token": token,
	}, nil
}

func Register(user *models.User) (map[string]interface{}, error) {
	if user.Name == "" || user.Email == "" || user.Password == "" {
		return nil, errors.New("user data is empty")
	}

	user, err := SaveUser(user)
	if err != nil {
		return nil, err
	}

	jwtUser := utils.NewJWT()
	token, err := jwtUser.CreateToken(utils.CustomClaims{
		Id:   user.ID,
		Name: user.Name,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    "jia_hua",
		},
	})

	return map[string]interface{}{
		"token": token,
	}, nil
}
