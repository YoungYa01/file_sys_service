package routes

import (
	"gin_back/app/controllers"
	"gin_back/app/middleware"
	"github.com/gin-gonic/gin"
)

func SetApiGroupRoutes(router *gin.Engine) *gin.Engine {

	publicGroup := router.Group("/api")

	publicGroup.POST("/login", controllers.Login)
	publicGroup.GET("/login-test", controllers.LoginTest)
	publicGroup.GET("/register", controllers.Register)
	publicGroup.GET("/carousel", controllers.CarouselList)

	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.Auth())

	apiGroup.POST("/upload", controllers.Upload)
	// 轮播图
	apiGroup.POST("/carousel", controllers.CreateCarousel)
	apiGroup.PUT("/carousel/:id", controllers.UpdateCarousel)
	apiGroup.DELETE("/carousel/:id", controllers.DeleteCarousel)
	// 用户
	apiGroup.GET("/users", controllers.UserList)
	apiGroup.POST("/users", controllers.CreateUser)
	apiGroup.PUT("/users/:id", controllers.UpdateUser)
	apiGroup.DELETE("/users/:id", controllers.DeleteUser)
	// 角色
	apiGroup.GET("/roles", controllers.RoleList)
	apiGroup.POST("/roles", controllers.CreateRole)
	apiGroup.PUT("/roles/:id", controllers.UpdateRole)
	apiGroup.DELETE("/roles/:id", controllers.DeleteRole)
	// 通知
	apiGroup.GET("/notification", controllers.NotificationList)
	apiGroup.POST("/notification", controllers.NotificationCreate)
	apiGroup.PUT("/notification/:id", controllers.NotificationUpdate)
	apiGroup.DELETE("/notification/:id", controllers.NotificationDelete)
	// 收集任务
	apiGroup.GET("/collection", controllers.CollectionList)
	apiGroup.POST("/collection", controllers.CollectionCreate)
	apiGroup.PUT("/collection/:id", controllers.CollectionUpdate)
	apiGroup.DELETE("/collection/:id", controllers.CollectionDelete)
	return router
}
