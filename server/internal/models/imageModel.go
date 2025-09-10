package models

import (
	"fmt"
	"mime/multipart"
	"server/image"
	"server/internal/env"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Image struct {
	gorm.Model
	Hash     string `json:"hash" gorm:"primaryKey;unique;not null"`
	Filename string `json:"filename" gorm:"not null"`
	Tags     []Tag  `json:"tags"  gorm:"many2many:image_tags"`
	Format   string `json:"format" gorm:"not null"`
}

type ImageModel struct {
	db   *gorm.DB
	Tags *TagModel
}

type ImageQuery struct {
	NameContains string   `json:"nameContains"`
	WithTags     []string `json:"withTags"`
}

func (model *ImageModel) SearchImages(query ImageQuery, cursor int, limit int) ([]Image, error) {
	var images []Image
	result := model.db.Model(&Image{}).Order("id desc")

	println(query.NameContains)

	if query.NameContains != "" {
		result = result.Where("filename LIKE ?", "%"+query.NameContains+"%").Find(&images)
	}

	if len(query.WithTags) > 0 {
		result = result.
			Joins("JOIN image_tags ON image_tags.image_id = images.id").
			Joins("JOIN tags ON tags.id = image_tags.tag_id").
			Where("tags.name IN ?", query.WithTags).
			Group("images.id").
			Having("COUNT(DISTINCT tags.id) = ?", len(query.WithTags)).
			Find(&images)
	}

	result = result.Limit(limit).Offset(cursor).Find(&images)

	if result.Error != nil {
		return nil, result.Error
	}

	return images, nil
}

func (model *ImageModel) ConstructImage(name string, tagsNames []string, format string) *Image {
	tags := ConstructTagsByNames(tagsNames)
	image := &Image{Filename: name, Tags: tags, Format: format}
	model.AddAutoTags(image)
	return image
}

func (model *ImageModel) AddAutoTags(image *Image) {
	if image == nil {
		return
	}

	if image.Format == "gif" {
		image.Tags = append(image.Tags, Tag{Name: "gif"})
	}
}

func (model *ImageModel) AddImage(image *Image) error {
	if model.containsImageHash(image.Hash) {
		return fmt.Errorf("image with hash %s already exists", image.Hash)
	}

	if model.imageNameExists(image.Filename) {
		return fmt.Errorf("image with name %s already exists", image.Filename)
	}

	if err := model.Tags.CheckTags(image.Tags); err != nil {
		return err
	}

	result := model.db.Create(image)
	if result.Error != nil {
		return fmt.Errorf("failed to insert image: %v", result.Error)
	}

	return nil
}

func (model *ImageModel) GetImageByName(name string) (*Image, error) {
	var img Image
	result := model.db.Where("filename = ?", name).First(&img)

	if result.Error != nil {
		return nil, fmt.Errorf("failed retrieving image from the db: %v", result.Error)
	}

	if img.Hash == "" {
		return nil, fmt.Errorf("image with name %s not found", name)
	}

	return &img, nil
}

func (model *ImageModel) SaveImage(img *Image, file *multipart.FileHeader, c *gin.Context) error {
	hash, err := image.HashFile(file)
	if err != nil {
		return err
	}
	img.Hash = hash

	file.Filename = img.Filename + "." + img.Format
	dst := env.GetEnvString("UPLOADS_PATH") + file.Filename
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return err
	}

	if err := model.AddImage(img); err != nil {
		return fmt.Errorf("failed to add image to storage: %w", err)
	}

	return nil
}

func (model *ImageModel) GetImages(cursor int, limit int) ([]Image, error) {
	var images []Image
	result := model.db.Model(&Image{}).Order("id desc").Limit(limit).Offset(cursor).Find(&images)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve images: %v", result.Error)
	}

	return images, nil
}

func (model *ImageModel) containsImageHash(hash string) bool {
	var count int64
	model.db.Model(&Image{}).Where("hash = ?", hash).Count(&count)
	return count > 0
}

func (model *ImageModel) imageNameExists(name string) bool {
	var count int64
	model.db.Model(&Image{}).Where("filename = ?", name).Count(&count)
	return count > 0
}
