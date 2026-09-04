package log

import (
	"time"

	"gorm.io/datatypes"
)

type EntryResponse struct {
	ID            uint
	CreatedAt     time.Time
	Action        string
	ObjectType    string
	AffectedObjID uint
	UserID        uint
	Data          datatypes.JSON
}

func NewEntryResponse(entry *AuditEntry) EntryResponse {
	return EntryResponse{
		ID:            entry.ID,
		CreatedAt:     entry.CreatedAt,
		Action:        entry.Action,
		ObjectType:    entry.ObjectType,
		AffectedObjID: entry.AffectedObjID,
		UserID:        entry.UserID,
		Data:          entry.Data,
	}
}
