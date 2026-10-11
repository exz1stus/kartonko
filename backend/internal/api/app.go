package api

import (
	"context"
	"server/internal/api/transaction"
	"server/internal/board"
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
	BoardService      board.BoardService
	Storage           storage.Storage
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
	boardRepo := board.NewBoardRepository(db)

	runner := transaction.GormRunner{DB: db}

	objectService := image.NewObjectService(
		storage,
		imageRepo,
	)
	embeddingsService := embeddings.NewEmbeddingsService(
		objectService,
		embClient,
	)
	logService := log.NewLogService(logRepo, db, objectService, embeddingsService)
	imageService := image.NewImageService(
		imageRepo,
		userRepo,
		logService,
		objectService,
		embeddingsService,
		runner,
	)

	boardService := board.NewBoardService(boardRepo, userRepo, logService, runner)

	userService := user.NewUserService(userRepo)
	//TODO: temporary for dev
	userService.SetPrivilege(context.Background(), 1, 1)

	tagService := tag.NewTagService(
		tagRepo,
		logService,
		userRepo,
		runner,
	)

	return &App{
		ImageService:      imageService,
		UserService:       userService,
		TagService:        tagService,
		LogService:        logService,
		ObjectService:     objectService,
		EmbeddingsService: embeddingsService,
		BoardService:      boardService,
		Storage:           storage,
	}
}
