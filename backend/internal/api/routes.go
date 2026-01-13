package api

import (
	"net/http"
	"server/internal/env"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (api *api) initRoutes() {
	api.router = gin.Default()
	r := api.router

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{env.GetEnvString("FRONTEND_ORIGIN")},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.GET("/health", api.GetHealthCheckRequest)

	r.GET("/user/:username", api.GetUserRequest)

	r.GET("/image/:name", api.GetImageByNameRequest)
	r.GET("/image/raw/:name", api.GetRawImageByNameRequest)
	r.Static("/image/thumb", env.GetEnvString("THUMBNAILS_PATH"))

	r.GET("/images", api.GetImagesByQueryRequest)
	r.GET("/log", api.GetAuditLogEntriesRequest)
	r.GET("/tags", api.GetTagsRequest)

	r.POST("/auth/login", api.LoginRequest)
	r.POST("/auth/register", api.RegisterRequest)
	r.GET("/auth/logout", api.LogoutRequest)

	r.GET("/auth/google", api.GoogleLoginRequest)
	r.GET("/auth/google/callback", api.GoogleCallbackRequest)

	authGroup := r.Group("/")
	authGroup.Use(api.AuthMiddleware())
	{
		authGroup.GET("/me", api.GetMeRequest)
		authGroup.POST("/upload", api.PostImageRequest)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})
}
