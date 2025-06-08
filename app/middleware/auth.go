package middleware

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		if token == "" {
			c.JSON(200, models.Error(401, "token不能为空"))
			c.Abort()
			return
		}
		// 修改Claims类型为CustomClaims
		claims, err := jwt.ParseWithClaims(
			token,
			&models.CustomClaims{}, // 使用自定义Claims
			func(token *jwt.Token) (interface{}, error) {
				return []byte(config.AppGlobalConfig.Jwt.Secret), nil
			})
		if err != nil {
			handleJWTError(c, err)
			return
		}
		// 修改类型断言
		if customClaims, ok := claims.Claims.(*models.CustomClaims); ok && claims.Valid {
			c.Set("claims", customClaims) // 存储完整的自定义Claims
			c.Set("userId", customClaims.UserID)
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
			c.JSON(401, models.Error(401, "登录已过期，请重新登录"))
		case ve.Errors&jwt.ValidationErrorSignatureInvalid != 0:
			c.JSON(401, models.Error(401, "系统签名错误"))
		default:
			c.JSON(401, models.Error(401, "登录无效，请重新登录"))
		}
	} else {
		c.JSON(401, models.Error(401, "token解析失败"))
	}
	c.Abort()
}
