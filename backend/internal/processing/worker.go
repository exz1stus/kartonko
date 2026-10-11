package processing

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ImageSource interface {
	GetRawImageData(context.Context, uint) (io.ReadCloser, error)
}

type Embedder interface {
	Upsert(context.Context, uint, []byte) error
}

type Worker struct {
	db       *gorm.DB
	images   ImageSource
	embedder Embedder
}

func NewWorker(db *gorm.DB, images ImageSource, embedder Embedder) *Worker {
	return &Worker{db: db, images: images, embedder: embedder}
}

// RunOne locks only one job while the external embedding call runs. Deletion
// takes the same lock before removing the image, so it cannot race an upsert.
func (w *Worker) RunOne(ctx context.Context) (bool, error) {
	worked := false
	err := w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var jobs []Job
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("kind = ? AND status IN ? AND available_at <= ?", EmbeddingJob, []string{Pending, Failed}, time.Now()).
			Order("available_at, id").Limit(1).Find(&jobs).Error; err != nil {
			return err
		}
		if len(jobs) == 0 {
			return nil
		}
		worked = true
		job := jobs[0]
		attempts := job.Attempts + 1
		if err := tx.Model(&job).Updates(map[string]any{"status": Running, "attempts": attempts}).Error; err != nil {
			return err
		}

		if err := w.process(ctx, job.ImageID); err != nil {
			status := Pending
			if attempts >= 5 {
				status = Failed
			}
			message := err.Error()
			if len(message) > 1000 {
				message = message[:1000]
			}
			// Failed jobs remain eligible so an embedding outage cannot strand images.
			// Cap the delay and arithmetic even when the service is down for a long time.
			retryAttempt := attempts
			if retryAttempt > 8 {
				retryAttempt = 8
			}
			backoff := time.Duration(5*retryAttempt*retryAttempt) * time.Second
			if backoff > 5*time.Minute {
				backoff = 5 * time.Minute
			}
			if err := tx.Model(&job).Updates(map[string]any{
				"status": status, "available_at": time.Now().Add(backoff), "last_error": message,
			}).Error; err != nil {
				return err
			}
			log.Printf("embedding image %d failed (attempt %d): %s", job.ImageID, attempts, message)
			if status == Failed {
				return tx.Table("image_metadata").Where("id = ? AND deleted_at IS NULL", job.ImageID).
					Update("embedding_status", Failed).Error
			}
			return nil
		}
		if err := tx.Model(&job).Updates(map[string]any{"status": Ready, "last_error": ""}).Error; err != nil {
			return err
		}
		return tx.Table("image_metadata").Where("id = ? AND deleted_at IS NULL", job.ImageID).
			Update("embedding_status", Ready).Error
	})
	return worked, err
}

func (w *Worker) process(ctx context.Context, imageID uint) error {
	reader, err := w.images.GetRawImageData(ctx, imageID)
	if err != nil {
		return fmt.Errorf("read image: %w", err)
	}
	data, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return fmt.Errorf("read image bytes: %w", readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close image data: %w", closeErr)
	}
	if err := w.embedder.Upsert(ctx, imageID, data); err != nil {
		return fmt.Errorf("upsert embedding: %w", err)
	}
	return nil
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		worked, err := w.RunOne(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("embedding worker: %v", err)
		}
		if worked && err == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
	}
}
