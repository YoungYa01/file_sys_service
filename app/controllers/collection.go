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
	c.JSON(http.StatusOK, collectionListService)
}

func TCCollectionList(c *gin.Context) {
	collectionListService, err := services.TCCollectionListService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, collectionListService)
}

func TCSubmit(c *gin.Context) {
	collectionSubmitService, err := services.CollectionSubmitService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, collectionSubmitService)
}

func CollectionDetail(c *gin.Context) {
	collectionDetailService, err := services.CollectionDetailService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, collectionDetailService)
}

func CollectionSubmitDetail(c *gin.Context) {
	collectionSubmitDetailService, err := services.CollectionSubmitDetailService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, collectionSubmitDetailService)
}

func CollectionCreate(c *gin.Context) {
	// 1. 接收并验证请求参数
	var creator models.Collection
	if err := c.ShouldBindJSON(&creator); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(400, "参数错误"+err.Error()))
		return
	}
	claims, _ := c.Get("claims")
	id := claims.(*models.CustomClaims).UserID
	creator.Founder = id

	// 2. 调用服务层创建方法
	if err := services.CreateCollectionService(creator); err != nil {
		// 根据错误类型返回不同状态码
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "参数校验失败") {
			c.JSON(http.StatusOK, models.Error(http.StatusBadRequest, "参数校验失败"+err.Error()))
			return
		}

		c.JSON(http.StatusOK, models.Error(statusCode, "创建失败"+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(creator))
}

func CollectionUpdate(c *gin.Context) {
	collectionUpdateService, err := services.UpdateCollectionService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, collectionUpdateService)
}

func CollectionDelete(c *gin.Context) {
	collectionDeleteService, err := services.DeleteCollectionService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, collectionDeleteService)
}
