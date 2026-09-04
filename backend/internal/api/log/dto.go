package log

import (
	"time"

	logpkg "server/internal/log"

	"gorm.io/datatypes"
)

type EntryResponse struct {
	ID            uint           `json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	Action        string         `json:"action"`
	ObjectType    string         `json:"object_type"`
	AffectedObjID uint           `json:"affected_obj_id"`
	UserID        uint           `json:"user_id"`
	Data          datatypes.JSON `json:"data" swaggertype:"object"`
} // @name EntryResponse

func FromServiceEntry(entry *logpkg.EntryResponse) EntryResponse {
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

func FromServiceEntries(entries []logpkg.EntryResponse) []EntryResponse {
	resp := make([]EntryResponse, len(entries))
	for i, e := range entries {
		resp[i] = FromServiceEntry(&e)
	}
	return resp
}
