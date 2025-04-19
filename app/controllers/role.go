package controllers

import (
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RoleList(c *gin.Context) {
	roleListService, err := services.RoleListService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, roleListService)
}

func CreateRole(c *gin.Context) {
	roleCreateService, err := services.RoleCreateService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, roleCreateService)
}

func UpdateRole(c *gin.Context) {
	roleUpdateService, err := services.RoleUpdateService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, roleUpdateService)
}

func DeleteRole(c *gin.Context) {
	roleDeleteService, err := services.RoleDeleteService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, roleDeleteService)
}
