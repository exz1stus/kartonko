package board

import (
	"context"
	stderrors "errors"
	"fmt"
	apperrors "server/internal/errors"
	"server/internal/image"
	"strings"

	"gorm.io/gorm"
)

type BoardRepository interface {
	WithTx(tx *gorm.DB) BoardRepository

	Create(ctx context.Context, board *Board) error
	Get(ctx context.Context, id uint) (*Board, error)
	GetBySlug(ctx context.Context, slug string) (*Board, error)
	SlugExists(ctx context.Context, slug string, exceptID uint) (bool, error)
	Update(ctx context.Context, board *Board) error
	Delete(ctx context.Context, id uint) error

	List(ctx context.Context, name string, userID uint, cursor, limit int) ([]Board, error)

	ListItems(ctx context.Context, boardID uint, prefix string, tags []string, cursor, limit int) ([]BoardItem, error)
	ImageIDs(ctx context.Context, boardID uint) ([]uint, error)
	BoardIDsForImage(ctx context.Context, imageID uint) ([]uint, error)

	AddItem(ctx context.Context, boardID uint, item *BoardItem) error
	GetItem(ctx context.Context, boardID uint, imageID uint) (*BoardItem, error)
	UpdateItem(ctx context.Context, boardID uint, item *BoardItem) error
	DeleteItem(ctx context.Context, boardID uint, imageID uint) error
}

type boardRepository struct {
	db *gorm.DB
}

func (r *boardRepository) List(ctx context.Context, name string, userID uint, cursor, limit int) ([]Board, error) {
	var boards []Board

	db := r.db.WithContext(ctx)
	query := db.Model(&Board{}).
		Joins(`LEFT JOIN (
			SELECT board_items.board_id, COUNT(DISTINCT board_items.image_id) AS image_count
			FROM board_items
			JOIN image_metadata ON image_metadata.id = board_items.image_id
				AND image_metadata.deleted_at IS NULL
			GROUP BY board_items.board_id
		) AS image_counts ON image_counts.board_id = boards.id`).
		Select("boards.*, COALESCE(image_counts.image_count, 0) AS image_count")
	if name != "" {
		query = query.Where("LOWER(boards.name) LIKE ?", "%"+strings.ToLower(name)+"%")
	}
	if userID != 0 {
		query = query.Where("boards.user_id = ?", userID)
	}
	err := image.ApplyCursorLimit(query, cursor, limit).
		Order("image_count DESC, boards.name ASC, boards.id ASC").
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
	db := r.db.WithContext(ctx)
	if err := db.Where("id = ?", id).First(&board).Error; err != nil {
		return &board, err
	}
	if err := db.Model(&BoardItem{}).
		Joins("JOIN image_metadata ON image_metadata.id = board_items.image_id AND image_metadata.deleted_at IS NULL").
		Where("board_items.board_id = ?", id).Count(&board.ImageCount).Error; err != nil {
		return &board, err
	}
	return &board, nil
}

func (r *boardRepository) GetBySlug(ctx context.Context, slug string) (*Board, error) {
	var board Board
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&board).Error; err != nil {
		return nil, err
	}
	return r.Get(ctx, board.ID)
}

func (r *boardRepository) SlugExists(ctx context.Context, slug string, exceptID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&Board{}).
		Where("slug = ? AND id <> ?", slug, exceptID).Count(&count).Error
	return count > 0, err
}

func (r *boardRepository) AddItem(ctx context.Context, boardID uint, item *BoardItem) error {
	item.BoardID = boardID

	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return apperrors.ErrAlreadyExists
		}
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
		Joins("JOIN image_metadata ON image_metadata.id = board_items.image_id AND image_metadata.deleted_at IS NULL").
		Where("board_items.board_id = ? AND board_items.image_id = ?", boardID, imageID).
		Preload("Image").
		Preload("Image.Tags").
		First(&item).Error
	return &item, err
}

func (r *boardRepository) ListItems(ctx context.Context, boardID uint, prefix string, tags []string, cursor, limit int) ([]BoardItem, error) {
	var items []BoardItem

	db := r.db.WithContext(ctx).Model(&BoardItem{}).
		Joins("JOIN image_metadata ON image_metadata.id = board_items.image_id AND image_metadata.deleted_at IS NULL").
		Where("board_items.board_id = ?", boardID)
	if prefix != "" {
		db = db.Where("image_metadata.filename LIKE ?", prefix+"%")
	}
	if len(tags) > 0 {
		matchingImages := r.db.Table("image_tags").
			Select("image_tags.image_metadata_id").
			Joins("JOIN tags ON tags.id = image_tags.tag_id AND tags.deleted_at IS NULL").
			Where("tags.name IN ?", tags).
			Group("image_tags.image_metadata_id").
			Having("COUNT(DISTINCT tags.name) = ?", len(tags))
		db = db.Where("board_items.image_id IN (?)", matchingImages)
	}

	err := image.ApplyCursorLimit(db, cursor, limit).
		Preload("Image").
		Preload("Image.Tags").
		Order("board_items.id DESC").
		Find(&items).Error

	return items, err
}

func (r *boardRepository) ImageIDs(ctx context.Context, boardID uint) ([]uint, error) {
	ids := make([]uint, 0)
	err := r.db.WithContext(ctx).Model(&BoardItem{}).
		Joins("JOIN image_metadata ON image_metadata.id = board_items.image_id AND image_metadata.deleted_at IS NULL").
		Where("board_items.board_id = ?", boardID).
		Order("board_items.image_id").
		Pluck("board_items.image_id", &ids).Error
	return ids, err
}

func (r *boardRepository) BoardIDsForImage(ctx context.Context, imageID uint) ([]uint, error) {
	ids := make([]uint, 0)
	err := r.db.WithContext(ctx).Model(&BoardItem{}).
		Joins("JOIN boards ON boards.id = board_items.board_id AND boards.deleted_at IS NULL").
		Where("board_items.image_id = ?", imageID).
		Order("board_items.board_id").
		Pluck("board_items.board_id", &ids).Error
	return ids, err
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
