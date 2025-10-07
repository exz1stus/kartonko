package models

import (
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuditLog struct {
	db *gorm.DB
}

func (model *AuditLog) AddEntryType(name string) error {
	entryType := &EntryType{Name: name}
	err := model.db.Model(&EntryType{}).Create(entryType).Error
	if err != nil {
		return fmt.Errorf("failed to insert entry type: %v", err.Error())
	}
	return nil
}

func (model *AuditLog) OnImageCreated(image *Image, user *User) error {
	return model.addEntry("image_created", user.Model.ID, image.ID, nil)
}

func (model *AuditLog) addEntry(entryTypeName string, userID uint, affectedObjID uint, data datatypes.JSON) error {
	var entryType EntryType
	err := model.db.Model(&EntryType{}).Where("name = ?", entryTypeName).First(&entryType).Error
	if err != nil {
		return fmt.Errorf("failed to retrieve entry type: %v", err.Error())
	}

	entry := &AuditEntry{
		UserID:        userID,
		EntryTypeID:   entryType.ID,
		AffectedObjID: affectedObjID,
		Data:          data,
	}

	err = model.db.Model(&AuditEntry{}).Create(entry).Error
	if err != nil {
		return fmt.Errorf("failed to insert entry: %v", err.Error())
	}

	return nil
}

func (model *AuditLog) GetEntries(cursor int, limit int) ([]AuditEntry, error) {
	var entries []AuditEntry
	err := model.db.Model(&AuditEntry{}).Order("id desc").Limit(limit).Offset(cursor).Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve entries: %v", err.Error())
	}
	return entries, nil
}
