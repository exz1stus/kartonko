package image

import (
	"fmt"
	"server/internal/errors"
	"server/internal/tag"

	"gorm.io/gorm"
)

type ImageRepository interface {
	WithTx(tx *gorm.DB) ImageRepository

	Create(image *ImageMetadata) error

	Update(image *ImageMetadata) error

	GetByID(id uint) (*ImageMetadata, error)
	GetByName(name string) (*ImageMetadata, error)
	GetByHash(hash string) (*ImageMetadata, error)
	Search(query *Query) ([]ImageMetadata, error)

	DeleteByID(id uint) error
	DeleteByIDs(ids []uint) error

	Count(query *Query) (int64, error)

	ExistsByHash(hash string) (bool, error)
	ExistsByName(name string) (bool, error)

	AttachTags(imageID uint, tags []tag.Tag) error
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

func (r *imageRepository) Create(image *ImageMetadata) error {
	return r.db.Create(image).Error
}

func (r *imageRepository) Update(image *ImageMetadata) error {
	return r.db.Updates(image).Error
}

func (r *imageRepository) GetByID(id uint) (*ImageMetadata, error) {
	var image ImageMetadata
	err := r.db.Preload("Tags").Where("id = ?", id).First(&image).Error
	return &image, err
}

func (r *imageRepository) GetByName(name string) (*ImageMetadata, error) {
	var image ImageMetadata
	err := r.db.Preload("Tags").Where("filename = ?", name).First(&image).Error
	return &image, err
}

func (r *imageRepository) GetByHash(hash string) (*ImageMetadata, error) {
	var image ImageMetadata
	err := r.db.Preload("Tags").Where("hash = ?", hash).First(&image).Error
	return &image, err
}

func (r *imageRepository) Search(query *Query) ([]ImageMetadata, error) {
	var images []ImageMetadata
	db := r.db.Model(&ImageMetadata{}).Order("image_metadata.id desc")

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

func (r *imageRepository) DeleteByID(id uint) error {
	return r.db.
		Where("id = ?", id).
		Delete(&ImageMetadata{}).
		Error
}

func (r *imageRepository) DeleteByIDs(ids []uint) error {
	return r.db.
		Where("id IN ?", ids).
		Delete(&ImageMetadata{}).
		Error
}

func (r *imageRepository) Count(query *Query) (int64, error) {
	db := r.db.Model(&ImageMetadata{})
	if query != nil {
		db = applyFilters(db, query)
	}

	var count int64
	err := db.Count(&count).Error
	return count, err
}

func (r *imageRepository) ExistsByHash(hash string) (bool, error) {
	var count int64
	err := r.db.Model(&ImageMetadata{}).Where("hash = ?", hash).Count(&count).Error
	return count > 0, err
}

func (r *imageRepository) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&ImageMetadata{}).Where("filename = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *imageRepository) AttachTags(imageID uint, tags []tag.Tag) error {
	tagNames := tag.TagsToStrings(tags)
	if len(tags) == 0 {
		return nil
	}
	var dbTags []tag.Tag
	if err := r.db.Where("name IN ?", tagNames).Find(&dbTags).Error; err != nil {
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
	if err := r.db.Model(image).Association("Tags").Append(&dbTags); err != nil {
		return fmt.Errorf("failed to associate tags with image: %w", err)
	}

	return nil
}

func applyFilters(db *gorm.DB, query *Query) *gorm.DB {
	if query.Prefix != "" {
		db = db.Where("filename LIKE ?", query.Prefix+"%")
	}

	if len(query.Tags) > 0 {
		// Use HAVING COUNT to ensure ALL tags are present (AND logic)
		db = db.Distinct().
			Joins("JOIN image_tags ON image_tags.image_metadata_id = image_metadata.id").
			Joins("JOIN tags ON tags.id = image_tags.tag_id").
			Where("tags.name IN ?", query.Tags).
			Group("image_metadata.id").
			Having("COUNT(DISTINCT tags.name) = ?", len(query.Tags))
	}

	if query.User != nil {
		db = db.Where("user_id = ?", query.User.ID)
	}

	if query.Limit != 0 {
		db = db.Limit(query.Limit)
	}

	if query.Cursor != 0 {
		db = db.Offset(query.Cursor)
	}

	db = db.Debug()

	return db
}
