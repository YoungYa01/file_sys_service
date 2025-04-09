package controllers

import (
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
)

func RoleList(c *gin.Context) {
	roleListService, err := services.RoleListService(c)
	if err != nil {
		return
	}
	c.JSON(200, roleListService)
}

func CreateRole(c *gin.Context) {
	roleCreateService, err := services.RoleCreateService(c)
	if err != nil {
		return
	}
	c.JSON(200, roleCreateService)
}

func UpdateRole(c *gin.Context) {
	roleUpdateService, err := services.RoleUpdateService(c)
	if err != nil {
		return
	}
	c.JSON(200, roleUpdateService)
}

func DeleteRole(c *gin.Context) {
	roleDeleteService, err := services.RoleDeleteService(c)
	if err != nil {
		return
	}
	c.JSON(200, roleDeleteService)
}
