package services

import (
	"gin_back/app/models"
	"gin_back/config"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func MyTaskListService(c *gin.Context) (models.Result, error) {
	claims, _ := c.Get("claims")
	userId := claims.(*models.CustomClaims).UserID

	page := c.GetInt("page")
	pageSize := c.GetInt("pageSize")

	var myTaskList []models.CollectionSubmitter
	if err := config.DB.Model(&models.CollectionSubmitter{}).
		Where("user_id = ?", userId).
		Find(&myTaskList).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
	}
	// 根据CollectionSubmitter的collection_id去重，然后再查询
	var collectionIds []uint
	for _, v := range myTaskList {
		collectionIds = append(collectionIds, v.CollectionID)
	}
	log.Printf("collectionIds: %v", collectionIds)

	var baseCollections []models.CollectionCreator
	var total int64
	if err := config.DB.Model(&models.CollectionCreator{}).
		Where("id IN ?", collectionIds).
		Count(&total).
		Order("created_at DESC").
		Scopes(
			Paginate(c),
		).
		Find(&baseCollections).Error; err != nil {
		return models.Fail(http.StatusInternalServerError, "查询失败"), err
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
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Data:     result,
	}), nil
}
