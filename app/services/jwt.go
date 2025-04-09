package services

import (
	"fmt"
	"gin_back/app/models"
	"gin_back/config"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type jwtService struct {
}

var JwtService = new(jwtService)

func (s jwtService) CreateToken(u models.Login) (string, error) {
	expireSeconds := config.AppGlobalConfig.Jwt.Expire
	if expireSeconds <= 0 {
		return "Token过期时间配置错误", fmt.Errorf("invalid expire time")
	}

	now := time.Now()

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireSeconds) * time.Second)),
		ID:        u.Username,
		Issuer:    config.AppGlobalConfig.Jwt.Issuer,
		NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)), // 调整为 30 秒前生效
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(config.AppGlobalConfig.Jwt.Secret))

	if err != nil {
		return "Token生成失败", err
	}

	return token, nil
}
