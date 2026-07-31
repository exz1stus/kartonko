package models

import (
	"gorm.io/gorm"
)

type Board struct {
	gorm.Model

	Name string `json:"name"`

	Images []ImageMetadata `json:"images"  gorm:"many2many:board_images;constraint:OnDelete:CASCADE;"`

	UserID uint `json:"user_id" gorm:"not null"`
	User   User `json:"user" gorm:"foreignKey:UserID"`
}
