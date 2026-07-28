package models

import (
	"time"

	"gorm.io/datatypes"
)

type AuditEntry struct {
	ID        uint      `gorm:"primarykey, not null, autoincrement"`
	CreatedAt time.Time `gorm:"not null"`

	UserID        uint           `json:"user_id" gorm:"not null"`
	Action        string         `json:"action" gorm:"not null"`
	ObjectType    string         `json:"object_type" gorm:"not null"`
	AffectedObjID uint           `json:"affected_obj_id" gorm:"not null"`
	Data          datatypes.JSON `json:"data"`

	User User `json:"user" gorm:"foreignKey:UserID"`
}
