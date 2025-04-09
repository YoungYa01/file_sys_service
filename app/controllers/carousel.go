package controllers

import (
	"gin_back/app/services"
	"github.com/gin-gonic/gin"
)

func CarouselList(c *gin.Context) {
	carouselListService, err := services.CarouselService.CarouselListService(c)
	if err != nil {
		return
	}
	c.JSON(200, carouselListService)
}

func CreateCarousel(c *gin.Context) {
	carouselCreateService, err := services.CarouselService.CarouselCreateService(c)
	if err != nil {
		return
	}
	c.JSON(200, carouselCreateService)
}

func UpdateCarousel(c *gin.Context) {
	carouselUpdateService, err := services.CarouselService.CarouselUpdateService(c)
	if err != nil {
		return
	}
	c.JSON(200, carouselUpdateService)
}

func DeleteCarousel(c *gin.Context) {
	carouselDeleteService, err := services.CarouselService.CarouselDeleteService(c)
	if err != nil {
		return
	}
	c.JSON(200, carouselDeleteService)
}
