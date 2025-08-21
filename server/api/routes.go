package main

import (
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
		AllowOrigins:     []string{frontendOrigin},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	rh := app.RequestHandler

	v1 := r.Group("api/v1")
	{
		v1.StaticFile("/", "../frontend/index.html")
		v1.Static("/static", "../frontend")

		v1.GET("/image/:name", rh.GetImageByNameRequest)
		v1.GET("/search-tags/:query", rh.GetSearchTagsRequest)
	}

	authGroup := v1.Group("/")
	authGroup.Use(app.AuthMiddleware())
	{
		authGroup.GET("/profile", rh.GetProfileRequest)
		authGroup.POST("/upload", rh.PostImageRequest)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1.POST("auth/login", rh.LoginRequest)
	v1.POST("auth/register", rh.RegisterRequest)
	v1.POST("auth/logout", rh.LogoutRequest)

	r.GET("/auth/google", rh.GoogleLoginRequest)
	r.GET("/auth/google/callback", rh.GoogleCallbackRequest)
}
