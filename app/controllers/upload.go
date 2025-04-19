package controllers

import (
	"gin_back/app/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"os"
	"path"
)

func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, models.Error(http.StatusBadRequest, "文件为空，上传失败"))
		return
	}
	// 创建上传目录
	uploadDir := "./upload/"
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		c.JSON(http.StatusOK, models.Error(http.StatusInternalServerError, "创建上传目录失败"))
		return
	}

	// 生成新的文件名
	ext := path.Ext(file.Filename)
	newFileName := uuid.New().String() + ext

	// 保存文件
	dst := uploadDir + newFileName
	err = c.SaveUploadedFile(file, dst)
	if err != nil {
		c.JSON(http.StatusOK, models.Error(500, "上传失败"))
		return
	}

	c.JSON(http.StatusOK, models.Success(dst))
}
