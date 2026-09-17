package board

import (
	"server/internal/image"
	"server/internal/user"
	"time"

	"gorm.io/gorm"
)

type Board struct {
	gorm.Model

	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`

	UserID uint      `json:"user_id" gorm:"not null"`
	User   user.User `json:"user" gorm:"foreignKey:UserID"`

	Items []BoardItem `json:"items"  gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE;"`
}

type BoardItem struct {
	ID uint `gorm:"primaryKey"`

	BoardID uint `gorm:"not null;uniqueIndex:idx_board_image"`
	ImageID uint `gorm:"not null;uniqueIndex:idx_board_image"`

	Board Board               `gorm:"foreignKey:BoardID;constraint:OnDelete:CASCADE;"`
	Image image.ImageMetadata `gorm:"foreignKey:ImageID;constraint:OnDelete:CASCADE;"`

	CreatedAt time.Time
}
