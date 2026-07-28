package dto

import (
	"server/internal/models"
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

func ConstructEntryResponse(entry *models.AuditEntry) EntryResponse {
	return EntryResponse{
		entry.ID,
		entry.CreatedAt,
		entry.Action,
		entry.ObjectType,
		entry.AffectedObjID,
		entry.UserID,
		entry.Data,
	}
}
