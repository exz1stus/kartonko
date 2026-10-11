package processing

import (
	"time"

	"gorm.io/gorm"
)

const (
	EmbeddingJob = "embed_image"
	Pending      = "pending"
	Running      = "running"
	Ready        = "ready"
	Failed       = "failed"
	Cancelled    = "cancelled"
)

// Job is stored with the image transaction so committed images always have work queued.
type Job struct {
	ID          uint      `gorm:"primaryKey"`
	ImageID     uint      `gorm:"not null;uniqueIndex:idx_job_image_kind"`
	Kind        string    `gorm:"not null;uniqueIndex:idx_job_image_kind"`
	Status      string    `gorm:"not null;index"`
	Attempts    uint      `gorm:"not null;default:0"`
	AvailableAt time.Time `gorm:"not null;index"`
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Queue interface {
	Enqueue(tx *gorm.DB, imageID uint) error
	Cancel(tx *gorm.DB, imageIDs []uint) error
}

type GormQueue struct{}

func (GormQueue) Enqueue(tx *gorm.DB, imageID uint) error {
	return tx.Create(&Job{ImageID: imageID, Kind: EmbeddingJob, Status: Pending, AvailableAt: time.Now()}).Error
}

func (GormQueue) Cancel(tx *gorm.DB, imageIDs []uint) error {
	if len(imageIDs) == 0 {
		return nil
	}
	return tx.Model(&Job{}).Where("image_id IN ?", imageIDs).Update("status", Cancelled).Error
}
