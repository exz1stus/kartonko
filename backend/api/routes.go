package main

import (
	"net/http"
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

	r.GET("/health", rh.GetHealthCheckRequest)

	r.GET("/user/:username", rh.GetUserRequest)

	r.GET("/image/:name", rh.GetImageByNameRequest)
	r.GET("/raw-image/:name", rh.GetRawImageByNameRequest)

	r.GET("/images", rh.GetImagesByQueryRequest)
	r.GET("/log", rh.GetAuditLogEntriesRequest)
	r.GET("/tags", rh.GetTagsRequest)

	r.POST("/auth/login", rh.LoginRequest)
	r.POST("/auth/register", rh.RegisterRequest)
	r.GET("/auth/logout", rh.LogoutRequest)

	r.GET("/auth/google", rh.GoogleLoginRequest)
	r.GET("/auth/google/callback", rh.GoogleCallbackRequest)

	authGroup := r.Group("/")
	authGroup.Use(app.AuthMiddleware())
	{
		authGroup.GET("/me", rh.GetMeRequest)
		authGroup.POST("/upload", rh.PostImageRequest)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})
}
