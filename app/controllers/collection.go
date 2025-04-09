package controllers

import (
	"gin_back/app/models"
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func CollectionList(c *gin.Context) {
	collectionListService, err := services.CollectionListService(c)
	if err != nil {
		return
	}
	c.JSON(200, collectionListService)
}

func CollectionCreate(c *gin.Context) {
	// 1. 接收并验证请求参数
	var creator models.Collection
	if err := c.ShouldBindJSON(&creator); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "参数错误"+err.Error()))
		return
	}
	// 2. 调用服务层创建方法
	if err := services.CreateCollectionService(creator); err != nil {
		// 根据错误类型返回不同状态码
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "参数校验失败") {
			c.JSON(200, models.Error(400, "参数校验失败"+err.Error()))
			return
		}

		c.JSON(200, models.Error(statusCode, "创建失败"+err.Error()))
		return
	}

	c.JSON(200, models.Success(creator))
}

func CollectionUpdate(c *gin.Context) {
	collectionUpdateService, err := services.UpdateCollectionService(c)
	if err != nil {
		return
	}
	c.JSON(200, collectionUpdateService)
}

func CollectionDelete(c *gin.Context) {
	collectionDeleteService, err := services.DeleteCollectionService(c)
	if err != nil {
		return
	}
	c.JSON(200, collectionDeleteService)
}
