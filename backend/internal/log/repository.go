package log

import (
	"fmt"

	"gorm.io/gorm"
)

type LogRepository interface {
	WithTx(tx *gorm.DB) LogRepository
	Create(entry *AuditEntry) error
	GetEntries(cursor int, limit int) ([]AuditEntry, error)
	GetGroups(cursor int, limit int) ([]AuditGroup, error)
	GetGroupEntries(groupID uint) ([]AuditEntry, error)
}

func (r *logRepository) GetGroups(cursor int, limit int) ([]AuditGroup, error) {
	var groups []AuditGroup
	err := r.db.Preload("User", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Order("id DESC").Limit(limit).Offset(cursor).Find(&groups).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve audit groups: %w", err)
	}
	return groups, nil
}

func (r *logRepository) GetGroupEntries(groupID uint) ([]AuditEntry, error) {
	var entries []AuditEntry
	err := r.db.Model(&AuditEntry{}).
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Where("group_id = ?", groupID).Order("id DESC").Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve audit group entries: %w", err)
	}
	return entries, nil
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
	return r.db.Transaction(func(tx *gorm.DB) error { return createGroupedEntry(tx, entry) })
}

func (r *logRepository) GetEntries(cursor int, limit int) ([]AuditEntry, error) {
	var responses []AuditEntry
	if err := r.db.Model(&AuditEntry{}).
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Order("audit_entries.id DESC").
		Limit(limit).
		Offset(cursor).
		Find(&responses).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve entries: %w", err)
	}
	return responses, nil
}
