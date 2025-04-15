package controllers

import (
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
)

func OrgList(c *gin.Context) {
	orgListService, err := services.OrgListService(c)
	if err != nil {
		return
	}
	c.JSON(200, orgListService)
}

func OrgUserList(c *gin.Context) {
	orgUserListService, err := services.OrgUserListService()
	if err != nil {
		return
	}
	c.JSON(200, orgUserListService)
}

func OrgListOfChildren(c *gin.Context) {
	orgListOfChildrenService, err := services.GetChildren(c)
	if err != nil {
		return
	}
	c.JSON(200, orgListOfChildrenService)
}

func CreateOrg(c *gin.Context) {
	orgCreateService, err := services.CreateOrg(c)
	if err != nil {
		return
	}
	c.JSON(200, orgCreateService)
}

func UpdateOrg(c *gin.Context) {
	orgUpdateService, err := services.UpdateOrg(c)
	if err != nil {
		return
	}
	c.JSON(200, orgUpdateService)
}

func DeleteOrg(c *gin.Context) {
	orgDeleteService, err := services.DeleteOrg(c)
	if err != nil {
		return
	}
	c.JSON(200, orgDeleteService)
}
