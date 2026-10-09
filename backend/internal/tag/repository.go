package tag

import (
	"context"
	stderrors "errors"
	"fmt"
	apperrors "server/internal/errors"

	"gorm.io/gorm"
)

type TagRepository interface {
	WithTx(tx *gorm.DB) TagRepository

	Create(ctx context.Context, tag *Tag) error
	Get(ctx context.Context, tagID uint) (*Tag, error)
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, tagID uint) error

	SearchPrefix(ctx context.Context, prefix string, cursor int, limit int) ([]Tag, error)

	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistMany(ctx context.Context, tags []Tag) ([]bool, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) WithTx(tx *gorm.DB) TagRepository {
	return NewTagRepository(tx)
}

func (r *tagRepository) Get(ctx context.Context, tagID uint) (*Tag, error) {
	var tag Tag
	err := r.db.WithContext(ctx).Where("id = ?", tagID).First(&tag).Error
	return &tag, err
}

func (r *tagRepository) Update(ctx context.Context, tag *Tag) error {
	result := r.db.WithContext(ctx).Updates(tag)
	if result.Error != nil {
		if stderrors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return apperrors.ErrDuplicateName
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *tagRepository) Delete(ctx context.Context, tagID uint) error {
	return r.db.
		WithContext(ctx).
		Where("id = ?", tagID).
		Delete(&Tag{}).
		Error
}

func (r *tagRepository) Create(ctx context.Context, tag *Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *tagRepository) SearchPrefix(ctx context.Context, prefix string, cursor int, limit int) ([]Tag, error) {
	var tags []Tag
	if err := r.db.WithContext(ctx).Model(&Tag{}).
		Where("name LIKE ?", prefix+"%").
		Offset(cursor).
		Limit(limit).
		Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *tagRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Tag{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *tagRepository) ExistMany(ctx context.Context, tags []Tag) ([]bool, error) {
	tagNames := make([]string, len(tags))

	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	var retrievedTags []Tag
	if err := r.db.
		WithContext(ctx).
		Where("name IN ?", tagNames).
		Find(&retrievedTags).
		Error; err != nil {
		return nil, fmt.Errorf("failed to check tags in database: %w", err)
	}

	exists := make(map[string]bool, len(retrievedTags))
	for _, tag := range retrievedTags {
		exists[tag.Name] = true
	}

	result := make([]bool, len(tags))
	for i, tag := range tags {
		result[i] = exists[tag.Name]
	}

	return result, nil
}
