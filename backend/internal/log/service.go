package log

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	apperrors "server/internal/errors"
	"server/internal/processing"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LogService interface {
	GetEntries(cursor int, limit int) ([]EntryResponse, error)
	Revert(entryID, actorID uint) (EntryResponse, error)
	Log(
		tx *gorm.DB,
		action string,
		objectType string,
		userID uint,
		objectID uint,
		data any,
	) error
}

type logService struct {
	logs    LogRepository
	db      *gorm.DB
	objects interface {
		DeleteImageObjects(context.Context, string, string) error
	}
	embeddings interface {
		Delete(context.Context, uint) error
	}
}

func NewLogService(
	logs LogRepository,
	db *gorm.DB,
	objects interface {
		DeleteImageObjects(context.Context, string, string) error
	},
	embeddings interface {
		Delete(context.Context, uint) error
	},
) LogService {
	return &logService{logs: logs, db: db, objects: objects, embeddings: embeddings}
}

func (s *logService) Log(
	tx *gorm.DB,
	action string,
	objectType string,
	userID uint,
	objectID uint,
	data any,
) error {
	if data == nil {
		return apperrors.ErrBadRequest
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	entry := AuditEntry{
		UserID:        userID,
		Action:        action,
		ObjectType:    objectType,
		AffectedObjID: objectID,
		Data:          jsonData,
	}

	repo := s.logs.WithTx(tx)

	return repo.Create(&entry)
}

func (s *logService) GetEntries(cursor int, limit int) ([]EntryResponse, error) {
	entries, err := s.logs.GetEntries(cursor, limit)
	if err != nil {
		return nil, err
	}

	resp := make([]EntryResponse, 0, len(entries))
	for _, entry := range entries {
		resp = append(resp, NewEntryResponse(&entry))
	}
	return resp, nil
}

func (s *logService) GetGroups(cursor int, limit int) ([]GroupResponse, error) {
	groups, err := s.logs.GetGroups(cursor, limit)
	if err != nil {
		return nil, err
	}
	responses := make([]GroupResponse, 0, len(groups))
	for i := range groups {
		response, err := newGroupResponse(&groups[i])
		if err != nil {
			return nil, fmt.Errorf("decode audit group summary: %w", err)
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func (s *logService) GetGroupEntries(groupID uint) ([]EntryResponse, error) {
	entries, err := s.logs.GetGroupEntries(groupID)
	if err != nil {
		return nil, err
	}
	responses := make([]EntryResponse, 0, len(entries))
	for i := range entries {
		responses = append(responses, NewEntryResponse(&entries[i]))
	}
	return responses, nil
}

// Revert applies the inverse only while the affected object still has the state
// recorded by the entry. The row lock and marker prevent concurrent reverts.
func (s *logService) Revert(entryID, actorID uint) (EntryResponse, error) {
	if s.db == nil {
		return EntryResponse{}, fmt.Errorf("audit database is unavailable")
	}
	var result EntryResponse
	var deletedImage *imageToDelete
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var entry AuditEntry
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&entry, entryID).Error; err != nil {
			return err
		}
		if entry.RevertedAt != nil {
			return apperrors.ErrConflict
		}
		tagAction := entry.ObjectType == "tag" && (entry.Action == "create" || entry.Action == "delete" || entry.Action == "update")
		boardAction := entry.ObjectType == "board" && (entry.Action == "add_image" || entry.Action == "remove_image")
		imageAction := entry.ObjectType == "image" && entry.Action == "create"
		if !tagAction && !boardAction && !imageAction {
			return apperrors.ErrBadRequest
		}
		if imageAction {
			if err := (processing.GormQueue{}).Cancel(tx, []uint{entry.AffectedObjID}); err != nil {
				return err
			}
			var err error
			deletedImage, err = revertImageCreate(tx, &entry)
			if err != nil {
				return err
			}
		} else if boardAction {
			if err := revertBoardItem(tx, &entry); err != nil {
				return err
			}
		} else {
			if err := revertTag(tx, &entry); err != nil {
				return err
			}
		}
		wasRevertable := reversible(&entry)
		now := time.Now()
		entry.RevertedAt, entry.RevertedByID = &now, &actorID
		if err := tx.Model(&entry).Updates(map[string]any{"reverted_at": now, "reverted_by_id": actorID}).Error; err != nil {
			return err
		}
		if wasRevertable {
			if err := tx.Model(&AuditGroup{}).Where("id = ? AND revertable_count > 0", entry.GroupID).
				UpdateColumn("revertable_count", gorm.Expr("revertable_count - 1")).Error; err != nil {
				return err
			}
		}
		var original map[string]any
		_ = json.Unmarshal(entry.Data, &original)
		original["entry_id"] = entry.ID
		payload, err := json.Marshal(original)
		if err != nil {
			return err
		}
		if err := s.logs.WithTx(tx).Create(&AuditEntry{UserID: actorID, Action: "revert", ObjectType: entry.ObjectType, AffectedObjID: entry.AffectedObjID, Data: datatypes.JSON(payload)}); err != nil {
			return err
		}
		result = NewEntryResponse(&entry)
		return nil
	})
	if err != nil {
		return EntryResponse{}, err
	}
	if deletedImage != nil {
		ctx := context.Background()
		if err := s.embeddings.Delete(ctx, deletedImage.id); err != nil {
			return result, fmt.Errorf("image reverted; remove embedding: %w", err)
		}
		if err := s.objects.DeleteImageObjects(ctx, deletedImage.hash, deletedImage.format); err != nil {
			return result, fmt.Errorf("image reverted; remove image objects: %w", err)
		}
	}
	return result, nil
}

type imageToDelete struct {
	id     uint
	hash   string
	format string
}

func revertImageCreate(tx *gorm.DB, entry *AuditEntry) (*imageToDelete, error) {
	var data struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(entry.Data, &data); err != nil || data.Name == "" {
		return nil, apperrors.ErrBadRequest
	}
	var current struct {
		ID        uint
		Hash      string
		Format    string
		DeletedAt gorm.DeletedAt
	}
	if err := tx.Unscoped().Table("image_metadata").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", entry.AffectedObjID).Take(&current).Error; err != nil {
		return nil, err
	}
	if current.DeletedAt.Valid {
		return nil, apperrors.ErrConflict
	}
	if err := tx.Table("image_metadata").Where("id = ?", current.ID).Update("deleted_at", time.Now()).Error; err != nil {
		return nil, fmt.Errorf("revert image upload: %w", err)
	}
	return &imageToDelete{id: current.ID, hash: current.Hash, format: current.Format}, nil
}

