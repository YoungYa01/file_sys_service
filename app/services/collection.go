package services

import (
	"encoding/json"
	"fmt"
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
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
		return fmt.Errorf("时间格式错误: %v", err)
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
				Nickname:     user.Nickname,
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

func TCCollectionListService(c *gin.Context) (models.Result, error) {
	var baseCollections []models.CollectionCreator
	var total int64

	// 第一阶段：获取基础分页数据
	baseQuery := config.DB.Model(&models.CollectionCreator{}).
		Order("pinned DESC,created_at DESC")

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
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func CollectionSubmitService(c *gin.Context) (models.Result, error) {
	var tcSubmitter models.TCSubmitter
	claims, _ := c.Get("claims")
	id := claims.(*models.CustomClaims).UserID
	userName := claims.(*models.CustomClaims).UserName

	var user models.User

	if err := config.DB.Model(&models.User{}).Where("id = ?", id).First(&user).Error; err != nil {
		return models.Error(http.StatusBadRequest, "用户不存在"), err
	}

	log.Println(id, userName)

	if err := c.ShouldBindJSON(&tcSubmitter); err != nil {
		return models.Error(http.StatusBadRequest, "参数错误"+err.Error()), err
	}

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var collection models.CollectionCreator
	var total int64
	if err := tx.Model(&models.CollectionCreator{}).
		Where("id = ?", tcSubmitter.CollectionID).
		First(&collection).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
	}

	tx.Model(&models.CollectionSubmitter{}).
		Where("collection_id = ? AND user_id = ?", tcSubmitter.CollectionID, id).
		Count(&total)

	if collection.Access != "some" && total <= int64(collection.FileNumber) {
		// 首先删掉原来的然后再创建新的
		if err := tx.Where("user_id = ? AND collection_id = ?", id, tcSubmitter.CollectionID).
			Delete(&models.CollectionSubmitter{}).Error; err != nil {
			tx.Rollback()
			return models.Fail(http.StatusInternalServerError, "删除失败"), err
		}

		var collectionSubmitters []models.CollectionSubmitter
		processCount := len(tcSubmitter.File)
		for index := 0; index < processCount; index++ {
			collectionSubmitters = append(collectionSubmitters, models.CollectionSubmitter{
				CollectionID: tcSubmitter.CollectionID,
				UserID:       uint(id),
				TaskStatus:   2,
				UserName:     userName,
				Nickname:     user.Nickname,
				SubmitTime:   time.Now(),
				Sort:         index + 1,
				FilePath:     tcSubmitter.File[index].FilePath,
				FileName:     tcSubmitter.File[index].FileName,
			})
		}
		if err := tx.Create(&collectionSubmitters).Error; err != nil {
			tx.Rollback()
			return models.Fail(http.StatusInternalServerError, "创建提交者失败"), err
		}
		return models.Success(fmt.Sprintf("成功提交%d个文件", processCount)), tx.Commit().Error
	} else {
		// 查询所有需要处理的记录
		var collectionSubmitters []models.CollectionSubmitter
		err := tx.Model(&models.CollectionSubmitter{}).
			Where("user_id = ? AND collection_id = ?", id, tcSubmitter.CollectionID).
			Order("sort ASC").
			Find(&collectionSubmitters).Error
		if err != nil {
			return models.Fail(http.StatusInternalServerError, "查询失败"), err
		}

		// 确定最小处理数量
		processCount := min(len(collectionSubmitters), len(tcSubmitter.File))

		// 只处理有对应文件的部分
		for index := 0; index < processCount; index++ {
			submitter := &collectionSubmitters[index]
			submitter.FilePath = tcSubmitter.File[index].FilePath
			submitter.FileName = tcSubmitter.File[index].FileName
			submitter.SubmitTime = time.Now()
			submitter.TaskStatus = 2
			submitter.Sort = index + 1

			if err := tx.Select("file_path", "file_name", "submit_time", "task_status").
				Where("id = ?", submitter.ID).
				Updates(submitter).Error; err != nil {
				tx.Rollback()
				return models.Fail(http.StatusInternalServerError, fmt.Sprintf("第%d个文件更新失败", index+1)), err
			}
		}

		// 自动跳过超出文件数量的记录
		return models.Success(fmt.Sprintf("成功提交%d个文件", processCount)), tx.Commit().Error
	}

}
func CollectionListService(c *gin.Context) (models.Result, error) {
	var baseCollections []models.CollectionCreator
	var total int64

	claims, _ := c.Get("claims")
	id := claims.(*models.CustomClaims).UserID

	// 第一阶段：获取基础分页数据
	baseQuery := config.DB.Model(&models.CollectionCreator{}).
		Where("founder = ?", id).
		Order("pinned DESC,created_at DESC")

	if err := baseQuery.Count(&total).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
	}

	if err := baseQuery.
		Scopes(
			searchCLParams(c),
			Paginate(c)).
		Find(&baseCollections).
		Error; err != nil {
		log.Println("查询失败", err)
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
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
		Founder              models.User                  `json:"founder"`
		Reviewers            []models.CollectionReviewer  `json:"reviewers"`
		Submitters           []models.CollectionSubmitter `json:"submitters"`
		CollectionSubmitters []models.CollectionSubmitter `json:"submitted_files"`
	}

	claims, _ := c.Get("claims")
	id := claims.(*models.CustomClaims).UserID

	var collection models.CollectionCreator
	if err := config.DB.Where("id = ?", collectionId).First(&collection).Error; err != nil {
		log.Println("id错误", err)
		return models.Fail(http.StatusInternalServerError, "id错误"+err.Error()), err
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

	var submitters []models.CollectionSubmitter

	config.DB.Model(&models.CollectionSubmitter{}).
		Order("sort ASC").
		Where("collection_id = ?", collectionId).
		Scan(&submitters)

	var collectionSubmitters []models.CollectionSubmitter
	err := config.DB.Model(&models.CollectionSubmitter{}).
		Where("user_id = ? AND collection_id = ?", id, collectionId). // 关键修复
		Order("sort ASC").
		Find(&collectionSubmitters).Error
	if err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
	}

	result := Details{
		CollectionCreator:    collection,
		Reviewers:            reviewers,
		Founder:              founders,
		Submitters:           submitters,
		CollectionSubmitters: collectionSubmitters,
	}
	return models.Success(result), nil
}

