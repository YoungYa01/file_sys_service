package services

import (
	"encoding/json"
	"fmt"
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"time"
)

func searchCLParams(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		title := c.Query("title")
		if title != "" {
			db.Where("title LIKE ?", "%"+title+"%")
		}
		fileType := c.Query("file_type")
		if fileType != "" {
			db.Where("file_type LIKE ?", "%"+fileType+"%")
		}
		access := c.Query("access")
		if access != "" {
			db.Where("access = ?", access)
		}
		founder := c.Query("founder")
		if founder != "" {
			db.Where("founder = ?", founder)
		}
		status := c.Query("status")
		if status != "" {
			db.Where("status = ?", status)
		}
		return db
	}
}

func CreateCollectionService(creator models.Collection) error {
	// 开启事务
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建主记录
	collection := models.CollectionCreator{
		Title:           creator.Title,
		Content:         creator.Content,
		FileType:        creator.FileType,
		Access:          creator.Access,
		AccessPwd:       creator.AccessPwd,
		FileNumber:      creator.FileNumber,
		Founder:         creator.Founder,
		Status:          creator.Status,
		Pinned:          creator.Pinned,
		SubmittedNumber: 0, // 初始化已提交数量
		TotalNumber:     len(creator.Submitters),
		EndTime:         creator.EndTime,
		CreatedAt:       time.Now(),
	}

	if err := tx.Create(&collection).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建主记录失败: %v", err)
	}
	// 必须刷新对象获取自增ID
	var newCollection models.Collection
	tx.Where("id = ?", collection.ID).First(&newCollection)

	// 提交者处理
	for _, submitterID := range creator.Submitters {
		// 查询用户的信息，根据submitterID
		var user models.User
		if err := tx.Where("id = ?", submitterID).First(&user).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("查询用户信息失败: %v (用户ID: %d)", err, submitterID)
		}
		for i := range creator.FileNumber {
			submitter := models.CollectionSubmitter{
				CollectionID: collection.ID, // 使用正确的uint类型
				UserID:       uint(submitterID),
				TaskStatus:   1, // 根据你的定义设置默认值
				UserName:     user.Username,
				Sort:         i + 1, //可能要提交多个文件
			}

			if err := tx.Create(&submitter).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("创建提交者失败: %v (用户ID: %d)", err, submitterID)
			}
		}
	}

	// 审核者处理
	for orderIdx, reviewerID := range creator.Reviewers {
		// 查询用户的信息，根据submitterID
		var user models.User
		if err := tx.Where("id = ?", reviewerID).First(&user).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("查询用户信息失败: %v (用户ID: %d)", err, reviewerID)
		}
		reviewer := models.CollectionReviewer{
			CollectionID: collection.ID, // 使用正确的uint类型
			UserID:       uint(reviewerID),
			ReviewOrder:  orderIdx + 1,
			UserName:     user.Username,
		}
		if err := tx.Create(&reviewer).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("创建审核者失败: %v (用户ID: %d)", err, reviewerID)
		}
	}

	return tx.Commit().Error
}

