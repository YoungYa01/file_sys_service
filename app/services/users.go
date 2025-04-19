package services

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func searchByParams(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		username := c.Query("username") // 从URL参数获取title参数
		if username != "" {
			// 使用LIKE进行模糊查询，并防止SQL注入
			return db.Where("username LIKE ?", "%"+username+"%")
		}
		role := c.Query("role")
		if role != "" {
			return db.Where("role = ?", role)
		}
		status := c.Query("status")
		if status != "" {
			return db.Where("status = ?", status)
		}
		return db
	}
}

func UserListService(c *gin.Context) (models.Result, error) {
	type UserOrg struct {
		OrgName     string `json:"org_name"`
		OrgLogo     string `json:"org_logo"`
		Leader      string `json:"leader"`
		Description string `json:"description"`
	}
	type UserWithRole struct {
		models.User
		UserOrg
		RoleName       string `json:"role_name"`
		Description    string `json:"role_description"`
		RolePermission string `json:"role_permission"`
	}

	var userList []UserWithRole
	var total int64

	baseQuery := config.DB.Model(&models.User{}).
		Select("users.*, roles.role_name, roles.description, roles.permission").
		Joins("LEFT JOIN roles ON users.role_id = roles.id").Order("`created_at` DESC")

	if err := baseQuery.Count(&total).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "获取总数失败"), err
	}

	if err := baseQuery.
		Scopes(
			searchByParams(c),
			Paginate(c)).
		Find(&userList).
		Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
	}

	for u := range userList {
		userList[u].Password = ""
		config.DB.
			Model(&models.Organization{}).
			Where("id = ?", userList[u].OrgId).
			First(&userList[u].UserOrg)
	}

	pagination := models.PaginationResponse{
		Page:     c.GetInt("page"),
		PageSize: c.GetInt("pageSize"),
		Total:    total,
		Data:     userList,
	}
	return models.Success(pagination), nil
}

func UserCreateService(c *gin.Context) (models.Result, error) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		return models.Fail(http.StatusBadRequest, "参数错误"), err
	}
	if user.Username == "" || user.Password == "" || user.RoleID == 0 {
		return models.Fail(http.StatusBadRequest, "参数错误"), nil
	}
	if err := config.DB.Create(&user).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "创建失败"), err
	}
	return models.Success(models.SuccessWithMsg("创建成功")), nil
}

func UserUpdateService(c *gin.Context) (models.Result, error) {
	var user models.User
	id := c.Param("id")
	if err := config.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "请检查id是否正确"), err
	}
	if err := c.ShouldBindJSON(&user); err != nil {
		return models.Fail(http.StatusBadRequest, "参数错误"), err
	}
	if user.ID == 0 || user.Username == "" || user.RoleID == 0 {
		return models.Fail(http.StatusBadRequest, "参数错误"), nil
	}
	if err := config.DB.Where("id = ?", user.ID).Updates(&user).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "更新失败"), err
	}
	return models.Success("更新成功"), nil
}

func UserDeleteService(c *gin.Context) (models.Result, error) {
	var user models.User
	id := c.Param("id")
	if err := config.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "请检查id是否正确"), err
	}
	if err := config.DB.Where("id = ?", user.ID).Delete(&user).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "删除失败"), err
	}
	return models.Success("删除成功"), nil
}
