package api

import (
	"net/http"
	"server/internal/env"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/watchakorn-18k/scalar-go"
	scalargin "github.com/watchakorn-18k/scalar-go/middleware/gin"
)

type Handler interface {
	RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup)
}

func (api *Api) InitRoutes(authMiddleware gin.HandlerFunc) {
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

	r.Use(HeadBypassMiddleware)

	protected := r.Group("/", authMiddleware)

	registerRoutes := func(location string, h Handler) {
		public := r.Group(location)
		protected := protected.Group(location)
		h.RegisterRoutes(public, protected)
	}

	registerRoutes("/auth", api.authHandler)
	registerRoutes("/image", api.imageHandler)
	registerRoutes("/tags", api.tagHandler)
	registerRoutes("/user", api.userHandler)
	registerRoutes("/log", api.logHandler)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/openapi.json", func(c *gin.Context) {
		c.File("./docs/openapi.json")
	})

	r.GET("/docs", scalargin.Handler(&scalar.Options{
		SpecURL: "./docs/openapi.json",
		CustomOptions: scalar.CustomOptions{
			PageTitle: "kartonko API",
		},
		DarkMode: true,
	}))

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})

	r.GET("/health", api.GetHealthCheck)
}
