package repositories

import (
	"server/internal/models"

	"gorm.io/gorm"
)

type BoardRepository interface {
	WithTx(tx *gorm.DB) ImageRepository

	Create(image *models.ImageMetadata) error
}

type boardRepository struct {
	db *gorm.DB
}

func NewBoardRepository(db *gorm.DB) BoardRepository {
	return &imageRepository{db: db}
}

func (r *boardRepository) WithTx(tx *gorm.DB) BoardRepository {
	return NewImageRepository(tx)
}

func (r *boardRepository) Create(board *models.Board) error {
	return r.db.Create(board).Error
}
