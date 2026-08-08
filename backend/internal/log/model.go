package log

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"server/internal/user"
)

type AuditEntry struct {
	gorm.Model

	UserID        uint           `json:"user_id" gorm:"not null"`
	Action        string         `json:"action" gorm:"not null"`
	ObjectType    string         `json:"object_type" gorm:"not null"`
	AffectedObjID uint           `json:"affected_obj_id" gorm:"not null"`
	Data          datatypes.JSON `json:"data"`

	User user.User `json:"user" gorm:"foreignKey:UserID"`
}