package services

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type jwtService struct {
}

var JwtService = new(jwtService)

func (s jwtService) CreateToken(u models.Login) (string, error) {
	expirationTime := time.Now().Add(time.Duration(config.AppGlobalConfig.Jwt.Expire) * time.Hour)

	claims := &models.CustomClaims{
		UserID:   u.ID, // 确保这里注入用户ID
		UserName: u.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "file_sys",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppGlobalConfig.Jwt.Secret))
}
