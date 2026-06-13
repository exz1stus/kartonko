package models

import (
	"fmt"

	"gorm.io/gorm"
)

type ImageMetadata struct {
	gorm.Model
	Hash     string `json:"hash" gorm:"not null"`
	Filename string `json:"filename" gorm:"not null"`
	Tags     []Tag  `json:"tags"  gorm:"many2many:image_tags;constraint:OnDelete:CASCADE;"`
	Format   string `json:"format" gorm:"not null"`
	Width    uint   `json:"width" gorm:"not null"`
	Height   uint   `json:"height" gorm:"not null"`
	UserID   uint   `json:"user_id" gorm:"not null;default:1"`
	User     User   `json:"user"`
}

type ImageModel struct {
	db   *gorm.DB
	Tags *TagModel
}

type ImageQuery struct {
	NameContains string
	WithTags     []string
	User         *User
}

func EmptyQuery() ImageQuery {
	return ImageQuery{
		NameContains: "",
		WithTags:     []string{},
		User:         nil,
	}
}

func (model *ImageModel) applyFilters(db *gorm.DB, query ImageQuery) *gorm.DB {
	if query.NameContains != "" {
		db = db.Where("filename LIKE ?", "%"+query.NameContains+"%")
	}

	if len(query.WithTags) > 0 {
		db = db.Joins("JOIN image_tags ON image_tags.image_id = image_metadata.id").
			Joins("JOIN tags ON tags.id = image_tags.tag_id").
			Where("tags.name IN ?", query.WithTags).
			Group("image_metadata.id").
			Having("COUNT(DISTINCT tags.id) = ?", len(query.WithTags))
	}

	if query.User != nil {
		db = db.Where("user_id = ?", query.User.ID)
	}

	return db
}

func (model *ImageModel) SearchImages(query ImageQuery, cursor int, limit int) ([]ImageMetadata, error) {
	var images []ImageMetadata
	db := model.db.Preload("Tags").Model(&ImageMetadata{}).Order("image_metadata.id desc")

	db = model.applyFilters(db, query)

	db = db.Limit(limit).Offset(cursor).Find(&images)

	err := db.Error
	if err != nil {
		return nil, err
	}

	return images, nil
}

// TODO: Rename to ConstructImageMetadata
func (model *ImageModel) ConstructImage(name string, tagsNames []string, format string, width uint, height uint, userID uint) *ImageMetadata {
	tags := ConstructTagsByNames(tagsNames)
	image := &ImageMetadata{
		Filename: name,
		Tags:     tags,
		Format:   format,
		Width:    width,
		Height:   height,
		UserID:   userID,
	}

	return image
}

func (model *ImageModel) AddImage(image *ImageMetadata) error {

	if model.containsImageHash(image.Hash) {
		return fmt.Errorf("image with hash %s already exists", image.Hash)
	}

	if model.imageNameExists(image.Filename) {
		return fmt.Errorf("image with name %s already exists", image.Filename)
	}

	if err := model.Tags.CheckTags(image.Tags); err != nil {
		return err
	}

	tagNames := TagsToStrings(image.Tags)
	var dbTags []Tag
	if err := model.db.Where("name IN ?", tagNames).Find(&dbTags).Error; err != nil {
		return fmt.Errorf("failed to retrieve tags: %w", err)
	}

	image.Tags = dbTags

	if err := model.db.Create(image).Error; err != nil {
		return fmt.Errorf("failed to insert image: %w", err)
	}

	return nil
}

func (model *ImageModel) GetImageByHash(hash string) (*ImageMetadata, error) {
	var img ImageMetadata
	result := model.db.Preload("Tags").Where("hash = ?", hash).First(&img)

	if result.Error != nil {
		return nil, fmt.Errorf("failed retrieving image from the db: %w", result.Error)
	}

	if img.Hash == "" {
		return nil, fmt.Errorf("image with hash %s not found", hash)
	}

	return &img, nil
}

func (model *ImageModel) GetImageByID(id uint64) (*ImageMetadata, error) {
	var img ImageMetadata
	result := model.db.Preload("Tags").Where("id = ?", id).First(&img)

	if result.Error != nil {
		return nil, fmt.Errorf("failed retrieving image from the db: %w", result.Error)
	}

	if img.Hash == "" {
		return nil, fmt.Errorf("image with id %d not found", id)
	}

	return &img, nil
}

func (model *ImageModel) GetImageByName(name string) (*ImageMetadata, error) {
	var img ImageMetadata
	result := model.db.Preload("Tags").Where("filename = ?", name).First(&img)

	if result.Error != nil {
		return nil, fmt.Errorf("failed retrieving image from the db: %w", result.Error)
	}

	if img.Hash == "" {
		return nil, fmt.Errorf("image with name %s not found", name)
	}

	return &img, nil
}

func (model *ImageModel) GetImages(cursor int, limit int) ([]ImageMetadata, error) {
	var images []ImageMetadata
	result := model.db.Model(&ImageMetadata{}).Preload("Tags").Order("id desc").Limit(limit).Offset(cursor).Find(&images)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve images: %w", result.Error)
	}

	return images, nil
}

func (model *ImageModel) DeleteImagesByQuery(query ImageQuery, userID uint) ([]ImageMetadata, error) {
	var images []ImageMetadata
	db := model.db.Preload("Tags").Model(&ImageMetadata{}).Order("image_metadata.id desc")

	db = model.applyFilters(db, query)

	if err := db.Find(&images).Error; err != nil {
		return nil, fmt.Errorf("failed to find images for deletion: %w", err)
	}

	if len(images) == 0 {
		return nil, nil
	}

	for _, img := range images {
		if !model.UserCanEdit(&img, userID) {
			return nil, fmt.Errorf(
				"permission denied: user %d cannot delete image '%s' (ID: %d)",
				userID, img.Filename, img.ID,
			)
		}
	}

	if err := model.db.Delete(&images).Error; err != nil {
		return nil, fmt.Errorf("failed to delete images: %w", err)
	}

	return images, nil
}

func (model *ImageModel) UserCanEdit(image *ImageMetadata, userID uint) bool {
	if image.UserID == userID {
		return true
	}

	var user User

	err := model.db.
		Where("id = ? AND privileage = ?", userID, Moderator).
		First(&user).Error

	if err != nil {
		return false
	}

	return true
}

func (model *ImageModel) DeleteImage(image *ImageMetadata, userID uint) error {
	if image == nil {
		return fmt.Errorf("image is nil")
	}

	if !model.UserCanEdit(image, userID) {
		return fmt.Errorf(
			"cannot delete %s: user %d is not owner or moderator",
			image.Filename,
			userID,
		)
	}

	if err := model.db.Delete(&image).Error; err != nil {
		return fmt.Errorf("cannot delete %s: %w", image.Filename, err)
	}

	return nil
}

func (model *ImageModel) containsImageHash(hash string) bool {
	var count int64
	model.db.Model(&ImageMetadata{}).Where("hash = ?", hash).Count(&count)
	return count > 0
}

func (model *ImageModel) imageNameExists(name string) bool {
	var count int64
	model.db.Model(&ImageMetadata{}).Where("filename = ?", name).Count(&count)
	return count > 0
}
