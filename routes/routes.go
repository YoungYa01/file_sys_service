package routes

import (
	"gin_back/app/controllers"
	"gin_back/app/middleware"
	"gin_back/config"
	"github.com/gin-gonic/gin"
)

func SetApiGroupRoutes(router *gin.Engine) *gin.Engine {

	publicGroup := router.Group("/api")

	publicGroup.POST("/login", controllers.Login)
	publicGroup.GET("/login-test", controllers.LoginTest)
	publicGroup.GET("/register", controllers.Register)
	publicGroup.GET("/carousel", controllers.CarouselList)

	apiGroup := router.Group("/api")
	apiGroup.Use(middleware.Auth(), middleware.LogMiddleware(config.DB))

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
	apiGroup.POST("/users/upload", controllers.UploadUser)
	// 当前用户信息
	apiGroup.GET("/userinfo", controllers.UserDetail)
	// 角色
	apiGroup.GET("/roles", controllers.RoleList)
	apiGroup.POST("/roles", controllers.CreateRole)
	apiGroup.PUT("/roles/:id", controllers.UpdateRole)
	apiGroup.DELETE("/roles/:id", controllers.DeleteRole)
	// 部门
	apiGroup.GET("/organization", controllers.OrgList)
	apiGroup.GET("/organization/users", controllers.OrgUserList)
	apiGroup.GET("/organization/:parent_id", controllers.OrgListOfChildren)
	apiGroup.POST("/organization", controllers.CreateOrg)
	apiGroup.PUT("/organization/:id", controllers.UpdateOrg)
	apiGroup.DELETE("/organization/:id", controllers.DeleteOrg)
	// 日志
	apiGroup.GET("/log", controllers.LogList)
	apiGroup.DELETE("/log", controllers.DeleteLog)
	// 通知
	apiGroup.GET("/notification", controllers.NotificationList)
	apiGroup.POST("/notification", controllers.NotificationCreate)
	apiGroup.PUT("/notification/:id", controllers.NotificationUpdate)
	apiGroup.DELETE("/notification/:id", controllers.NotificationDelete)
	// 收集任务
	apiGroup.GET("/collection", controllers.CollectionList)
	apiGroup.GET("/collection/task-center", controllers.TCCollectionList)
	apiGroup.POST("/collection/task-center", controllers.TCSubmit)
	apiGroup.GET("/collection/:id", controllers.CollectionDetail)
	apiGroup.GET("/collection/submit/:id", controllers.CollectionSubmitDetail)
	apiGroup.POST("/collection", controllers.CollectionCreate)
	apiGroup.PUT("/collection/:id", controllers.CollectionUpdate)
	apiGroup.DELETE("/collection/:id", controllers.CollectionDelete)
	// 审核中心
	apiGroup.GET("/review", controllers.ReviewList)
	apiGroup.GET("/review/:id", controllers.ReviewDetailList)
	apiGroup.PUT("/review/status", controllers.ReviewStatus)
	apiGroup.GET("/review/export", controllers.ReviewExport)
	// 我的任务中心
	apiGroup.GET("/my-task", controllers.MyTaskList)
	return router
}
