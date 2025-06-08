package services

import (
	"encoding/json"
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
)

func OrgListService(c *gin.Context) (models.Result, error) {
	var orgList []models.Organization

	baseQuery := config.DB.Model(&models.Organization{}).Order("`sort` ASC")
	if err := baseQuery.
		Scopes(
			searchOrgParams(c)).
		Find(&orgList).
		Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"+err.Error()), err
	}

	// 转换为树形结构
	treeData := ConvertToTree(orgList)

	return models.Success(treeData), nil
}

func OrgUserListService() (models.Result, error) {
	var orgList []models.Organization
	baseQuery := config.DB.Model(&models.Organization{}).Order("`sort` ASC")
	if err := baseQuery.Find(&orgList).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"+err.Error()), err
	}
	// 转换为树形结构
	treeData := ConvertUserToTree(orgList)
	b, _ := json.MarshalIndent(treeData, "", "  ")
	log.Println(string(b))
	return models.Success(treeData), nil
}

// 接口：根据父节点获取子节点
func GetChildren(c *gin.Context) (models.Result, error) {
	parentID := c.Query("parent_id")
	var children []models.Organization
	config.DB.Where("parent_id = ?", parentID).Find(&children)

	return models.Success(children), nil
}

func UpdateOrg(c *gin.Context) (models.Result, error) {
	var org models.Organization
	if err := c.ShouldBindJSON(&org); err != nil {
		return models.Fail(http.StatusBadRequest, "参数错误"+err.Error()), err
	}
	if err := config.DB.Where("id = ?", org.ID).Updates(&org).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "更新失败"+err.Error()), err
	}
	return models.Success("更新成功"), nil
}

func CreateOrg(c *gin.Context) (models.Result, error) {
	var org models.Organization
	if err := c.ShouldBindJSON(&org); err != nil {
		return models.Fail(http.StatusBadRequest, "参数错误"+err.Error()), err
	}
	if err := config.DB.Create(&org).Error; err != nil {
		return models.Fail(http.StatusBadRequest, "创建失败"+err.Error()), err
	}
	return models.Success("创建成功"), nil
}

func DeleteOrg(c *gin.Context) (models.Result, error) {
	id := c.Param("id")
	var organizations []models.Organization
	if err := config.DB.Where("parent_id = ?", id).Find(&organizations).Error; err != nil {
		return models.Error(http.StatusBadRequest, "删除失败"+err.Error()), err
	}
	marshalIndent, err := json.MarshalIndent(organizations, "", "  ")
	if err != nil {
		return models.Result{}, err
	}
	log.Println("organizations is", string(marshalIndent))
	if len(organizations) > 0 {
		return models.Error(http.StatusBadRequest, "请先删除子部门"), nil
	}
	org := models.Organization{}
	config.DB.Where("id = ?", id).First(&org)
	// 删除本地/upload目录下的logo图片
	if org.OrgLogo != "" {
		err := os.Remove("." + org.OrgLogo)
		if err != nil {
			log.Println("删除失败" + err.Error())
			return models.Error(http.StatusBadRequest, "删除失败"+err.Error()), err
		}
	}
	if err := config.DB.Where("id = ?", id).Delete(&models.Organization{}).Error; err != nil {
		return models.Error(http.StatusBadRequest, "删除失败"), err
	}
	return models.Success("删除成功"), nil
}

func searchOrgParams(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		orgName := c.Query("org_name")
		if orgName != "" {
			db.Where("org_name LIKE ?", "%"+orgName+"%")
		}
		status := c.Query("status")
		if status != "" {
			db.Where("status = ?", status)
		}
		return db
	}

}

// ConvertToTree 根据org的id与parent_id转为树状结构
func ConvertToTree(orgs []models.Organization) []*models.Organization {
	// 创建 ID 到节点的映射
	orgMap := make(map[int]*models.Organization)
	for i := range orgs {
		org := &orgs[i]
		org.Children = []*models.Organization{} // 初始化空子节点列表
		orgMap[org.ID] = org
	}

	var roots []*models.Organization
	for _, org := range orgs {
		if org.ParentId == 0 {
			// 根节点
			roots = append(roots, orgMap[org.ID])
		} else {
			// 查找父节点并挂载
			if parent, exists := orgMap[org.ParentId]; exists {
				parent.Children = append(parent.Children, orgMap[org.ID])
			}
		}
	}
	return roots
}

func ConvertUserToTree(orgs []models.Organization) []*models.Organization {
	orgMap := make(map[int]*models.Organization)
	orgIDs := make([]int, 0, len(orgs))

	// 初始化组织映射并收集所有组织ID
	for i := range orgs {
		org := &orgs[i]
		org.Children = []*models.Organization{} // 初始化子节点
		org.Users = []*models.User{}            // 初始化用户列表
		orgMap[org.ID] = org
		orgIDs = append(orgIDs, org.ID)
	}

	// 批量查询所有相关用户
	var users []*models.User
	if err := config.DB.Where("org_id IN ?", orgIDs).Find(&users).Error; err != nil {
		log.Printf("Failed to fetch users: %v", err)
	}

	// 按org_id分组用户
	userMap := make(map[int][]*models.User)
	for _, user := range users {
		userMap[user.OrgId] = append(userMap[user.OrgId], user)
	}

	// 确保所有组织都有用户条目（即使为空）
	for _, id := range orgIDs {
		if _, exists := userMap[id]; !exists {
			userMap[id] = []*models.User{}
		}
	}

	// 分配用户到组织节点
	for id, org := range orgMap {
		org.Users = userMap[id]
	}

	// 构建树形结构
	var roots []*models.Organization
	for _, org := range orgs {
		if org.ParentId == 0 {
			roots = append(roots, orgMap[org.ID])
		} else {
			if parent, exists := orgMap[org.ParentId]; exists {
				parent.Children = append(parent.Children, orgMap[org.ID])
			}
		}
	}

	return roots
}
