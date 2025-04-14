package services

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/golang-jwt/jwt/v4"
	"strconv"
	"time"
)

type loginService struct {
}

func (*loginService) LoginService(l models.Login) (models.Result, error) {
	type UserWithRole struct {
		models.User
		RoleName   string `json:"role_name"`
		Permission string `json:"permission"`
	}
	var user UserWithRole

	config.DB.Table("users").
		Select("users.*, roles.role_name, roles.permission").
		Joins("left join roles on users.role_id = roles.id").
		Where("username = ? and password = ?", l.Username, l.Password).
		Scan(&user)

	if user.ID == 0 {
		return models.Fail(401, "用户名或密码错误"), nil
	}

	if user.Status == "0" {
		return models.Fail(401, "用户已被禁用"), nil
	}

	expireHours := config.AppGlobalConfig.Jwt.Expire
	expirationTime := time.Now().Add(time.Duration(expireHours) * time.Hour)

	claims := &models.CustomClaims{
		UserID:   user.ID,
		UserName: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "file_sys",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppGlobalConfig.Jwt.Secret))

	if err != nil {
		models.Error(500, "系统错误")
	}

	response := models.LoginResponse{
		Token:      tokenString,
		Age:        user.Age,
		Email:      user.Email,
		ID:         user.ID,
		RoleName:   user.RoleName,
		Username:   user.Username,
		Sex:        user.Sex,
		Avatar:     user.Avatar,
		Phone:      user.Phone,
		Status:     user.Status,
		Permission: user.Permission,
	}
	return models.Success(response), nil
}

func (*loginService) RegisterService(r models.Register) (models.Result, error) {
	var user models.User

	config.DB.Where("username = ?", r.Username).Find(&user)

	if user.ID != 0 {
		return models.Fail(401, "用户名已存在"), nil
	}
	user.Age, _ = strconv.Atoi(r.Age)
	user.Email = r.Email
	user.Password = r.Password
	user.Username = r.Username
	config.DB.Create(&user)
	if user.ID == 0 {
		return models.Fail(500, "注册失败"), nil
	}
	return models.Success(models.SuccessWithMsg("注册成功")), nil
}

var LoginService = new(loginService)
