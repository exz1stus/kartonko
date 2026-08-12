package api

import (
	"fmt"
	"server/internal/database"
	embeddings "server/internal/embedding"
	"server/internal/env"
	"server/internal/image"
	"server/internal/log"
	"server/internal/storage"
	"server/internal/tag"

	authapi "server/internal/api/auth"
	imageapi "server/internal/api/image"
	logapi "server/internal/api/log"
	tagapi "server/internal/api/tag"
	userapi "server/internal/api/user"

	"server/internal/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
}

type apiBuilder struct {
	storage   storage.Storage
	db        *gorm.DB
	embClient embeddings.EmbeddingsClient
	router    *gin.Engine

	imageRepo image.ImageRepository
	tagRepo   tag.TagRepository
	userRepo  user.UserRepository
	logRepo   log.LogRepository

	imageService      image.ImageService
	tagService        tag.TagService
	UserService       user.UserService
	logService        log.LogService
	objectService     image.ObjectService
	embeddingsService embeddings.EmbeddingsService

	imageHandler *imageapi.Handler
	authHandler  *authapi.Handler
	tagHandler   *tagapi.Handler
	userHandler  *userapi.Handler
	logHandler   *logapi.Handler
}

func NewApiBuilder(
	db *gorm.DB,
	storage storage.Storage,
	embClient embeddings.EmbeddingsClient,
) *apiBuilder {
	return &apiBuilder{
		db:        db,
		storage:   storage,
		embClient: embClient,
	}
}

func (b *apiBuilder) InitRepos() *apiBuilder {
	b.imageRepo = image.NewImageRepository(b.db)
	b.userRepo = user.NewUserRepository(b.db)
	b.tagRepo = tag.NewTagRepository(b.db)
	b.logRepo = log.NewLogRepository(b.db)

	return b
}

func (b *apiBuilder) InitServices() *apiBuilder {
	b.logService = log.NewLogService(b.logRepo)
	b.objectService = image.NewObjectService(
		b.storage,
		b.imageRepo,
	)
	b.embeddingsService = embeddings.NewEmbeddingsService(
		b.objectService,
		b.embClient,
	)
	b.imageService = image.NewImageService(
		b.db,
		b.imageRepo,
		b.logService,
		b.objectService,
		b.embeddingsService,
	)
	b.UserService = user.NewUserService(b.userRepo)
	b.tagService = tag.NewTagService(
		b.db,
		b.tagRepo,
		b.logService,
	)

	return b
}

func (b *apiBuilder) InitHandlers(jwtSecret string) *apiBuilder {
	b.authHandler = authapi.NewAuthHandler(jwtSecret, b.UserService)
	b.userHandler = userapi.NewUserHandler(b.UserService)
	b.logHandler = logapi.NewLogHandler(b.logService)
	b.imageHandler = imageapi.NewImageHandler(
		b.imageService,
		b.objectService,
		b.UserService,
		b.tagService,
	)
	b.tagHandler = tagapi.NewTagHandler(b.tagService)

	return b
}

func (b *apiBuilder) Build() *Api {
	return &Api{
		nil,

		b.imageHandler,
		b.authHandler,
		b.tagHandler,
		b.userHandler,
		b.logHandler,
	}
}

// ImageService returns the image service assembled by the builder.
// It is primarily useful when constructing black-box HTTP test fixtures.
func (b *apiBuilder) ImageService() image.ImageService { return b.imageService }

// TagService returns the tag service assembled by the builder.
// It is primarily useful when constructing black-box HTTP test fixtures.
func (b *apiBuilder) TagService() tag.TagService { return b.tagService }

func MustInitApi() *Api {
	db := database.MustInitDB()
	storage := storage.MustInitGarageClient()
	embeddingsClient := embeddings.NewHTTPEmbeddingClient()
	jwtSecret := env.GetEnvString("JWT_SECRET")

	builder := NewApiBuilder(db, storage, embeddingsClient).
		InitRepos().
		InitServices().
		InitHandlers(jwtSecret)

	//TODO: temporary for development, remove later
	builder.UserService.SetPrivilege(1, 1)

	api := builder.Build()
	api.InitRoutes(api.authHandler.AuthMiddleware)

	return api
}

func (api *Api) Run() {
	api.router.Run(fmt.Sprintf(":%s", env.GetEnvString("BACKEND_PORT")))
}

func (api *Api) Router() *gin.Engine {
	return api.router
}
