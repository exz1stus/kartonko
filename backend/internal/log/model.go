package log

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"server/internal/user"
	"time"
)

type AuditEntry struct {
	gorm.Model

	UserID        uint           `json:"user_id" gorm:"not null"`
	Action        string         `json:"action" gorm:"not null"`
	ObjectType    string         `json:"object_type" gorm:"not null"`
	AffectedObjID uint           `json:"affected_obj_id" gorm:"not null"`
	Data          datatypes.JSON `json:"data"`
	RevertedAt    *time.Time     `json:"reverted_at"`
	RevertedByID  *uint          `json:"reverted_by_id"`

	User user.User `json:"user" gorm:"foreignKey:UserID"`
}
