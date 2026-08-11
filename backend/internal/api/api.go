package api

import (
	"fmt"
	"server/internal/database"
	embeddingspkg "server/internal/embedding"
	"server/internal/env"
	"server/internal/image"
	"server/internal/log"
	"server/internal/storage"
	"server/internal/tag"
	"server/internal/user"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// swagger:model
type ErrorResponse struct {
	Error string `json:"error"`
}

// swagger:model
type Response gin.H

type api struct {
	router *gin.Engine

	imageService      image.ImageService
	objectService     image.ObjectService
	userService       user.UserService
	tagService        tag.TagService
	logService        log.LogService
	embeddingsService embeddingspkg.EmbeddingsService

	storage   storage.Storage
	jwtSecret string
	db        *gorm.DB
}

func MustInitApi() *api {
	db := database.MustInitDB()
	storage := storage.MustInitGarageClient()
	embeddingsClient := embeddingspkg.NewHTTPEmbeddingClient()

	imageRepo := image.NewImageRepository(db)
	userRepo := user.NewUserRepository(db)
	tagRepo := tag.NewTagRepository(db)
	logRepo := log.NewLogRepository(db)

	logService := log.NewLogService(logRepo)
	objectService := image.NewObjectService(storage, imageRepo)
	embeddingsService := embeddingspkg.NewEmbeddingsService(objectService, embeddingsClient)
	userService := user.NewUserService(userRepo)
	imageService := image.NewImageService(db, imageRepo, logService, objectService, embeddingsService)
	tagService := tag.NewTagService(db, tagRepo, logService)

	//TODO: temporary for development, remove later
	userService.SetPrivilege(1, 1)

	api := &api{
		imageService:      imageService,
		userService:       userService,
		tagService:        tagService,
		logService:        logService,
		objectService:     objectService,
		embeddingsService: embeddingsService,

		storage:   storage,
		jwtSecret: env.GetEnvString("JWT_SECRET"),
		db:        db,
	}

	api.initRoutes(api.AuthMiddleware())

	return api
}

func (api *api) Run() {
	api.router.Run(fmt.Sprintf(":%s", env.GetEnvString("BACKEND_PORT")))
}

func (api *api) Router() *gin.Engine {
	return api.router
}

const defaultLimit = 100

func parseCursorLimit(c *gin.Context) (int, int, error) {
	cursorStr := c.Query("cursor")
	limitStr := c.Query("limit")

	var cursor = 0
	var limit = defaultLimit
	var err error

	if cursorStr != "" {
		cursor, err = strconv.Atoi(cursorStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid cursor parameter: %w", err)
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid limit parameter: %w", err)
		}
	}

	return cursor, limit, nil
}
