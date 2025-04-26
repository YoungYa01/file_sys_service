package services

import (
	"archive/zip"
	"fmt"
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func reviewListByParams(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		title := c.Query("title")
		if title != "" {
			db.Where("title LIKE ?", "%"+title+"%")
		}
		return db
	}
}

func ReviewListService(c *gin.Context) (models.Result, error) {
	claims, _ := c.Get("claims")
	userID := claims.(*models.CustomClaims).UserID

	var reviewers []models.CollectionReviewer
	var total int64
	// 先统计total
	if err := config.DB.Model(&models.CollectionReviewer{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
	}
	// 然后分页获取数据
	if err := config.DB.Model(&models.CollectionReviewer{}).
		Where("user_id = ?", userID).
		Scopes(Paginate(c)).
		Find(&reviewers).
		Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
	}
	var collections []models.CollectionCreator
	title := c.Query("title")
	for _, r := range reviewers {
		var collection models.CollectionCreator
		config.DB.Model(&models.CollectionCreator{}).
			Where("id = ?", r.CollectionID).
			First(&collection)
		config.DB.Model(&models.CollectionReviewer{}).
			Where("collection_id = ?", r.CollectionID).
			Order("review_order").
			Find(&collection.Reviewers)
		// 模糊，如果包含title就算
		if strings.Contains(collection.Title, title) {
			collections = append(collections, collection)
		}
	}
	pagination := models.PaginationResponse{
		Page:     c.GetInt("page"),
		PageSize: c.GetInt("pageSize"),
		Total:    total,
		Data:     collections,
	}
	return models.Success(pagination), nil
}

func ReviewDetailListService(c *gin.Context) (models.Result, error) {
	type SubmitInfo struct {
		ID           uint      `json:"id"`
		SubmitTime   time.Time `json:"submit_time"`
		FilePath     string    `json:"file_path"`
		FileName     string    `json:"file_name"`
		Recommend    string    `json:"recommend"`
		TaskStatus   uint      `json:"task_status"`
		ReviewStatus uint      `json:"review_status"`
		ReviewTime   time.Time `json:"review_time"`
		Sort         int       `json:"sort"`
	}
	type CollectionSubmitterGroup struct {
		CollectionID uint         `json:"collection_id"`
		UserID       uint         `json:"user_id"`
		UserName     string       `json:"user_name"`
		Nickname     string       `json:"nickname"`
		Submits      []SubmitInfo `json:"submits"`
	}

	var collectionId = c.Param("id")

	var collectionSubmitters []models.CollectionSubmitter

	baseQuery := config.DB.Model(&models.CollectionSubmitter{}).
		Where("collection_id = ?", collectionId)

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
				TaskStatus:   submitter.TaskStatus,
				ReviewStatus: submitter.ReviewStatus,
				ReviewTime:   submitter.ReviewTime,
				SubmitTime:   submitter.SubmitTime,
				FilePath:     submitter.FilePath,
				FileName:     submitter.FileName,
				Recommend:    submitter.Recommend,
				Sort:         submitter.Sort,
			})
			groupedData[key] = group
		} else {
			groupedData[key] = CollectionSubmitterGroup{
				CollectionID: submitter.CollectionID,
				UserID:       submitter.UserID,
				UserName:     submitter.UserName,
				Nickname:     submitter.Nickname,
				Submits: []SubmitInfo{
					{
						ID:           submitter.ID,
						TaskStatus:   submitter.TaskStatus,
						SubmitTime:   submitter.SubmitTime,
						ReviewStatus: submitter.ReviewStatus,
						ReviewTime:   submitter.ReviewTime,
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

	// 根据user_name过滤
	if username := c.Query("user_name"); username != "" {
		var filteredResult []CollectionSubmitterGroup
		for _, group := range result {
			if strings.Contains(group.UserName, username) {
				filteredResult = append(filteredResult, group)
			}
		}
		result = filteredResult
	}

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

func ReviewUpdateStatusService(c *gin.Context) (models.Result, error) {
	var collectionSubmitter models.CollectionSubmitter
	if err := c.ShouldBindJSON(&collectionSubmitter); err != nil {
		return models.Fail(http.StatusBadRequest, "参数错误"), err
	}
	if err := config.DB.Model(&models.CollectionSubmitter{}).
		Where("id = ?", collectionSubmitter.ID).
		Update("review_status", collectionSubmitter.ReviewStatus).
		Update("recommend", collectionSubmitter.Recommend).
		Update("review_time", time.Now()).
		Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "更新失败"), err
	}
	return models.Success("更新成功"), nil
}

func ReviewExportService(c *gin.Context) (models.Result, error) {
	ids := c.Query("ids")
	log.Println("ids", ids)
	if ids == "" {
		return models.Fail(http.StatusBadRequest, "ids 参数不能为空"), nil
	}

	var idList []uint
	for _, idStr := range strings.Split(ids, ",") {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return models.Fail(http.StatusBadRequest, "ids 参数格式错误"), nil
		}
		idList = append(idList, uint(id))
	}

	files, err := getFilesByIDs(idList)
	if err != nil {
		return models.Fail(http.StatusInternalServerError, "查询文件失败"), err
	}

	// 创建临时压缩文件
	zippedFileName := "download_" + time.Now().Format("20060102150405") + ".zip"
	zippedFilePath := filepath.Join(os.TempDir(), zippedFileName)

	zipFile, err := os.Create(zippedFilePath)
	if err != nil {
		return models.Fail(http.StatusInternalServerError, "创建压缩包失败"), err
	}
	defer zipFile.Close()
	defer os.Remove(zippedFilePath) // 确保最终删除临时文件

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	for _, file := range files {
		getwd, _ := os.Getwd()
		filePath := filepath.FromSlash(filepath.Join(getwd, file.FilePath))
		fileInZip := file.FileName

		log.Println("filePath", filePath)
		if file.FilePath == "" {
			continue
		}

		err := func() error {
			fileToZip, err := os.Open(filePath)
			if err != nil {
				log.Println("fileToZip err is: ", err)
				return err
			}
			defer fileToZip.Close()

			stat, err := fileToZip.Stat()
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(stat)
			if err != nil {
				return err
			}
			header.Name = fileInZip

			writerFile, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}

			_, err = io.Copy(writerFile, fileToZip)
			return err
		}()

		if err != nil {
			return models.Fail(http.StatusInternalServerError, "添加文件到压缩包失败: "+err.Error()), err
		}
	}

	// 必须关闭writer和zipFile以确保所有数据写入磁盘
	writer.Close()
	zipFile.Close()

	// 设置响应头
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zippedFileName))
	c.Header("Content-Transfer-Encoding", "binary")

	// 发送文件并终止后续处理
	c.File(zippedFilePath)

	// 返回空Result避免框架追加JSON响应
	return models.Result{}, nil
}

type File struct {
	ID       uint   `json:"id"`
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
}

func getFilesByIDs(ids []uint) ([]File, error) {
	var files []File
	if err := config.DB.Model(&models.CollectionSubmitter{}).
		Where("id IN ?", ids).
		Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}
