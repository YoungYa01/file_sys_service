package controllers

import (
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

func MyTaskList(c *gin.Context) {
	taskListService, err := services.MyTaskListService(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, taskListService)
}
