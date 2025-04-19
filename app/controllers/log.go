package controllers

import (
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

func LogList(c *gin.Context) {
	logListService, err := services.LogListService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, logListService)
}

func DeleteLog(c *gin.Context) {
	logDeleteService, err := services.LogDeleteService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, logDeleteService)
}
