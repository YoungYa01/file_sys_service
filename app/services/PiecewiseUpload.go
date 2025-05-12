package services

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	uploadDir = "/uploads"
)

// uploadHandler 处理文件分片上传
func uploadHandler(c *gin.Context) {
	chunkIndex := c.Query("index")
	fileName := c.Query("filename")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	chunkDir := filepath.Join(uploadDir, fileName+".chunks")
	err = os.MkdirAll(chunkDir, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	chunkPath := filepath.Join(chunkDir, chunkIndex)
	err = c.SaveUploadedFile(file, chunkPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "分片上传成功", "index": chunkIndex})
}

// uploadedHandler 返回已上传的分片
func uploadedHandler(c *gin.Context) {
	fileName := c.Query("filename")
	chunkDir := filepath.Join(uploadDir, fileName+".chunks")

	files, err := os.ReadDir(chunkDir)
	if err != nil {
		c.Data(http.StatusOK, "application/json", []byte("[]"))
		return
	}

	var uploaded []string
	for _, f := range files {
		if !f.IsDir() {
			uploaded = append(uploaded, f.Name())
		}
	}

	response, _ := json.Marshal(uploaded)
	c.Data(http.StatusOK, "application/json", response)
}

// mergeHandler 合并所有分片
func mergeHandler(c *gin.Context) {
	fileName := c.Query("filename")
	totalChunksStr := c.Query("total")
	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分片总数"})
		return
	}

	chunkDir := filepath.Join(uploadDir, fileName+".chunks")
	finalFilePath := filepath.Join(uploadDir, fileName)

	// 删除已存在的目标文件
	if _, err := os.Stat(finalFilePath); err == nil {
		os.Remove(finalFilePath)
	}

	finalFile, err := os.Create(finalFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建合并文件失败: " + err.Error()})
		return
	}
	defer finalFile.Close()

	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(chunkDir, strconv.Itoa(i))
		if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("缺少分片 %d，无法合并", i)})
			return
		}

		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer chunkFile.Close()

		_, err = io.Copy(finalFile, chunkFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		os.Remove(chunkPath)
	}

	os.RemoveAll(chunkDir)
	c.JSON(http.StatusOK, gin.H{"message": "文件合并完成", "path": finalFilePath})
}
