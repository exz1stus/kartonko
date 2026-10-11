package log

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ActionSummary struct {
	Action     string `json:"action"`
	ObjectType string `json:"object_type"`
	BoardName  string `json:"board_name,omitempty"`
	BoardSlug  string `json:"board_slug,omitempty"`
	Count      uint   `json:"count"`
} // @name AuditActionSummary

type GroupResponse struct {
	ID              uint            `json:"id"`
	UserID          uint            `json:"user_id"`
	Actor           *ActorResponse  `json:"actor"`
	LastAt          time.Time       `json:"last_at"`
	EntryCount      uint            `json:"entry_count"`
	RevertableCount uint            `json:"revertable_count"`
	Summaries       []ActionSummary `json:"summaries"`
} // @name AuditGroup

func newGroupResponse(group *AuditGroup) (GroupResponse, error) {
	var summaries []ActionSummary
	if err := json.Unmarshal(group.Summaries, &summaries); err != nil {
		return GroupResponse{}, err
	}
	return GroupResponse{
		ID: group.ID, UserID: group.UserID, Actor: newActorResponse(&group.User),
		LastAt: group.LastAt, EntryCount: group.EntryCount,
		RevertableCount: group.RevertableCount, Summaries: summaries,
	}, nil
}

func belongsToGroup(group *AuditGroup, entry *AuditEntry) bool {
	if group.UserID != entry.UserID || group.LastAt.IsZero() {
		return false
	}
	previous, current := group.LastAt.UTC(), entry.CreatedAt.UTC()
	return previous.Year() == current.Year() && previous.YearDay() == current.YearDay() &&
		current.Sub(previous) >= 0 && current.Sub(previous) < 5*time.Minute
}

func reversible(entry *AuditEntry) bool {
	if entry.RevertedAt != nil {
		return false
	}
	switch entry.ObjectType {
	case "image":
		return entry.Action == "create"
	case "board":
		return entry.Action == "add_image" || entry.Action == "remove_image"
	case "tag":
		return entry.Action == "create" || entry.Action == "delete" || entry.Action == "update"
	default:
		return false
	}
}

func addSummary(group *AuditGroup, entry *AuditEntry) error {
	var summaries []ActionSummary
	if len(group.Summaries) > 0 {
		if err := json.Unmarshal(group.Summaries, &summaries); err != nil {
			return err
		}
	}
	boardName := ""
	boardSlug := ""
	if entry.ObjectType == "board" {
		var boardData struct {
			BoardName string `json:"board_name"`
			BoardSlug string `json:"board_slug"`
		}
		_ = json.Unmarshal(entry.Data, &boardData)
		boardName = boardData.BoardName
		boardSlug = boardData.BoardSlug
	}
	updated := false
	for i := range summaries {
		if summaries[i].Action == entry.Action && summaries[i].ObjectType == entry.ObjectType && summaries[i].BoardName == boardName {
			if summaries[i].BoardSlug == "" {
				summaries[i].BoardSlug = boardSlug
			}
			summaries[i].Count++
			updated = true
			break
		}
	}
	if !updated {
		summaries = append(summaries, ActionSummary{Action: entry.Action, ObjectType: entry.ObjectType, BoardName: boardName, BoardSlug: boardSlug, Count: 1})
	}
	data, err := json.Marshal(summaries)
	if err != nil {
		return err
	}
	group.Summaries = datatypes.JSON(data)
	group.EntryCount++
	if reversible(entry) {
		group.RevertableCount++
	}
	group.LastAt = entry.CreatedAt
	group.LastEntryID = entry.ID
	return nil
}

func createGroupedEntry(tx *gorm.DB, entry *AuditEntry) error {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	var group AuditGroup
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Order("last_entry_id DESC").First(&group).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || !belongsToGroup(&group, entry) {
		if err := tx.Create(entry).Error; err != nil {
			return err
		}
		group = AuditGroup{ID: entry.ID, UserID: entry.UserID}
		if err := addSummary(&group, entry); err != nil {
			return err
		}
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		entry.GroupID = group.ID
		return tx.Model(entry).Update("group_id", group.ID).Error
	}
	entry.GroupID = group.ID
	if err := tx.Create(entry).Error; err != nil {
		return err
	}
	if err := addSummary(&group, entry); err != nil {
		return err
	}
	return tx.Model(&group).Updates(map[string]any{
		"last_at": group.LastAt, "last_entry_id": group.LastEntryID,
		"entry_count": group.EntryCount, "revertable_count": group.RevertableCount,
		"summaries": group.Summaries,
	}).Error
}

// BackfillGroups assigns existing development log rows to stable groups once.
func BackfillGroups(db *gorm.DB) error {
	var entries []AuditEntry
	if err := db.Where("group_id = 0").Order("id ASC").Find(&entries).Error; err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var group AuditGroup
		var firstID, lastID uint
		flush := func() error {
			if firstID == 0 {
				return nil
			}
			if err := tx.Create(&group).Error; err != nil {
				return err
			}
			return tx.Model(&AuditEntry{}).Where("id BETWEEN ? AND ? AND group_id = 0", firstID, lastID).
				Update("group_id", group.ID).Error
		}
		for i := range entries {
			entry := &entries[i]
			if firstID == 0 || !belongsToGroup(&group, entry) {
				if err := flush(); err != nil {
					return fmt.Errorf("backfill audit group: %w", err)
				}
				group = AuditGroup{ID: entry.ID, UserID: entry.UserID}
				firstID = entry.ID
			}
			lastID = entry.ID
			if err := addSummary(&group, entry); err != nil {
				return err
			}
		}
		return flush()
	})
}
