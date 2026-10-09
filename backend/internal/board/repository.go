package board

import (
	"context"
	"fmt"
	"server/internal/image"

	"gorm.io/gorm"
)

type BoardRepository interface {
	WithTx(tx *gorm.DB) BoardRepository

	Create(ctx context.Context, board *Board) error
	Get(ctx context.Context, id uint) (*Board, error)
	Update(ctx context.Context, board *Board) error
	Delete(ctx context.Context, id uint) error

	List(ctx context.Context, cursor, limit int) ([]Board, error)

	ListItems(ctx context.Context, boardID uint, cursor, limit int) ([]BoardItem, error)

	AddItem(ctx context.Context, boardID uint, item *BoardItem) error
	GetItem(ctx context.Context, boardID uint, imageID uint) (*BoardItem, error)
	UpdateItem(ctx context.Context, boardID uint, item *BoardItem) error
	DeleteItem(ctx context.Context, boardID uint, imageID uint) error
}

type boardRepository struct {
	db *gorm.DB
}

func (r *boardRepository) List(ctx context.Context, cursor, limit int) ([]Board, error) {
	var boards []Board

	db := r.db.WithContext(ctx)

	err := image.ApplyCursorLimit(db, cursor, int(limit)).
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

func (r *boardRepository) Create(ctx context.Context, board *Board) error {
	return r.db.WithContext(ctx).Create(board).Error
}

func (r *boardRepository) Update(ctx context.Context, board *Board) error {
	result := r.db.WithContext(ctx).Updates(board)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *boardRepository) Delete(ctx context.Context, id uint) error {
	return r.db.
		WithContext(ctx).
		Where("id = ?", id).
		Delete(&Board{}).
		Error
}

func (r *boardRepository) Get(ctx context.Context, id uint) (*Board, error) {
	var board Board
	err := r.db.WithContext(ctx).Preload("Items").Where("id = ?", id).First(&board).Error
	return &board, err
}

func (r *boardRepository) AddItem(ctx context.Context, boardID uint, item *BoardItem) error {
	item.BoardID = boardID

	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("failed to create board item: %w", err)
	}

	return nil
}

func (r *boardRepository) DeleteItem(ctx context.Context, boardID uint, imageID uint) error {
	return r.db.
		WithContext(ctx).
		Where("board_id = ? AND image_id = ?", boardID, imageID).
		Delete(&BoardItem{}).Error

}

func (r *boardRepository) GetItem(ctx context.Context, boardID uint, imageID uint) (*BoardItem, error) {
	var item BoardItem
	err := r.db.
		WithContext(ctx).
		Where("board_id = ? AND image_id = ?", boardID, imageID).
		Preload("Image").
		First(&item).Error
	return &item, err
}

func (r *boardRepository) ListItems(ctx context.Context, boardID uint, cursor, limit int) ([]BoardItem, error) {
	var items []BoardItem

	db := r.db.WithContext(ctx)

	err := image.ApplyCursorLimit(db, cursor, limit).
		Preload("Image").
		Where("board_id = ?", boardID).
		Order("id DESC").
		Find(&items).Error

	return items, err
}

func (r *boardRepository) UpdateItem(ctx context.Context, boardID uint, item *BoardItem) error {
	result := r.db.
		WithContext(ctx).
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
