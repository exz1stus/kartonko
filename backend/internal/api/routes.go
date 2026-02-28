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

	r.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodHead {
			c.Request.Method = http.MethodGet
		}
		c.Next()
	})

	r.GET("/health", api.GetHealthCheck)

	r.GET("/user/:name", api.GetUserByName)
	r.GET("/user/id/:id", api.GetUserByID)

	r.GET("/image/:name", api.GetImageByName)
	r.GET("/image/id/:id", api.GetImageByID)
	r.GET("/image/raw/:name", api.GetRawImageByName)
	r.Static("/image/thumb", env.GetEnvString("THUMBNAILS_PATH"))

	r.GET("/images", api.GetImagesByQuery)
	r.GET("/log", api.GetAuditLogEntries)
	r.GET("/tags", api.GetTags)

	r.POST("/auth/login", api.PostLogin)
	r.POST("/auth/register", api.PostRegister)
	r.POST("/auth/logout", api.PostLogout)

	r.GET("/auth/google", api.GetGoogleLogin)
	r.GET("/auth/google/callback", api.GetGoogleCallback)

	authGroup := r.Group("/")
	authGroup.Use(api.AuthMiddleware())
	{
		authGroup.GET("/me", api.GetMe)
		authGroup.POST("/upload", api.PostImage)
		authGroup.DELETE("/image/:name", api.DeleteImageByName)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})
}