func CollectionListService(c *gin.Context) (models.Result, error) {
	var baseCollections []models.CollectionCreator
	var total int64

	// 第一阶段：获取基础分页数据
	baseQuery := config.DB.Model(&models.CollectionCreator{})

	if err := baseQuery.
		Scopes(
			searchCLParams(c),
			Paginate(c)).
		Count(&total).
		Find(&baseCollections).
		Error; err != nil {
		log.Println("查询失败", err)
		return models.Fail(500, "查询失败"), err
	}

	log.Println("total is: ", total)
	log.Println("baseCollections is: ", baseCollections)

	// 第二阶段：批量获取关联数据
	collectionIDs := make([]uint, len(baseCollections))
	for i, c := range baseCollections {
		collectionIDs[i] = c.ID
	}
	log.Println("collectionIDs is: ", collectionIDs)

	// 获取创始人信息
	var founders []models.User

	config.DB.Model(&models.CollectionCreator{}).
		Select("collections.id, users.username").
		Joins("LEFT JOIN users ON collections.founder = users.id").
		Where("collections.id IN (?)", collectionIDs).
		Scan(&founders)

	log.Println("founders is: ", founders)

	//// 获取提交者信息（同前）
	//var submitters []models.CollectionSubmitter
	//
	//config.DB.Model(&models.CollectionSubmitter{}).
	//	Select("collection_submitters.collection_id, users.id, users.username").
	//	Joins("LEFT JOIN users ON collection_submitters.user_id = users.id").
	//	Where("collection_submitters.collection_id IN (?)", collectionIDs).
	//	Scan(&submitters)
	//
	//log.Println("submitters is: ", submitters)

	// 获取审核者信息（同前）
	var reviewers []models.CollectionReviewer

	config.DB.Model(&models.CollectionReviewer{}).
		Select("collection_reviewers.collection_id, users.id, users.username").
		Joins("LEFT JOIN users ON collection_reviewers.user_id = users.id").
		Where("collection_reviewers.collection_id IN (?)", collectionIDs).
		Scan(&reviewers)

	log.Println("reviewers is: ", reviewers)

	// 构建最终响应
	result := make([]struct {
		models.CollectionCreator
		Founder struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
		}
	}, len(baseCollections))
	for i, c := range baseCollections {
		result[i].CollectionCreator = c
		result[i].Reviewers = findReviewers(c.ID, reviewers)
		//result[i].Submitters = findSubmitters(c.ID, submitters)
		f := findFounder(c.Founder, founders)
		result[i].Founder = struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
		}{Name: f.Username, ID: f.ID}
	}

	log.Println("result is: ", result)

	return models.Success(models.PaginationResponse{
		Page:     c.GetInt("page"),
		PageSize: c.GetInt("pageSize"),
		Total:    total,
		Data:     result,
	}), nil
}

func findReviewers(id uint, reviewers []models.CollectionReviewer) []models.CollectionReviewer {
	var result []models.CollectionReviewer
	for _, r := range reviewers {
		if r.CollectionID == id {
			result = append(result, r)
		}
	}
	return result
}
func findFounder(founderID int, founders []models.User) models.User {
	for _, f := range founders {
		if f.ID == founderID {
			return f
		}
	}
	return models.User{}
}

func UpdateCollectionService(c *gin.Context) (models.Result, error) {
	var collection models.CollectionCreator
	if err := config.DB.Where("id = ?", c.Param("id")).First(&collection).Error; err != nil {
		log.Println("id错误", err)
		return models.Fail(500, "id错误"+err.Error()), err
	}

	if err := c.ShouldBindJSON(&collection); err != nil {
		marshalIndent, _ := json.MarshalIndent(collection, "", "  ")
		log.Println("参数错误", err, string(marshalIndent))
		return models.Fail(400, "参数错误"+err.Error()), err
	}

	log.Println("collection is", collection)
	if err := config.DB.Save(&collection).Error; err != nil {
		log.Println("更新失败", err)
		return models.Fail(500, "更新失败"+err.Error()), err
	}
	return models.Success(models.SuccessWithMsg("更新成功")), nil
}

func DeleteCollectionService(c *gin.Context) (models.Result, error) {
	// 开启事务
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var collection models.CollectionCreator

	if err := tx.Where("id = ?", c.Param("id")).First(&collection).Error; err != nil {
		log.Println("id错误", err)
		tx.Rollback()
		return models.Error(400, "id错误"+err.Error()), err
	}
	if err := tx.Where("collection_id = ?", c.Param("id")).Delete(&models.CollectionSubmitter{}).Error; err != nil {
		log.Println("删除提交者失败", err)
		tx.Rollback()
		return models.Error(500, "删除提交者失败"+err.Error()), err
	}
	if err := tx.Where("collection_id = ?", c.Param("id")).Delete(&models.CollectionReviewer{}).Error; err != nil {
		log.Println("删除审核者失败", err)
		tx.Rollback()
		return models.Error(500, "删除审核者失败"+err.Error()), err
	}
	if err := tx.Where("id = ?", c.Param("id")).Delete(&collection).Error; err != nil {
		log.Println("删除失败", err)
		tx.Rollback()
		return models.Error(500, "删除失败"+err.Error()), err
	}
	if err := tx.Commit().Error; err != nil {
		log.Println("提交事务失败", err)
		tx.Rollback()
		return models.Error(500, "提交事务失败"+err.Error()), err
	}
	return models.Success(models.SuccessWithMsg("删除成功")), nil
}
