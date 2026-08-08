package image

import (
	"fmt"
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

	AttachTags(image *ImageMetadata, tags []tag.Tag) error
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

func (r *imageRepository) AttachTags(image *ImageMetadata, tags []tag.Tag) error {
	tagNames := tag.TagsToStrings(tags)
	if len(tags) == 0 {
		return nil
	}
	var dbTags []tag.Tag
	if err := r.db.Where("name IN ?", tagNames).Find(&dbTags).Error; err != nil {
		return fmt.Errorf("failed to retrieve tags: %w", err)
	}

	if len(dbTags) != len(tags) {
		return fmt.Errorf("not all tags found: expected %d, got %d for tags %v", len(tags), len(dbTags), tagNames)
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

//////////////////////////////////

// func (model *ImageRepository) Search(query models.ImageQuery, cursor int, limit int) ([]models.ImageMetadata, error) {
// 	var images []models.ImageMetadata
// 	db := model.Db.Preload("Tags").Model(&models.ImageMetadata{}).Order("image_metadata.id desc")

// 	db = model.applyFilters(db, query)

// 	db = db.Limit(limit).Offset(cursor).Find(&images)

// 	err := db.Error
// 	if err != nil {
// 		return nil, err
// 	}

// 	return images, nil
// }

// func (model *ImageRepository) CreateImage(image *models.ImageMetadata) error {
// 	if model.containsImageHash(tx, image.Hash) {
// 		return fmt.Errorf("image with hash %s already exists", image.Hash)
// 	}

// 	if model.imageNameExists(tx, image.Filename) {
// 		return fmt.Errorf("image with name %s already exists", image.Filename)
// 	}

// 	if err := model.Tags.checkForAllowedTags(tx, image.Tags); err != nil {
// 		return err
// 	}

// 	tagNames := TagsToStrings(image.Tags)
// 	var dbTags []Tag
// 	if err := tx.Where("name IN ?", tagNames).Find(&dbTags).Error; err != nil {
// 		return fmt.Errorf("failed to retrieve tags: %w", err)
// 	}

// 	image.Tags = nil

// 	if err := tx.Create(image).Error; err != nil {
// 		return fmt.Errorf("failed to insert image: %w", err)
// 	}

// 	if err := tx.Model(image).Association("Tags").Append(dbTags); err != nil {
// 		return fmt.Errorf("failed to associate tags with image: %w", err)
// 	}

// 	return nil
// }

// func (model *ImageRepository) GetImageByHash(hash string) (*models.ImageMetadata, error) {
// 	var img models.ImageMetadata
// 	result := model.Db.Preload("Tags").Where("hash = ?", hash).First(&img)

// 	if result.Error != nil {
// 		return nil, fmt.Errorf("failed retrieving image from the db: %w", result.Error)
// 	}

// 	if img.Hash == "" {
// 		return nil, fmt.Errorf("image with hash %s not found", hash)
// 	}

// 	return &img, nil
// }

// func (model *ImageRepository) GetImageByID(id uint64) (*models.ImageMetadata, error) {
// 	var img models.ImageMetadata
// 	result := model.Db.Preload("Tags").Where("id = ?", id).First(&img)

// 	if result.Error != nil {
// 		return nil, fmt.Errorf("failed retrieving image from the db: %w", result.Error)
// 	}

// 	if img.Hash == "" {
// 		return nil, fmt.Errorf("image with id %d not found", id)
// 	}

// 	return &img, nil
// }

// func (model *ImageRepository) GetImageByName(name string) (*models.ImageMetadata, error) {
// 	var img models.ImageMetadata
// 	result := model.Db.Preload("Tags").Where("filename = ?", name).First(&img)

// 	if result.Error != nil {
// 		return nil, fmt.Errorf("failed retrieving image from the db: %w", result.Error)
// 	}

// 	if img.Hash == "" {
// 		return nil, fmt.Errorf("image with name %s not found", name)
// 	}

// 	return &img, nil
// }

// func (model *ImageRepository) GetImageCount() (int64, error) {
// 	var count int64
// 	err := model.Db.Model(&models.ImageMetadata{}).Count(&count).Error

// 	return count, err
// }

// func (model *ImageRepository) GetImages(cursor int, limit int) ([]models.ImageMetadata, error) {
// 	var images []models.ImageMetadata
// 	result := model.Db.Model(&models.ImageMetadata{}).Preload("Tags").Order("id desc").Limit(limit).Offset(cursor).Find(&images)
// 	if result.Error != nil {
// 		return nil, fmt.Errorf("failed to retrieve images: %w", result.Error)
// 	}

// 	return images, nil
// }

// func (model *ImageRepository) DeleteImagesByQuery(db *gorm.DB, query ImageQuery, userID uint) ([]models.ImageMetadata, error) {
// 	var images []models.ImageMetadata
// 	tx := db.Preload("Tags").Model(&models.ImageMetadata{}).Order("image_metadata.id desc")

// 	tx = model.applyFilters(tx, query)

// 	if err := tx.Find(&images).Error; err != nil {
// 		return nil, fmt.Errorf("failed to find images for deletion: %w", err)
// 	}

// 	if len(images) == 0 {
// 		return nil, nil
// 	}

// 	for _, img := range images {
// 		if !model.UserCanEdit(db, &img, userID) {
// 			return nil, fmt.Errorf(
// 				"permission denied: user %d cannot delete image '%s' (ID: %d)",
// 				userID, img.Filename, img.ID,
// 			)
// 		}
// 	}

// 	// Use a clean query for delete to avoid table name conflicts from JOINs in applyFilters
// 	imageIDs := make([]uint, len(images))
// 	for i, img := range images {
// 		imageIDs[i] = img.ID
// 	}

// 	if err := db.Where("id IN ?", imageIDs).Delete(&models.ImageMetadata{}).Error; err != nil {
// 		return nil, fmt.Errorf("failed to delete images: %w", err)
// 	}

// 	return images, nil
// }

// func (model *ImageRepository) UserCanEdit(db *gorm.DB, image *models.ImageMetadata, userID uint) bool {
// 	if image.UserID == userID {
// 		return true
// 	}

// 	var user User
// 	// Use a clean query on the users table, not the transaction with image JOINs
// 	err := db.Unscoped().Model(&User{}).Where("id = ? AND privilege = ?", userID, Moderator).First(&user).Error
// 	return err == nil
// }

// func (model *ImageRepository) DeleteImage(db *gorm.DB, image *models.ImageMetadata, userID uint) error {
// 	if image == nil {
// 		return fmt.Errorf("image is nil")
// 	}

// 	if !model.UserCanEdit(db, image, userID) {
// 		return fmt.Errorf(
// 			"cannot delete %s: user %d is not owner or moderator",
// 			image.Filename,
// 			userID,
// 		)
// 	}

// 	if err := db.Delete(&image).Error; err != nil {
// 		return fmt.Errorf("cannot delete %s: %w", image.Filename, err)
// 	}

// 	return nil
// }

// func (model *ImageRepository) containsImageHash(db *gorm.DB, hash string) bool {
// 	var count int64
// 	db.Model(&models.ImageMetadata{}).Where("hash = ?", hash).Count(&count)
// 	return count > 0
// }

// func (model *ImageRepository) imageNameExists(db *gorm.DB, name string) bool {
// 	var count int64
// 	db.Model(&models.ImageMetadata{}).Where("filename = ?", name).Count(&count)
// 	return count > 0
// }
