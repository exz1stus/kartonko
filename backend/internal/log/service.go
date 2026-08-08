package log

import (
	"encoding/json"

	"gorm.io/gorm"
)

type LogService interface {
	GetEntries(cursor int, limit int) ([]EntryResponse, error)
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
	logs LogRepository
}

func NewLogService(logs LogRepository) LogService {
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
		resp = append(resp, ConstructEntryResponse(&entry))
	}
	return resp, nil
}