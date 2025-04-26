package controllers

import (
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UserList(c *gin.Context) {
	userListService, err := services.UserListService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, userListService)
}

func CreateUser(c *gin.Context) {
	userCreateService, err := services.UserCreateService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, userCreateService)
}

func UserDetail(c *gin.Context) {
	userDetailService, err := services.UserDetailService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, userDetailService)
}

func UpdateUser(c *gin.Context) {
	userUpdateService, err := services.UserUpdateService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, userUpdateService)
}

func DeleteUser(c *gin.Context) {
	userDeleteService, err := services.UserDeleteService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, userDeleteService)
}

func UploadUser(c *gin.Context) {
	err := services.UserUploadService(c)
	if err != nil {
		return
	}
}