func revertTag(tx *gorm.DB, entry *AuditEntry) error {
	var data struct {
		Name   string `json:"name"`
		Before string `json:"before"`
		After  string `json:"after"`
	}
	if err := json.Unmarshal(entry.Data, &data); err != nil {
		return apperrors.ErrBadRequest
	}
	var current struct {
		ID        uint
		Name      string
		DeletedAt gorm.DeletedAt
	}
	if err := tx.Unscoped().Table("tags").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", entry.AffectedObjID).Take(&current).Error; err != nil {
		return err
	}
	var changes map[string]any
	switch entry.Action {
	case "create":
		if data.Name == "" || current.DeletedAt.Valid || current.Name != data.Name {
			return apperrors.ErrConflict
		}
		changes = map[string]any{"deleted_at": time.Now()}
	case "delete":
		if data.Name == "" || !current.DeletedAt.Valid || current.Name != data.Name {
			return apperrors.ErrConflict
		}
		changes = map[string]any{"deleted_at": nil}
	case "update":
		if data.Before == "" || data.After == "" || current.DeletedAt.Valid || current.Name != data.After {
			return apperrors.ErrConflict
		}
		changes = map[string]any{"name": data.Before}
	}
	if err := tx.Table("tags").Where("id = ?", entry.AffectedObjID).Updates(changes).Error; err != nil {
		return fmt.Errorf("revert tag: %w", err)
	}
	return nil
}

func revertBoardItem(tx *gorm.DB, entry *AuditEntry) error {
	var data struct {
		ImageID uint `json:"image_id"`
	}
	if err := json.Unmarshal(entry.Data, &data); err != nil || data.ImageID == 0 {
		return apperrors.ErrBadRequest
	}
	var count int64
	if err := tx.Table("boards").Where("id = ? AND deleted_at IS NULL", entry.AffectedObjID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return apperrors.ErrConflict
	}
	if err := tx.Table("image_metadata").Where("id = ? AND deleted_at IS NULL", data.ImageID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return apperrors.ErrConflict
	}
	if err := tx.Table("board_items").Where("board_id = ? AND image_id = ?", entry.AffectedObjID, data.ImageID).Count(&count).Error; err != nil {
		return err
	}
	if entry.Action == "add_image" {
		if count == 0 {
			return apperrors.ErrConflict
		}
		return tx.Exec("DELETE FROM board_items WHERE board_id = ? AND image_id = ?", entry.AffectedObjID, data.ImageID).Error
	}
	if count != 0 {
		return apperrors.ErrConflict
	}
	return tx.Table("board_items").Create(map[string]any{"board_id": entry.AffectedObjID, "image_id": data.ImageID, "created_at": time.Now()}).Error
}
