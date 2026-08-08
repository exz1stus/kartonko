package log

import (
	"fmt"

	"gorm.io/gorm"
)

type LogRepository interface {
	WithTx(tx *gorm.DB) LogRepository
	Create(entry *AuditEntry) error
	GetEntries(cursor int, limit int) ([]AuditEntry, error)
}

type logRepository struct {
	db *gorm.DB
}

func NewLogRepository(db *gorm.DB) LogRepository {
	return &logRepository{db: db}
}

func (r *logRepository) WithTx(tx *gorm.DB) LogRepository {
	return NewLogRepository(tx)
}

func (r *logRepository) Create(entry *AuditEntry) error {
	return r.db.Model(&AuditEntry{}).Create(entry).Error
}

func (r *logRepository) GetEntries(cursor int, limit int) ([]AuditEntry, error) {
	var responses []AuditEntry
	if err := r.db.Model(&AuditEntry{}).
		Order("audit_entries.id DESC").
		Limit(limit).
		Offset(cursor).
		Scan(&responses).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve entries: %w", err)
	}
	return responses, nil
}
