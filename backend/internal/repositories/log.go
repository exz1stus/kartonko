package repositories

import (
	"fmt"
	"server/internal/models"

	"gorm.io/gorm"
)

type LogRepository interface {
	WithTx(tx *gorm.DB) LogRepository
	Create(entry *models.AuditEntry) error
	GetEntries(cursor int, limit int) ([]models.AuditEntry, error)
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

func (r *logRepository) Create(entry *models.AuditEntry) error {
	return r.db.Model(&models.AuditEntry{}).Create(entry).Error
}

func (r *logRepository) GetEntries(cursor int, limit int) ([]models.AuditEntry, error) {
	var responses []models.AuditEntry
	if err := r.db.Model(&models.AuditEntry{}).
		Order("audit_entries.id DESC").
		Limit(limit).
		Offset(cursor).
		Scan(&responses).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve entries: %w", err)
	}
	return responses, nil
}
