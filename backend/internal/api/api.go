package api

import (
	"fmt"
	"server/internal/database"
	"server/internal/env"
	"server/internal/repositories"
	"server/internal/services"
	"server/internal/storage"
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

	imageService services.ImageService
	userService  services.UserService
	tagService   services.TagService
	logService   services.LogService

	storage   storage.Storage
	jwtSecret string
	db        *gorm.DB
}

func MustInitApi() *api {
	db := database.MustInitDB()
	storage := storage.MustInitGarageClient()

	imageRepo := repositories.NewImageRepository(db)
	userRepo := repositories.NewUserRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	logRepo := repositories.NewLogRepository(db)

	logService := services.NewLogService(logRepo)
	userService := services.NewUserService(userRepo, imageRepo)
	imageService := services.NewImageService(db, imageRepo, logService, storage)
	tagService := services.NewTagService(db, tagRepo, logService)

	//TODO: temporary for development, remove later
	userService.SetPrivilege(1, 1)

	api := &api{
		imageService: imageService,
		userService:  userService,
		tagService:   tagService,
		logService:   logService,

		storage:   storage,
		jwtSecret: env.GetEnvString("JWT_SECRET"),
		db:        db,
	}

	api.initRoutes()

	return api
}

func (api *api) Run() {
	api.router.Run(fmt.Sprintf(":%s", env.GetEnvString("BACKEND_PORT")))
}

// CleanTestDB truncates all test tables for test isolation
func (api *api) CleanTestDB() error {
	return api.db.Exec("TRUNCATE TABLE image_tags, image_metadata, tags, users, audit_entries RESTART IDENTITY CASCADE").Error
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
