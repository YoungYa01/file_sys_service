package main

import (
	"gin_back/app/middleware"
	"gin_back/config"
	"gin_back/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"time"
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

	application.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许的前端地址
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "*"
		},
		MaxAge: 12 * time.Hour,
	}))

	application.Static("/static", "./static")
	application.Static("/upload", "./upload")

	application = routes.SetApiGroupRoutes(application)

	application.Run(":8080")
}
