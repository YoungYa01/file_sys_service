package services

import (
	"encoding/json"
	"fmt"
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"sort"
	"strconv"
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
	const timeLayout = "2006-01-02 15:04:05"
	endTime, err := time.ParseInLocation(timeLayout, creator.EndTime, time.Local)
	if err != nil {
		return fmt.Errorf("时间格式错误，请使用 YYYY-MM-DD HH:mm:ss 格式")
	}
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
		EndTime:         endTime,
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
	baseQuery := config.DB.Model(&models.CollectionCreator{}).
		Order("created_at DESC")

	if err := baseQuery.Count(&total).Error; err != nil {
		return models.Fail(500, "查询失败"), err
	}

	if err := baseQuery.
		Scopes(
			searchCLParams(c),
			Paginate(c)).
		Find(&baseCollections).
		Error; err != nil {
		log.Println("查询失败", err)
		return models.Fail(500, "查询失败"), err
	}

	collectionIDs := make([]uint, len(baseCollections))
	for i, c := range baseCollections {
		collectionIDs[i] = c.ID
	}

	result := make([]struct {
		models.CollectionCreator
		Founder models.User
	}, len(baseCollections))
	for i, c := range baseCollections {
		result[i].CollectionCreator = c
		var reviewers []models.CollectionReviewer
		config.DB.Model(&models.CollectionReviewer{}).
			Where("collection_reviewers.collection_id = ?", c.ID).
			Scan(&reviewers)
		result[i].Reviewers = reviewers

		var f models.User
		config.DB.Model(&models.User{}).
			Where("id = ?", c.Founder).
			First(&f)
		if f.ID != 0 {
			result[i].Founder = f
		}
	}

	return models.Success(models.PaginationResponse{
		Page:     c.GetInt("page"),
		PageSize: c.GetInt("pageSize"),
		Total:    total,
		Data:     result,
	}), nil
}

func CollectionDetailService(c *gin.Context) (models.Result, error) {
	collectionId := c.Param("id")
	type Details struct {
		models.CollectionCreator
		Founder   models.User                 `json:"founder"`
		Reviewers []models.CollectionReviewer `json:"reviewers"`
	}
	var collection models.CollectionCreator
	if err := config.DB.Where("id = ?", collectionId).First(&collection).Error; err != nil {
		log.Println("id错误", err)
		return models.Fail(500, "id错误"+err.Error()), err
	}
	var founders models.User

	config.DB.Model(&models.CollectionCreator{}).
		Select("collections.id, users.username").
		Joins("LEFT JOIN users ON collections.founder = users.id").
		Where("collections.id = ?", collectionId).
		First(&founders)

	var reviewers []models.CollectionReviewer

	config.DB.Model(&models.CollectionReviewer{}).
		Order("review_order ASC").
		Where("collection_id = ?", collectionId).
		Scan(&reviewers)

	result := Details{
		CollectionCreator: collection,
		Reviewers:         reviewers,
		Founder:           founders,
	}
	return models.Success(result), nil
}

func CollectionSubmitDetailService(c *gin.Context) (models.Result, error) {
	type SubmitInfo struct {
		SubmitTime time.Time `json:"submit_time"`
		FilePath   string    `json:"file_path"`
		FileName   string    `json:"file_name"`
		Recommend  string    `json:"recommend"`
		TaskStatus uint      `json:"task_status"`
		Sort       int       `json:"sort"`
	}
	type CollectionSubmitterGroup struct {
		ID           uint         `json:"id"`
		CollectionID uint         `json:"collection_id"`
		UserID       uint         `json:"user_id"`
		UserName     string       `json:"user_name"`
		Submits      []SubmitInfo `json:"submits"`
	}

	var collectionId = c.Param("id")
	var submitter = c.Query("user_name")
	var collectionSubmitters []models.CollectionSubmitter
	baseQuery := config.DB.Model(&models.CollectionSubmitter{}).
		Where("collection_id = ?", collectionId)

	if submitter != "" {
		baseQuery = baseQuery.Where("user_name LIKE ?", "%"+submitter+"%")
	}

	if err := baseQuery.Find(&collectionSubmitters).Error; err != nil {
		log.Println("查询失败", err)
		return models.Fail(500, "查询失败"+err.Error()), err
	}
	// 按照任务 ID、提交者 ID 和用户名分组
	groupedData := make(map[string]CollectionSubmitterGroup)
	for _, submitter := range collectionSubmitters {
		key := fmt.Sprintf("%d-%d-%s", submitter.CollectionID, submitter.UserID, submitter.UserName)
		if group, exists := groupedData[key]; exists {
			group.Submits = append(group.Submits, SubmitInfo{
				TaskStatus: submitter.TaskStatus,
				SubmitTime: submitter.SubmitTime,
				FilePath:   submitter.FilePath,
				FileName:   submitter.FileName,
				Recommend:  submitter.Recommend,
				Sort:       submitter.Sort,
			})
			groupedData[key] = group
		} else {
			groupedData[key] = CollectionSubmitterGroup{
				ID:           submitter.ID,
				CollectionID: submitter.CollectionID,
				UserID:       submitter.UserID,
				UserName:     submitter.UserName,
				Submits: []SubmitInfo{
					{
						TaskStatus: submitter.TaskStatus,
						SubmitTime: submitter.SubmitTime,
						FilePath:   submitter.FilePath,
						FileName:   submitter.FileName,
						Recommend:  submitter.Recommend,
						Sort:       submitter.Sort,
					},
				},
			}
		}
	}
	// 将分组后的数据转换为切片
	var result []CollectionSubmitterGroup
	for _, group := range groupedData {
		result = append(result, group)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].UserID < result[j].UserID
	})

	// 获取分页参数
	current, _ := strconv.Atoi(c.Query("current"))
	if current <= 0 {
		current = 1
	}

	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	if pageSize <= 0 {
		pageSize = 10 // 默认每页显示10条
	}
	// 计算分页
	totalPages := int64(len(result))
	startIndex := (current - 1) * pageSize
	endIndex := current * pageSize

	if startIndex > int(totalPages) {
		startIndex = int(totalPages)
	}
	if endIndex > int(totalPages) {
		endIndex = int(totalPages)
	}

	// 分页后的数据
	pagedResult := result[startIndex:endIndex]

	pagination := models.PaginationResponse{
		Page:     current,
		PageSize: pageSize,
		Total:    totalPages,
		Data:     pagedResult,
	}
	return models.Success(pagination), nil
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
