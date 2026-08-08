package log

import (
	"time"

	"gorm.io/datatypes"
)

type EntryResponse struct {
	ID            uint           `json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	Action        string         `json:"action"`
	ObjectType    string         `json:"object_type"`
	AffectedObjID uint           `json:"affected_obj_id"`
	UserID        uint           `json:"user_id"`
	Data          datatypes.JSON `json:"data"`
}

func ConstructEntryResponse(entry *AuditEntry) EntryResponse {
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
