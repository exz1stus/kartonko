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

	app.Router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendOrigin},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	rh := app.RequestHandler

	v1 := app.Router.Group("api/v1")
	{
		v1.StaticFile("/", "../frontend/index.html")
		v1.Static("/static", "../frontend")

		v1.GET("/image/:name", rh.GetImageByNameRequest)
		v1.GET("/search-tags/:query", rh.GetSearchTagsRequest)

		v1.POST("auth/login", rh.LoginRequest)
		v1.POST("auth/register", rh.RegisterRequest)
		v1.POST("auth/logout", rh.LogoutRequest)
	}

	authGroup := v1.Group("/")
	authGroup.Use(app.AuthMiddleware())
	{
		authGroup.POST("/upload", rh.PostImageRequest)
	}

	app.Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
