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

	api.initRoutes()

	return api
}

func (api *api) Run() {
	api.router.Run(fmt.Sprintf(":%s", env.GetEnvString("BACKEND_PORT")))
}

// Router returns the gin router for testing
func (api *api) Router() *gin.Engine {
	return api.router
}

// CleanTestDB truncates all test tables for test isolation
func (api *api) CleanTestDB() error {
	return api.db.Exec("TRUNCATE TABLE image_tags, image_metadata, tags, users, audit_entries RESTART IDENTITY CASCADE").Error
}

// MustInitAPIForTest creates an API instance for testing with a given database and storage
func MustInitAPIForTest(db *gorm.DB, storage storage.Storage) *api {
	return MustInitAPIForTestWithAuth(db, storage, nil, nil)
}

// MustInitAPIForTestWithAuth creates an API instance for testing with a custom auth middleware.
// If authMiddleware is nil, the default AuthMiddleware is used.
// If userService is provided, it will be used instead of creating a new one (and createTestUsers is skipped).
func MustInitAPIForTestWithAuth(db *gorm.DB, storage storage.Storage, authMiddleware func(*gin.RouterGroup), userService user.UserService) *api {
	// Use mock embeddings client for tests to avoid external dependency
	embeddingsClient := embeddingspkg.NewMockEmbeddingsClient()

	imageRepo := image.NewImageRepository(db)
	userRepo := user.NewUserRepository(db)
	tagRepo := tag.NewTagRepository(db)
	logRepo := log.NewLogRepository(db)

	logService := log.NewLogService(logRepo)
	objectService := image.NewObjectService(storage, imageRepo)
	embeddingsService := embeddingspkg.NewEmbeddingsService(objectService, embeddingsClient)

	if userService == nil {
		userService = user.NewUserService(userRepo)
		// Create test users
		createTestUsers(db, userService)
	}

	imageService := image.NewImageService(db, imageRepo, logService, objectService, embeddingsService)
	tagService := tag.NewTagService(db, tagRepo, logService)

	api := &api{
		imageService:      imageService,
		userService:       userService,
		tagService:        tagService,
		logService:        logService,
		objectService:     objectService,
		embeddingsService: embeddingsService,

		storage:   storage,
		jwtSecret: "test-secret",
		db:        db,
	}

	api.initRoutesWithAuth(authMiddleware)

	return api
}

func createTestUsers(db *gorm.DB, userService user.UserService) {
	users := []*user.User{
		{Username: "moderator", Email: "moderator@test.com", Privilege: user.Moderator, Provider: "test", ProviderID: "moderator"},
		{Username: "user1", Email: "user1@test.com", Privilege: user.Unprivileged, Provider: "test", ProviderID: "user1"},
		{Username: "user2", Email: "user2@test.com", Privilege: user.Unprivileged, Provider: "test", ProviderID: "user2"},
		{Username: "alice", Email: "alice@test.com", Privilege: user.Unprivileged, Provider: "test", ProviderID: "alice"},
		{Username: "user3", Email: "user3@test.com", Privilege: user.Unprivileged, Provider: "test", ProviderID: "user3"},
	}
	for _, u := range users {
		// Check if user already exists
		existing, err := userService.GetByUsername(u.Username)
		if err == nil && existing != nil {
			continue
		}
		if err := userService.Create(u); err != nil {
			panic(fmt.Sprintf("failed to create test user %s: %v", u.Username, err))
		}
	}
}

// AuthService returns the auth service (API itself implements auth via methods)
func (a *api) AuthService() *api { return a }

// UserService returns the user service
func (a *api) UserService() user.UserService { return a.userService }

// TagService returns the tag service
func (a *api) TagService() tag.TagService { return a.tagService }

// ImageService returns the image service
func (a *api) ImageService() image.ImageService { return a.imageService }

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
