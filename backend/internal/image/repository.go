package image

import (
	"context"
	stderrors "errors"
	"fmt"
	"server/internal/errors"
	"server/internal/tag"

	"gorm.io/gorm"
)

type ImageRepository interface {
	WithTx(tx *gorm.DB) ImageRepository

	Create(ctx context.Context, image *ImageMetadata) error

	Get(ctx context.Context, id uint) (*ImageMetadata, error)
	GetByName(ctx context.Context, name string) (*ImageMetadata, error)
	GetByHash(ctx context.Context, hash string) (*ImageMetadata, error)

	Update(ctx context.Context, image *ImageMetadata) error

	Delete(ctx context.Context, id uint) error
	DeleteByIDs(ctx context.Context, ids []uint) error

	Search(ctx context.Context, query *Query) ([]ImageMetadata, error)
	Count(ctx context.Context, query *Query) (int64, error)

	ExistsByHash(ctx context.Context, hash string) (bool, error)
	ExistsByName(ctx context.Context, name string) (bool, error)

	AttachTags(ctx context.Context, imageID uint, tags []tag.Tag) error
	ReplaceTags(ctx context.Context, imageID uint, tags []tag.Tag) error
}

type imageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) WithTx(tx *gorm.DB) ImageRepository {
	return NewImageRepository(tx)
}

func (r *imageRepository) Create(ctx context.Context, image *ImageMetadata) error {
	return r.db.WithContext(ctx).Create(image).Error
}

func (r *imageRepository) Update(ctx context.Context, image *ImageMetadata) error {
	err := r.db.WithContext(ctx).Updates(image).Error
	if stderrors.Is(err, gorm.ErrDuplicatedKey) {
		return errors.ErrDuplicateName
	}
	return err
}

func (r *imageRepository) Get(ctx context.Context, id uint) (*ImageMetadata, error) {
	var image ImageMetadata
	err := r.db.WithContext(ctx).Preload("Tags").Where("id = ?", id).First(&image).Error
	return &image, err
}

func (r *imageRepository) GetByName(ctx context.Context, name string) (*ImageMetadata, error) {
	var image ImageMetadata
	err := r.db.WithContext(ctx).Preload("Tags").Where("filename = ?", name).First(&image).Error
	return &image, err
}

func (r *imageRepository) GetByHash(ctx context.Context, hash string) (*ImageMetadata, error) {
	var image ImageMetadata
	err := r.db.WithContext(ctx).Preload("Tags").Where("hash = ?", hash).First(&image).Error
	return &image, err
}

func (r *imageRepository) Search(ctx context.Context, query *Query) ([]ImageMetadata, error) {
	var images []ImageMetadata
	db := r.db.WithContext(ctx).Model(&ImageMetadata{}).Order("image_metadata.id desc")

	db = applyFilters(db, query)

	if err := db.Find(&images).Error; err != nil {
		return nil, err
	}

	// Manually load tags for each image
	for i := range images {
		if err := r.db.Model(&images[i]).Association("Tags").Find(&images[i].Tags); err != nil {
			return nil, err
		}
	}

	return images, nil
}

func (r *imageRepository) Delete(ctx context.Context, id uint) error {
	return r.db.
		WithContext(ctx).
		Where("id = ?", id).
		Delete(&ImageMetadata{}).
		Error
}

func (r *imageRepository) DeleteByIDs(ctx context.Context, ids []uint) error {
	return r.db.
		WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&ImageMetadata{}).
		Error
}

func (r *imageRepository) Count(ctx context.Context, query *Query) (int64, error) {
	db := r.db.WithContext(ctx).Model(&ImageMetadata{})
	if query != nil {
		db = applyFilters(db, query)
	}

	var count int64
	err := db.Count(&count).Error
	return count, err
}

func (r *imageRepository) ExistsByHash(ctx context.Context, hash string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ImageMetadata{}).Where("hash = ?", hash).Count(&count).Error
	return count > 0, err
}

func (r *imageRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ImageMetadata{}).Where("filename = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *imageRepository) AttachTags(ctx context.Context, imageID uint, tags []tag.Tag) error {
	tagNames := tag.TagsToStrings(tags)
	if len(tags) == 0 {
		return nil
	}

	db := r.db.WithContext(ctx)

	var dbTags []tag.Tag
	if err := db.Where("name IN ?", tagNames).Find(&dbTags).Error; err != nil {
		return fmt.Errorf("failed to retrieve tags: %w", err)
	}

	if len(dbTags) != len(tags) {
		return fmt.Errorf(
			"%w: not all tags found: expected %d, got %d for tags %v",
			errors.ErrBadRequest,
			len(tags),
			len(dbTags),
			tagNames,
		)
	}

	image := &ImageMetadata{
		Model: gorm.Model{
			ID: imageID,
		},
	}
	if err := db.Model(image).Association("Tags").Append(&dbTags); err != nil {
		return fmt.Errorf("failed to associate tags with image: %w", err)
	}

	return nil
}

func (r *imageRepository) ReplaceTags(ctx context.Context, imageID uint, tags []tag.Tag) error {
	db := r.db.WithContext(ctx)
	tagNames := tag.TagsToStrings(tags)
	var dbTags []tag.Tag
	if len(tagNames) > 0 {
		if err := db.Where("name IN ?", tagNames).Find(&dbTags).Error; err != nil {
			return fmt.Errorf("failed to retrieve tags: %w", err)
		}
		if len(dbTags) != len(tags) {
			return fmt.Errorf(
				"%w: not all tags found: expected %d, got %d for tags %v",
				errors.ErrBadRequest,
				len(tags),
				len(dbTags),
				tagNames,
			)
		}
	}

	image := &ImageMetadata{Model: gorm.Model{ID: imageID}}
	if err := db.Model(image).Association("Tags").Replace(&dbTags); err != nil {
		return fmt.Errorf("failed to replace tags with image: %w", err)
	}
	return nil
}

func applyFilters(db *gorm.DB, query *Query) *gorm.DB {
	if query.Prefix != "" {
		db = db.Where("filename LIKE ?", query.Prefix+"%")
	}

	if len(query.Tags) > 0 {
		db = db.Distinct().
			Joins("JOIN image_tags ON image_tags.image_metadata_id = image_metadata.id").
			Joins("JOIN tags ON tags.id = image_tags.tag_id").
			Where("tags.name IN ?", query.Tags).
			Group("image_metadata.id").
			Having("COUNT(DISTINCT tags.name) = ?", len(query.Tags))
	}

	if query.User != nil {
		db = db.Where("image_metadata.user_id = ?", query.User.ID)
	}

	db = ApplyCursorLimit(db, query.Cursor, query.Limit)

	return db
}