func CollectionSubmitDetailService(c *gin.Context) (models.Result, error) {
	type SubmitInfo struct {
		ID           uint      `json:"id"`
		ReviewStatus uint      `json:"review_status"`
		ReviewTime   time.Time `json:"review_time"`
		SubmitTime   time.Time `json:"submit_time"`
		FilePath     string    `json:"file_path"`
		FileName     string    `json:"file_name"`
		Recommend    string    `json:"recommend"`
		TaskStatus   uint      `json:"task_status"`
		Sort         int       `json:"sort"`
	}
	type CollectionSubmitterGroup struct {
		ID           uint         `json:"id"`
		CollectionID uint         `json:"collection_id"`
		UserID       uint         `json:"user_id"`
		UserName     string       `json:"user_name"`
		Nickname     string       `json:"nickname"`
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
		return models.Fail(http.StatusInternalServerError, "查询失败"+err.Error()), err
	}
	// 按照任务 ID、提交者 ID 和用户名分组
	groupedData := make(map[string]CollectionSubmitterGroup)
	for _, submitter := range collectionSubmitters {
		key := fmt.Sprintf("%d-%d-%s", submitter.CollectionID, submitter.UserID, submitter.UserName)
		if group, exists := groupedData[key]; exists {
			group.Submits = append(group.Submits, SubmitInfo{
				ID:           submitter.ID,
				ReviewStatus: submitter.ReviewStatus,
				ReviewTime:   submitter.ReviewTime,
				TaskStatus:   submitter.TaskStatus,
				SubmitTime:   submitter.SubmitTime,
				FilePath:     submitter.FilePath,
				FileName:     submitter.FileName,
				Recommend:    submitter.Recommend,
				Sort:         submitter.Sort,
			})
			groupedData[key] = group
		} else {
			groupedData[key] = CollectionSubmitterGroup{
				ID:           submitter.ID,
				CollectionID: submitter.CollectionID,
				UserID:       submitter.UserID,
				UserName:     submitter.UserName,
				Nickname:     submitter.Nickname,
				Submits: []SubmitInfo{
					{
						ID:           submitter.ID,
						ReviewStatus: submitter.ReviewStatus,
						ReviewTime:   submitter.ReviewTime,
						TaskStatus:   submitter.TaskStatus,
						SubmitTime:   submitter.SubmitTime,
						FilePath:     submitter.FilePath,
						FileName:     submitter.FileName,
						Recommend:    submitter.Recommend,
						Sort:         submitter.Sort,
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
		return models.Fail(http.StatusInternalServerError, "id错误"+err.Error()), err
	}

	if err := c.ShouldBindJSON(&collection); err != nil {
		marshalIndent, _ := json.MarshalIndent(collection, "", "  ")
		log.Println("参数错误", err, string(marshalIndent))
		return models.Fail(http.StatusBadRequest, "参数错误"+err.Error()), err
	}

	log.Println("collection is", collection)
	if err := config.DB.Save(&collection).Error; err != nil {
		log.Println("更新失败", err)
		return models.Fail(http.StatusInternalServerError, "更新失败"+err.Error()), err
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
		return models.Error(http.StatusBadRequest, "id错误"+err.Error()), err
	}
	if err := tx.Where("collection_id = ?", c.Param("id")).Delete(&models.CollectionSubmitter{}).Error; err != nil {
		log.Println("删除提交者失败", err)
		tx.Rollback()
		return models.Error(http.StatusInternalServerError, "删除提交者失败"+err.Error()), err
	}
	if err := tx.Where("collection_id = ?", c.Param("id")).Delete(&models.CollectionReviewer{}).Error; err != nil {
		log.Println("删除审核者失败", err)
		tx.Rollback()
		return models.Error(http.StatusInternalServerError, "删除审核者失败"+err.Error()), err
	}
	if err := tx.Where("id = ?", c.Param("id")).Delete(&collection).Error; err != nil {
		log.Println("删除失败", err)
		tx.Rollback()
		return models.Error(http.StatusInternalServerError, "删除失败"+err.Error()), err
	}
	if err := tx.Commit().Error; err != nil {
		log.Println("提交事务失败", err)
		tx.Rollback()
		return models.Error(http.StatusInternalServerError, "提交事务失败"+err.Error()), err
	}
	return models.Success(models.SuccessWithMsg("删除成功")), nil
}
