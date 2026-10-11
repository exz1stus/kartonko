package log

import (
	"time"

	"server/internal/user"

	"gorm.io/datatypes"
)

type ActorResponse struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Privilege  string `json:"privilege"`
	PictureURL string `json:"picture_url"`
} // @name AuditActor

func newActorResponse(actor *user.User) *ActorResponse {
	if actor.ID == 0 {
		return nil
	}
	return &ActorResponse{
		ID:         actor.ID,
		Username:   actor.Username,
		Privilege:  actor.Privilege.String(),
		PictureURL: actor.PictureURL,
	}
}

type EntryResponse struct {
	ID            uint           `json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	Action        string         `json:"action"`
	ObjectType    string         `json:"object_type"`
	AffectedObjID uint           `json:"affected_obj_id"`
	UserID        uint           `json:"user_id"`
	Actor         *ActorResponse `json:"actor"`
	Data          datatypes.JSON `json:"data" swaggertype:"object"`
	RevertedAt    *time.Time     `json:"reverted_at"`
	RevertedByID  *uint          `json:"reverted_by_id"`
} // @name EntryResponse

func NewEntryResponse(entry *AuditEntry) EntryResponse {
	return EntryResponse{
		ID:            entry.ID,
		CreatedAt:     entry.CreatedAt,
		Action:        entry.Action,
		ObjectType:    entry.ObjectType,
		AffectedObjID: entry.AffectedObjID,
		UserID:        entry.UserID,
		Actor:         newActorResponse(&entry.User),
		Data:          entry.Data,
		RevertedAt:    entry.RevertedAt,
		RevertedByID:  entry.RevertedByID,
	}
}
