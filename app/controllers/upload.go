package controllers

import (
	"gin_back/app/models"
	"github.com/gin-gonic/gin"
)

func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(500, models.Error(400, "文件为空，上传失败"))
		return
	}
	err = c.SaveUploadedFile(file, "./upload/"+file.Filename)
	if err != nil {
		c.JSON(500, models.Error(500, "上传失败"))
		return
	}
	c.JSON(200, models.Success("/upload/"+file.Filename))
}
