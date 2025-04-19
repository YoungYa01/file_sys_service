package services

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func NotificationCreateService(c *gin.Context) (models.Result, error) {
	var notification models.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		return models.Fail(http.StatusBadRequest, "参数错误"), err
	}
	if notification.Title == "" || notification.Content == "" {
		return models.Fail(http.StatusBadRequest, "参数错误"), nil
	}
	if err := config.DB.Create(&notification).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "创建失败"), err
	}
	return models.Success("创建成功"), nil
}

func searchNotificationByParams(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		title := c.Query("title") // 从URL参数获取title参数
		if title != "" {
			// 使用LIKE进行模糊查询，并防止SQL注入
			return db.Where("title LIKE ?", "%"+title+"%")
		}
		founder := c.Query("founder")
		if founder != "" {
			return db.Where("founder = ?", founder)
		}
		pinned := c.Query("pinned")
		if pinned != "" {
			return db.Where("pinned = ?", pinned)
		}
		return db
	}
}

func NotificationListService(c *gin.Context) (models.Result, error) {
	type NotificationWithUser struct {
		models.Notification
		UserName string `json:"user_name"`
		UserID   int    `json:"user_id"`
	}
	var notification []NotificationWithUser
	var total int64

	baseQuery := config.DB.Model(&models.Notification{}).
		Select("notification.*, users.username as user_name, users.id as user_id").
		Joins("left join users on notification.founder = users.id").
		Order("pinned DESC, created_at DESC")

	if err := baseQuery.Count(&total).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "获取总数失败"), err
	}

	if err := baseQuery.
		Scopes(
			searchNotificationByParams(c),
			Paginate(c)).
		Find(&notification).
		Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
	}

	pagination := models.PaginationResponse{
		Page:     c.GetInt("page"),
		PageSize: c.GetInt("pageSize"),
		Total:    total,
		Data:     notification,
	}
	return models.Success(pagination), nil
}

func NotificationUpdateService(c *gin.Context) (models.Result, error) {
	var notification models.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		return models.Fail(http.StatusBadRequest, "参数错误"), err
	}
	if notification.ID == 0 {
		return models.Fail(http.StatusBadRequest, "参数错误"), nil
	}
	if err := config.DB.Where("id = ?", notification.ID).Updates(&notification).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "更新失败"), err
	}
	return models.Success("更新成功"), nil
}

func NotificationDeleteService(c *gin.Context) (models.Result, error) {
	var notification models.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		return models.Fail(http.StatusBadRequest, "参数错误"), err
	}
	if notification.ID == 0 {
		return models.Fail(http.StatusBadRequest, "参数错误"), nil
	}
	if err := config.DB.Where("id = ?", notification.ID).Delete(&notification).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "删除失败"), err
	}
	return models.Success("删除成功"), nil
}
