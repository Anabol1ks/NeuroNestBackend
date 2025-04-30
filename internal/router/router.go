package router

import (
	"NeuroNest/internal/auth"
	"NeuroNest/internal/config"
	"NeuroNest/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RouterConfig() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	r.Static("/avatars", config.UploadsPath)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	profileGroup := r.Group("/profile", auth.AuthMiddleware())
	{
		profileGroup.GET("/get", handlers.GetProfileHandler)
		profileGroup.PUT("/update", handlers.UpdateProfileHandler)
		profileGroup.POST("/upload-avatar", handlers.UploadAvatarHandler)
		profileGroup.DELETE("/delete-avatar", handlers.DeleteAvatarHandler)

	}
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/yandex/login", handlers.YandexLoginHandler)
		authGroup.GET("/yandex/callback", handlers.YandexCallbackHandler)
		authGroup.POST("/register", handlers.RegisterHandler)
		authGroup.POST("/login", handlers.LoginHandler)
		authGroup.POST("/refresh", handlers.RefreshToken)
	}
	return r
}
