package board

import (
	"gorm.io/gorm"
)

type BoardRepository interface {
	WithTx(tx *gorm.DB) BoardRepository

	Create(board *Board) error
}

type boardRepository struct {
	db *gorm.DB
}

func NewBoardRepository(db *gorm.DB) BoardRepository {
	return &boardRepository{db: db}
}

func (r *boardRepository) WithTx(tx *gorm.DB) BoardRepository {
	return NewBoardRepository(tx)
}

func (r *boardRepository) Create(board *Board) error {
	return r.db.Create(board).Error
}