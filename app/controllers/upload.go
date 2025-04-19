package controllers

import (
	"errors"
	"fmt"
	"gin_back/app/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//func Upload(c *gin.Context) {
//	file, err := c.FormFile("file")
//	if err != nil {
//		c.JSON(500, models.Error(400, "文件为空，上传失败"))
//		return
//	}
//	err = c.SaveUploadedFile(file, "./upload/"+file.Filename)
//	if err != nil {
//		c.JSON(500, models.Error(500, "上传失败"))
//		return
//	}
//	c.JSON(200, models.Success("/upload/"+file.Filename))
//}

func Upload(c *gin.Context) {
	// 配置参数（建议放在全局配置中）
	const (
		maxUploadSize = 50 << 20 // 50MB
		uploadDir     = "./upload/"
		publicURL     = "/upload/"
	)

	// 限制上传大小
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	file, err := c.FormFile("file")
	if err != nil {
		handleUploadError(c, err)
		return
	}

	// 验证文件类型
	if !isSafeFileType(file.Filename) {
		c.JSON(http.StatusOK, models.Error(400, "不支持的文件类型"))
		return
	}

	// 创建上传目录（跨平台兼容方式）
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("创建目录失败: %v", err)
		c.JSON(http.StatusOK, models.Error(500, "服务器存储错误"))
		return
	}

	// 生成安全文件名
	fileName := generateSafeFilename(file.Filename)

	// 保存文件（使用正确路径拼接）
	dst := filepath.Join(uploadDir, fileName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		log.Printf("文件保存失败: %v", err)
		c.JSON(http.StatusOK, models.Error(500, "文件保存失败"))
		return
	}

	// 返回可访问URL
	fileURL := fmt.Sprintf("%s%s", publicURL, fileName)
	c.JSON(http.StatusOK, models.Success(fileURL))
}

// 处理上传错误细节
func handleUploadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, http.ErrMissingFile):
		c.JSON(http.StatusOK, models.Error(400, "请选择上传文件"))
	case errors.Is(err, http.ErrContentLength):
		c.JSON(http.StatusOK, models.Error(413, "文件大小超过限制"))
	default:
		log.Printf("上传错误: %v", err)
		c.JSON(http.StatusOK, models.Error(500, "文件上传失败"))
	}
}

// 生成安全文件名
func generateSafeFilename(original string) string {
	ext := filepath.Ext(original) // 自动处理多扩展名情况
	return uuid.New().String() + ext
}

// 文件类型白名单校验
func isSafeFileType(filename string) bool {
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".pdf":  true,
		".docx": true,
		".doc":  true,
		".xlsx": true,
		".xls":  true,
		".pptx": true,
		".ppt":  true,
		".txt":  true,
		".zip":  true,
		".rar":  true,
		".7z":   true,
	}
	ext := strings.ToLower(filepath.Ext(filename))
	return allowed[ext]
}
