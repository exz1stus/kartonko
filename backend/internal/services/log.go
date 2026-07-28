package services

import (
	"encoding/json"
	"server/internal/api/dto"
	"server/internal/models"
	"server/internal/repositories"

	"gorm.io/gorm"
)

type LogService interface {
	GetEntries(cursor int, limit int) ([]dto.EntryResponse, error)
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
	logs repositories.LogRepository
}

func NewLogService(logs repositories.LogRepository) LogService {
	return &logService{logs: logs}
}

func (s *logService) Log(
	tx *gorm.DB,
	action string,
	objectType string,
	userID uint,
	objectID uint,
	data any,
) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	entry := models.AuditEntry{
		UserID:        userID,
		Action:        action,
		ObjectType:    objectType,
		AffectedObjID: objectID,
		Data:          jsonData,
	}

	repo := s.logs.WithTx(tx)

	return repo.Create(&entry)
}

func (s *logService) GetEntries(cursor int, limit int) ([]dto.EntryResponse, error) {
	entries, err := s.logs.GetEntries(cursor, limit)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.EntryResponse, 0, len(entries))
	for _, entry := range entries {
		resp = append(resp, dto.ConstructEntryResponse(&entry))
	}
	return resp, nil
}
