package board

import (
	"fmt"
	"server/internal/image"

	"gorm.io/gorm"
)

type BoardRepository interface {
	WithTx(tx *gorm.DB) BoardRepository

	Create(board *Board) error
	Get(id uint) (*Board, error)
	Update(board *Board) error
	Delete(id uint) error

	List(cursor, limit int) ([]Board, error)

	ListItems(boardID uint, cursor, limit int) ([]BoardItem, error)

	AddItem(boardID uint, item *BoardItem) error
	GetItem(boardID uint, imageID uint) (*BoardItem, error)
	UpdateItem(boardID uint, item *BoardItem) error
	DeleteItem(boardID uint, imageID uint) error
}

type boardRepository struct {
	db *gorm.DB
}

func (r *boardRepository) List(cursor, limit int) ([]Board, error) {
	var boards []Board

	err := image.ApplyCursorLimit(r.db, cursor, int(limit)).
		Order("id DESC").
		Find(&boards).Error

	return boards, err
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

func (r *boardRepository) Update(board *Board) error {
	result := r.db.Updates(board)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *boardRepository) Delete(id uint) error {
	return r.db.
		Where("id = ?", id).
		Delete(&Board{}).
		Error
}

func (r *boardRepository) Get(id uint) (*Board, error) {
	var board Board
	err := r.db.Preload("Items").Where("id = ?", id).First(&board).Error
	return &board, err
}

func (r *boardRepository) AddItem(boardID uint, item *BoardItem) error {
	item.BoardID = boardID

	if err := r.db.Create(item).Error; err != nil {
		return fmt.Errorf("failed to create board item: %w", err)
	}

	return nil
}

func (r *boardRepository) DeleteItem(boardID uint, imageID uint) error {
	return r.db.
		Where("board_id = ? AND image_id = ?", boardID, imageID).
		Delete(&BoardItem{}).Error

}

func (r *boardRepository) GetItem(boardID uint, imageID uint) (*BoardItem, error) {
	var item BoardItem
	err := r.db.
		Where("board_id = ? AND image_id = ?", boardID, imageID).
		Preload("Image").
		First(&item).Error
	return &item, err
}

func (r *boardRepository) ListItems(boardID uint, cursor, limit int) ([]BoardItem, error) {
	var items []BoardItem

	err := image.ApplyCursorLimit(r.db, cursor, limit).
		Preload("Image").
		Where("board_id = ?", boardID).
		Order("id DESC").
		Find(&items).Error

	return items, err
}

func (r *boardRepository) UpdateItem(boardID uint, item *BoardItem) error {
	result := r.db.
		Where("board_id = ? AND image_id = ?", boardID, item.ImageID).
		Updates(item)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
