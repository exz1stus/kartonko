package models

import (
	"time"

	"gorm.io/datatypes"
)

type EntryType struct {
	ID   uint   `json:"id" gorm:"primaryKey, not null, autoincrement"`
	Name string `json:"name" gorm:"not null, unique"`
}

type AuditEntry struct {
	ID        uint      `gorm:"primarykey, not null, autoincrement"`
	CreatedAt time.Time `gorm:"not null"`

	UserID        uint           `json:"user_id" gorm:"not null"`
	EntryTypeID   uint           `json:"entry_type_id" gorm:"not null"`
	AffectedObjID uint           `json:"affected_obj_id" gorm:"not null"`
	Data          datatypes.JSON `json:"data"`

	User User      `json:"user" gorm:"foreignKey:UserID"`
	Type EntryType `json:"type" gorm:"foreignKey:EntryTypeID"`
}

type ImageAddedEntry struct{}
type ImageEditedEntry struct {
	Old Image `json:"old"`
	New Image `json:"new"`
}
