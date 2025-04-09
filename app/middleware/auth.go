package middleware

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取token
		token := c.Request.Header.Get("token")
		if token == "" {
			c.JSON(200, models.Error(401, "token不能为空"))
			c.Abort()
			return
		}
		claims, err := jwt.ParseWithClaims(
			token,
			&jwt.StandardClaims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(config.AppGlobalConfig.Jwt.Secret), nil
			})
		if err != nil {
			handleJWTError(c, err)
			return
		}
		if standardClaims, ok := claims.Claims.(*jwt.StandardClaims); ok && claims.Valid {
			c.Set("claims", standardClaims)
			c.Next()
		} else {
			c.JSON(401, models.Error(401, "无效token"))
			c.Abort()
		}
	}
}

// 统一处理 JWT 错误
func handleJWTError(c *gin.Context, err error) {
	if ve, ok := err.(*jwt.ValidationError); ok {
		switch {
		case ve.Errors&jwt.ValidationErrorExpired != 0:
			c.JSON(401, models.Error(401, "token已过期"))
		case ve.Errors&jwt.ValidationErrorSignatureInvalid != 0:
			c.JSON(401, models.Error(401, "token签名错误"))
		default:
			c.JSON(401, models.Error(401, "无效token"))
		}
	} else {
		c.JSON(401, models.Error(401, "token解析失败"))
	}
	c.Abort()
}
