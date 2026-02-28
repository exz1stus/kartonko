package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuditLog struct {
	db *gorm.DB
}

func (model *AuditLog) addEntryType(name string) error {
	entryType := &EntryType{Name: name}
	err := model.db.Model(&EntryType{}).Create(entryType).Error
	if err != nil {
		return fmt.Errorf("failed to insert entry type: %w", err)
	}
	return nil
}

func (model *AuditLog) addEntry(entryTypeName string, userID uint, affectedObjID uint, data datatypes.JSON) error {
	var entryType EntryType
	err := model.db.Model(&EntryType{}).Where("name = ?", entryTypeName).First(&entryType).Error

	// if the entry type doesn't exist, create it
	if err == gorm.ErrRecordNotFound {
		err = model.addEntryType(entryTypeName)
		if err != nil {
			return fmt.Errorf("failed lazy creating entry: %s", entryTypeName)
		}
		err = model.db.Model(&EntryType{}).Where("name = ?", entryTypeName).First(&entryType).Error
	}

	if err != nil {
		return fmt.Errorf("failed to retrieve entry type: %w", err)
	}

	entry := &AuditEntry{
		UserID:        userID,
		EntryTypeID:   entryType.ID,
		AffectedObjID: affectedObjID,
		Data:          data,
	}

	err = model.db.Model(&AuditEntry{}).Create(entry).Error
	if err != nil {
		return fmt.Errorf("failed to insert entry: %w", err)
	}

	return nil
}

type EntryResponse struct {
	ID            uint           `json:"id"`
	CreatedAt     time.Time      `json:"created_at"`
	EntryType     string         `json:"entry_type"`
	UserID        uint           `json:"user_id"`
	AffectedObjID uint           `json:"affected_obj_id"`
	Data          datatypes.JSON `json:"data"`
}

func (model *AuditLog) GetEntries(cursor int, limit int) ([]EntryResponse, error) {
	var responses []EntryResponse

	err := model.db.
		Table("audit_entries").
		Select(`
			audit_entries.id,
			audit_entries.created_at,
			audit_entries.user_id,
			entry_types.name AS entry_type,
			audit_entries.affected_obj_id,
			audit_entries.data
		`).
		Joins(`
			LEFT JOIN entry_types
			ON entry_types.id = audit_entries.entry_type_id
		`).
		Order("audit_entries.id DESC").
		Limit(limit).
		Offset(cursor).
		Scan(&responses).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve entries: %w", err)
	}
	return responses, nil
}

type ImageEntryData struct {
	Name string `json:"name"`
}

func (model *AuditLog) AddImageCreated(image *Image, user *User) error {
	data, err := json.Marshal(&ImageEntryData{Name: image.Filename})
	if err != nil {
		return fmt.Errorf("failed to marshal image log entry data: %w", err)
	}
	return model.addEntry("image_created", user.Model.ID, image.ID, datatypes.JSON(data))
}

func (model *AuditLog) AddImageDeleted(image *Image, user *User) error {
	data, err := json.Marshal(&ImageEntryData{Name: image.Filename})
	if err != nil {
		return fmt.Errorf("failed to marshal image log entry data: %w", err)
	}
	return model.addEntry("image_deleted", user.Model.ID, image.ID, datatypes.JSON(data))
}
