package api

import (
	"qq_demo/middleware"
	"github.com/gin-gonic/gin"
)

func InitRouter() {
	r := gin.Default()
	r.Use(middleware.CORS())
	UserGroup := r.Group("/user")
	{
		UserGroup.POST("/register",UserRegist)
		UserGroup.POST("/login",UserLogin)
		UserGroup.GET("/user_info",middleware.JWTAuthMiddleware(),GetUserInfo)
	}
	//开启websocket连接
	r.GET("/ws/chat",StartWsChat)
	r.Run(":1226")
}