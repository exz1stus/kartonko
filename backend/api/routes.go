package main

import (
	"server/internal/env"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (app *Application) initRoutes() {
	app.Router = gin.Default()
	r := app.Router

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{env.GetEnvString("FRONTEND_ORIGIN")},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	rh := app.RequestHandler

	r.GET("/user/:username", rh.GetUserRequest)

	r.GET("/image/:name", rh.GetImageByNameRequest)
	r.GET("/raw-image/:name", rh.GetRawImageByNameRequest)
	r.GET("/images", rh.GetImagesRequest)
	r.GET("/search-images", rh.GetImageByQueryRequest)

	r.GET("/search-tags/:query", rh.GetSearchTagsRequest)
	r.GET("/autocomplete-tag", rh.GetTagAutoCompleteRequest)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("auth/login", rh.LoginRequest)
	r.POST("auth/register", rh.RegisterRequest)
	r.GET("auth/logout", rh.LogoutRequest)

	r.GET("/auth/google", rh.GoogleLoginRequest)
	r.GET("/auth/google/callback", rh.GoogleCallbackRequest)

	authGroup := r.Group("/")
	authGroup.Use(app.AuthMiddleware())
	{
		authGroup.GET("/me", rh.GetProfileRequest)
		authGroup.POST("/upload", rh.PostImageRequest)
	}
}
