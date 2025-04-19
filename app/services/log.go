package services

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func searchLogParams(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		userId := c.Query("user_id")
		if userId != "" {
			db.Where("user_id = ?", userId)
		}
		username := c.Query("username")
		if username != "" {
			db.Where("user_name = ?", username)
		}
		ip := c.Query("ip")
		if ip != "" {
			db.Where("ip = ?", ip)
		}
		os := c.Query("os")
		if os != "" {
			db.Where("os LIKE ?", "%"+os+"%")
		}
		method := c.Query("method")
		if method != "" {
			db.Where("method = ?", method)
		}
		apiUrl := c.Query("api_url")
		if apiUrl != "" {
			db.Where("api_url LIKE ?", "%"+apiUrl+"%")
		}
		return db
	}
}
func LogListService(c *gin.Context) (models.Result, error) {
	var logList []models.Log
	var total int64

	baseQuery := config.DB.Model(&models.Log{}).
		Order("`created_at` DESC")

	if err := baseQuery.
		Count(&total).
		Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
	}

	if err := baseQuery.
		Scopes(
			searchLogParams(c),
			Paginate(c)).
		Find(&logList).
		Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
	}
	pagination := models.PaginationResponse{
		Page:     c.GetInt("page"),
		PageSize: c.GetInt("pageSize"),
		Total:    total,
		Data:     logList,
	}
	return models.Success(pagination), nil
}

func LogDeleteService(c *gin.Context) (models.Result, error) {
	var log models.Log
	id := c.Param("id")
	if err := config.DB.Where("id = ?", id).Delete(&log).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "删除失败"), err
	}
	return models.Success(models.SuccessWithMsg("删除成功")), nil
}
