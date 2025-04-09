package models

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

func Paginate(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	result := Pagination{}
	if err := c.ShouldBindQuery(&result); err != nil {
		result.Page = 1
		result.PageSize = 10
	}
	offset := (result.Page - 1) * result.PageSize
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(result.PageSize)
	}
}

type PaginationResponse struct {
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"totalPages"`
	Data       interface{} `json:"data"`
}
