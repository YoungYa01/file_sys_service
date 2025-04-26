package services

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func searchByParams(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		username := c.Query("username") // 从URL参数获取title参数
		if username != "" {
			// 使用LIKE进行模糊查询，并防止SQL注入
			db.Where("username LIKE ?", "%"+username+"%")
		}
		role := c.Query("role")
		if role != "" {
			db.Where("role = ?", role)
		}
		status := c.Query("status")
		if status != "" {
			db.Where("status = ?", status)
		}
		nickname := c.Query("nickname")
		if nickname != "" {
			db.Where("nickname LIKE ?", "%"+nickname+"%")
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
		Scopes(searchByParams(c)).
		Select("users.*, roles.role_name, roles.description, roles.permission").
		Joins("LEFT JOIN roles ON users.role_id = roles.id").Order("`created_at` DESC")

	if err := baseQuery.Count(&total).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "获取总数失败"), err
	}

	if err := baseQuery.
		Scopes(Paginate(c)).
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

func UserDetailService(c *gin.Context) (models.Result, error) {
	var user models.User
	claims, _ := c.Get("claims")
	id := claims.(*models.CustomClaims).UserID
	if err := config.DB.Model(&models.User{}).
		Where("id = ?", id).
		First(&user).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "请检查id是否正确"), err
	}

	var hotmaps []models.CollectionSubmitter
	err := config.DB.Model(&models.CollectionSubmitter{}).
		Where("user_id = ?", id).
		Order("submit_time asc").
		Scan(&hotmaps).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
		return models.Fail(http.StatusInternalServerError, "数据库查询失败"), err
	}

	return models.Success(struct {
		User   models.User                  `json:"user"`
		HotMap []models.CollectionSubmitter `json:"hot_map"`
	}{
		User:   user,
		HotMap: hotmaps,
	}), nil
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

func UserUploadService(c *gin.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, models.Fail(http.StatusBadRequest, "请上传文件"))
		return err
	}
	if file.Header.Get("Content-Type") != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		c.JSON(http.StatusOK, models.Fail(http.StatusBadRequest, "请上传Excel文件"))
		return err
	}
	// 3. 保存临时文件
	tempPath := filepath.Join("uploads", file.Filename)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusOK, models.Fail(http.StatusBadRequest, "文件保存失败"))
		return err
	}
	// 4. 解析Excel数据
	f, err := excelize.OpenFile(tempPath)
	if err != nil {
		c.JSON(http.StatusOK, models.Fail(http.StatusBadRequest, "文件解析失败"))
		return err
	}
	defer f.Close()
	rows, err := f.GetRows("Sheet1")
	if err != nil || len(rows) < 2 {
		c.JSON(http.StatusOK, models.Fail(http.StatusBadRequest, "数据解析失败"))
		return err
	}
	var users []models.User
	for rowIdx, row := range rows[1:] {
		actualRow := rowIdx + 2 // Excel行号从1开始
		if len(row) < 3 {
			c.JSON(http.StatusOK, models.Fail(http.StatusBadRequest, "数据解析失败"))
			return err
		}
		nickname := row[0]
		workNo := row[1]
		pwd := row[2]
		depart := strings.TrimSpace(row[3])
		rl := strings.TrimSpace(row[4])
		log.Printf("第%d行数据：姓名=%s,工号=%s,密码=%s,部门=%s,角色=%s", actualRow, nickname, workNo, pwd, depart, rl)
		var role models.Role
		config.DB.Where("role_name = ?", rl).First(&role)
		if role.ID == 0 {
			continue
		}
		var organization models.Organization
		config.DB.Where("org_name = ?", depart).First(&organization)
		if organization.ID == 0 {
			continue
		}
		var u models.User
		config.DB.Model(&models.User{}).Where("username = ?", workNo).First(&u)
		if u.ID != 0 {
			continue
		}
		users = append(users, models.User{
			Nickname: nickname,
			Username: workNo,
			Password: pwd,
			RoleID:   role.ID,
			OrgId:    organization.ID,
		})
	}
	if len(users) == 0 {
		c.JSON(http.StatusOK, models.Success(users))
		return err
	}
	if err := config.DB.Create(&users).Error; err != nil {
		c.JSON(http.StatusOK, models.Fail(http.StatusBadRequest, "数据导入失败"))
		return err
	}

	c.JSON(http.StatusOK, models.Success(users))
	return nil
}
