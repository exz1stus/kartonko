package api

import (
	"fmt"
	"server/internal/database"
	embeddings "server/internal/embedding"
	"server/internal/env"
	"server/internal/storage"

	authapi "server/internal/api/auth"
	boardapi "server/internal/api/board"
	imageapi "server/internal/api/image"
	logapi "server/internal/api/log"
	tagapi "server/internal/api/tag"
	userapi "server/internal/api/user"

	"github.com/gin-gonic/gin"
)

// swagger:model
type Response gin.H

type Api struct {
	router *gin.Engine

	imageHandler *imageapi.Handler
	authHandler  *authapi.Handler
	tagHandler   *tagapi.Handler
	userHandler  *userapi.Handler
	logHandler   *logapi.Handler
	boardHandler *boardapi.Handler
}

func NewAPI(app *App, jwtSecret string) *Api {
	authHandler := authapi.NewAuthHandler(jwtSecret, app.UserService)
	userHandler := userapi.NewUserHandler(app.UserService)
	logHandler := logapi.NewLogHandler(app.LogService)
	imageHandler := imageapi.NewImageHandler(
		app.ImageService,
		app.ObjectService,
		app.UserService,
		app.TagService,
	)
	tagHandler := tagapi.NewTagHandler(app.TagService)
	boardHandler := boardapi.NewBoardHandler(app.BoardService)

	return &Api{
		nil,

		imageHandler,
		authHandler,
		tagHandler,
		userHandler,
		logHandler,
		boardHandler,
	}
}

func MustInitApi() *Api {
	db := database.MustInitPostgreDB()
	storage := storage.MustInitGarageClient()
	embeddingsClient := embeddings.NewHTTPEmbeddingClient()
	jwtSecret := env.GetEnvString("JWT_SECRET")

	app := NewApp(db, storage, embeddingsClient)
	api := NewAPI(app, jwtSecret)
	api.InitRoutes(api.authHandler.AuthMiddleware)

	return api
}

func (api *Api) Run() {
	api.router.Run(fmt.Sprintf(":%s", env.GetEnvString("BACKEND_PORT")))
}

func (api *Api) Router() *gin.Engine {
	return api.router
}
