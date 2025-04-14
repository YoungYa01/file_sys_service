package main

import (
	"gin_back/app/middleware"
	"gin_back/config"
	"gin_back/routes"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}

	// ✅ 检查数据库连接
	if config.DB == nil {
		log.Fatal("数据库未初始化")
	}

	log.Println("数据库连接成功")

	middleware.InitIPDatabase()

	application := gin.Default()

	application.Use(
		gin.Recovery(), // 官方恢复中间件
	)

	application.Static("/static", "./static")
	application.Static("/upload", "./upload")

	application = routes.SetApiGroupRoutes(application)

	application.Run(":8080")
}
