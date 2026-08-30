package api

import (
	"server/internal/api/transaction"
	embeddings "server/internal/embedding"
	"server/internal/image"
	"server/internal/log"
	"server/internal/storage"
	"server/internal/tag"
	"server/internal/user"

	"gorm.io/gorm"
)

type App struct {
	ImageService      image.ImageService
	UserService       user.UserService
	TagService        tag.TagService
	LogService        log.LogService
	ObjectService     image.ObjectService
	EmbeddingsService embeddings.EmbeddingsService
}

func NewApp(
	db *gorm.DB,
	storage storage.Storage,
	embClient embeddings.EmbeddingsClient,
) *App {
	imageRepo := image.NewImageRepository(db)
	userRepo := user.NewUserRepository(db)
	tagRepo := tag.NewTagRepository(db)
	logRepo := log.NewLogRepository(db)

	runner := transaction.GormRunner{DB: db}

	logService := log.NewLogService(logRepo)
	objectService := image.NewObjectService(
		storage,
		imageRepo,
	)
	embeddingsService := embeddings.NewEmbeddingsService(
		objectService,
		embClient,
	)
	imageService := image.NewImageService(
		imageRepo,
		logService,
		objectService,
		embeddingsService,
		runner,
	)

	userService := user.NewUserService(userRepo)
	//TODO: temporary for dev
	userService.SetPrivilege(1, 1)

	tagService := tag.NewTagService(
		tagRepo,
		logService,
		runner,
	)

	return &App{
		ImageService:      imageService,
		UserService:       userService,
		TagService:        tagService,
		LogService:        logService,
		ObjectService:     objectService,
		EmbeddingsService: embeddingsService,
	}
}
