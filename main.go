package main

import (
	"gin_back/app/middleware"
	"gin_back/config"
	"gin_back/routes"
	"github.com/gin-gonic/gin"
)

func main() {

	config.LoadConfig()

	application := gin.Default()

	application.Use(middleware.Logger())

	application.Static("/static", "./static")
	application.Static("/upload", "./upload")

	application = routes.SetApiGroupRoutes(application)

	//application.GET("/", func(c *gin.Context) {
	//	c.JSON(200, gin.H{
	//		"message": "Hello World!",
	//	})
	//})

	application.Run(":8080")
}
