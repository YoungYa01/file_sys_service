package controllers

import (
	"gin_back/app/models"
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Login(c *gin.Context) {

	var loginInfo models.Login

	if err := c.ShouldBindJSON(&loginInfo); err != nil {
		// 返回错误信息（自动处理 400 状态码）
		c.JSON(http.StatusBadRequest, models.Error(400, "请求参数错误"))
		return
	}

	loginService, err := services.LoginService.LoginService(loginInfo)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.Error(401, "登录失败"))
		return
	}
	c.JSON(200, loginService)
}

func Register(r *gin.Context) {
	register := models.Register{
		Username: r.PostForm("username"),
		Password: r.PostForm("password"),
		Email:    r.PostForm("email"),
		Age:      r.PostForm("age"),
	}
	registerService, err := services.LoginService.RegisterService(register)
	if err != nil {
		return
	}
	r.JSON(200, registerService)
}

func LoginTest(r *gin.Context) {
	r.JSON(200, models.Success("登录成功"))
}
