package controllers

import (
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
)

func NotificationList(c *gin.Context) {
	notificationCreateService, err := services.NotificationListService(c)
	if err != nil {
		return
	}
	c.JSON(200, notificationCreateService)
}

func NotificationCreate(c *gin.Context) {
	notificationCreateService, err := services.NotificationCreateService(c)
	if err != nil {
		return
	}
	c.JSON(200, notificationCreateService)
}

func NotificationUpdate(c *gin.Context) {
	notificationCreateService, err := services.NotificationUpdateService(c)
	if err != nil {
		return
	}
	c.JSON(200, notificationCreateService)
}

func NotificationDelete(c *gin.Context) {
	notificationCreateService, err := services.NotificationDeleteService(c)
	if err != nil {
		return
	}
	c.JSON(200, notificationCreateService)
}
