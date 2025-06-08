package services

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func searchRoleList(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		role := c.Query("role") // 从URL参数获取title参数
		if role != "" {
			// 使用LIKE进行模糊查询，并防止SQL注入
			return db.Where("role LIKE ?", "%"+role+"%")
		}
		status := c.Query("status")
		if status != "" {
			return db.Where("status = ?", status)
		}
		return db
	}
}

func RoleListService(c *gin.Context) (models.Result, error) {
	var roleList []models.Role
	var total int64
	baseQuery := config.DB.Model(&models.Role{}).Order("`sort` DESC")

	if err := baseQuery.Scopes(
		searchRoleList(c),
		Paginate(c)).
		Find(&roleList).
		Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"+err.Error()), err
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "获取总数失败"+err.Error()), err
	}
	pagination := models.PaginationResponse{
		Page:     c.GetInt("page"),
		PageSize: c.GetInt("pageSize"),
		Total:    total,
		Data:     roleList,
	}
	return models.Success(pagination), nil
}

func RoleCreateService(c *gin.Context) (models.Result, error) {
	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		log.Println("RoleCreateService err:", err)
		return models.Fail(http.StatusBadRequest, "参数错误"+err.Error()), err
	}
	// 查重
	if err := config.DB.Where("role = ?", role.RoleName).First(&role).Error; err == nil {
		return models.Fail(http.StatusBadRequest, "角色已存在"), nil
	}
	if role.RoleName == "" || role.Permission == "" || role.Status == "" || role.Sort == 0 {
		return models.Fail(http.StatusBadRequest, "参数错误"), nil
	}
	if err := config.DB.Create(&role).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "创建失败"+err.Error()), err
	}
	return models.Success("创建成功"), nil
}

func RoleUpdateService(c *gin.Context) (models.Result, error) {
	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		return models.Fail(http.StatusBadRequest, "参数错误"), err
	}
	if role.RoleName == "" || role.Permission == "" || role.Status == "" || role.Sort == 0 {
		return models.Fail(http.StatusBadRequest, "参数错误"), nil
	}
	if err := config.DB.Where("id = ?", role.ID).Updates(&role).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "更新失败"), err
	}
	return models.Success("更新成功"), nil
}

func RoleDeleteService(c *gin.Context) (models.Result, error) {
	var roleId = c.Param("id")
	var role models.Role
	if err := config.DB.Where("id = ?", roleId).First(&role).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "删除失败"+err.Error()), err
	}
	if err := config.DB.Where("id = ?", roleId).Delete(&role).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "删除失败"+err.Error()), err
	}
	return models.Success("删除成功"), nil
}
