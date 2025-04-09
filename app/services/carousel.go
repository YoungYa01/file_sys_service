package services

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type carouselService struct {
}

func SearchByTitle(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		searchKey := c.Query("title") // 从URL参数获取title参数
		if searchKey != "" {
			// 使用LIKE进行模糊查询，并防止SQL注入
			return db.Where("title LIKE ?", "%"+searchKey+"%")
		}
		return db
	}
}

func (*carouselService) CarouselListService(c *gin.Context) (models.Result, error) {
	var carousels []models.Carousel
	var total int64

	// 创建基础查询
	baseQuery := config.DB.Model(&models.Carousel{}).Order("`sort` ASC")

	// 执行分页查询（包含数据查询和总数统计）
	if err := baseQuery.
		Scopes(
			SearchByTitle(c),
			Paginate(c)).
		Find(&carousels).
		Error; err != nil {
		return models.Fail(500, "查询失败"), err
	}

	// 获取总数（排除分页条件）
	if err := baseQuery.Count(&total).Error; err != nil {
		return models.Fail(500, "获取总数失败"), err
	}

	// 构建分页响应
	pagination := models.PaginationResponse{
		Page:     c.GetInt("page"),
		PageSize: c.GetInt("pageSize"),
		Total:    total,
		Data:     carousels,
	}

	return models.Success(pagination), nil
}

func (s *carouselService) CarouselCreateService(c *gin.Context) (models.Result, error) {
	var carousel models.Carousel
	if err := c.ShouldBindJSON(&carousel); err != nil {
		return models.Fail(400, "参数错误"), err
	}
	if carousel.Title == "" || carousel.URL == "" {
		return models.Fail(400, "参数错误"), nil
	}
	if err := config.DB.Create(&carousel).Error; err != nil {
		return models.Fail(400, "创建失败"), err
	}
	return models.Success(models.SuccessWithMsg("创建成功")), nil
}

func (s *carouselService) CarouselUpdateService(c *gin.Context) (models.Result, error) {
	var carousel models.Carousel
	id := c.Param("id")
	if err := config.DB.Where("id = ?", id).First(&carousel).Error; err != nil {
		return models.Fail(400, "请检查id是否正确"), err
	}

	if err := c.ShouldBindJSON(&carousel); err != nil {
		return models.Fail(400, "参数错误"), err
	}

	if carousel.ID == 0 || carousel.Title == "" || carousel.URL == "" {
		return models.Fail(400, "参数错误"), nil
	}
	if err := config.DB.Where("id = ?", carousel.ID).Updates(&carousel).Error; err != nil {
		return models.Fail(400, "更新失败"), err
	}
	return models.Success(models.SuccessWithMsg("更新成功")), nil
}

func (s *carouselService) CarouselDeleteService(c *gin.Context) (models.Result, error) {
	id := c.Param("id")
	if err := config.DB.Where("id = ?", id).Delete(&models.Carousel{}).Error; err != nil {
		return models.Fail(400, "删除失败"), err
	}
	return models.Success(models.SuccessWithMsg("删除成功")), nil
}

// 增强的Paginate函数
func Paginate(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 绑定分页参数
		var params models.Pagination
		if err := c.ShouldBindQuery(&params); err != nil {
			params.Page = 1
			params.PageSize = 10
		}

		// 参数有效性校验
		if params.Page < 1 {
			params.Page = 1
		}
		if params.PageSize <= 0 || params.PageSize > 100 {
			params.PageSize = 10
		}

		// 存储参数到context
		c.Set("page", params.Page)
		c.Set("pageSize", params.PageSize)

		offset := (params.Page - 1) * params.PageSize
		return db.Offset(offset).Limit(params.PageSize)
	}
}

var CarouselService = new(carouselService)
