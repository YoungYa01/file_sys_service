package controllers

import (
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ReviewList(c *gin.Context) {
	listService, err := services.ReviewListService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, listService)
}

func ReviewDetailList(c *gin.Context) {
	listService, err := services.ReviewDetailListService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, listService)
}

func ReviewStatus(c *gin.Context) {
	listService, err := services.ReviewUpdateStatusService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, listService)
}

func ReviewExport(c *gin.Context) {
	exportService, err := services.ReviewExportService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, exportService)
}
