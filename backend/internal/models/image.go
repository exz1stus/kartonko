package models

import (
	"server/pkg/image"

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

// image model from DB must guarantee parsable format
func (img *ImageMetadata) ParseFormat() image.Format {
	format, _ := image.ParseFormat(img.Format)
	return format
}

func ConstructImageMetadata(name string, tagsNames []string, format string, width uint, height uint, userID uint) *ImageMetadata {
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
