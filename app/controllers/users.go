package controllers

import (
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
)

func UserList(c *gin.Context) {
	userListService, err := services.UserListService(c)
	if err != nil {
		return
	}
	c.JSON(200, userListService)
}

func CreateUser(c *gin.Context) {
	userCreateService, err := services.UserCreateService(c)
	if err != nil {
		return
	}
	c.JSON(200, userCreateService)
}

func UpdateUser(c *gin.Context) {
	userUpdateService, err := services.UserUpdateService(c)
	if err != nil {
		return
	}
	c.JSON(200, userUpdateService)
}

func DeleteUser(c *gin.Context) {
	userDeleteService, err := services.UserDeleteService(c)
	if err != nil {
		return
	}
	c.JSON(200, userDeleteService)
}
