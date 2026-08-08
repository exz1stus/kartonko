package board

import (
	"server/internal/image"
	"server/internal/user"

	"gorm.io/gorm"
)

type Board struct {
	gorm.Model

	Name string `json:"name"`

	Images []image.ImageMetadata `json:"images"  gorm:"many2many:board_images;constraint:OnDelete:CASCADE;"`

	UserID uint      `json:"user_id" gorm:"not null"`
	User   user.User `json:"user" gorm:"foreignKey:UserID"`
}