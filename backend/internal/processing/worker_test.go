package processing_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"server/internal/processing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type imageSource struct{ data []byte }

func (s imageSource) GetRawImageData(_ context.Context, _ uint) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.data)), nil
}

type embedder struct {
	id   uint
	data []byte
	err  error
}

func (e *embedder) Upsert(_ context.Context, id uint, data []byte) error {
	e.id, e.data = id, data
	return e.err
}

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:processing?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&processing.Job{}))
	require.NoError(t, db.Exec("CREATE TABLE IF NOT EXISTS image_metadata (id integer primary key, embedding_status text, deleted_at datetime)").Error)
	require.NoError(t, db.Exec("DELETE FROM image_metadata").Error)
	require.NoError(t, db.Exec("DELETE FROM jobs").Error)
	return db
}

func TestWorkerCompletesQueuedEmbedding(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.Exec("INSERT INTO image_metadata (id, embedding_status) VALUES (42, 'pending')").Error)
	require.NoError(t, (processing.GormQueue{}).Enqueue(db, 42))
	emb := &embedder{}
	worked, err := processing.NewWorker(db, imageSource{data: []byte("image")}, emb).RunOne(context.Background())
	require.NoError(t, err)
	require.True(t, worked)
	require.Equal(t, uint(42), emb.id)
	require.Equal(t, []byte("image"), emb.data)
	var job processing.Job
	require.NoError(t, db.Where("image_id = 42").First(&job).Error)
	require.Equal(t, processing.Ready, job.Status)
	var status string
	require.NoError(t, db.Table("image_metadata").Select("embedding_status").Where("id = 42").Scan(&status).Error)
	require.Equal(t, processing.Ready, status)
}

func TestWorkerRetriesFailedJobAndRecovers(t *testing.T) {
	db := testDB(t)
	require.NoError(t, db.Exec("INSERT INTO image_metadata (id, embedding_status) VALUES (43, 'pending')").Error)
	require.NoError(t, (processing.GormQueue{}).Enqueue(db, 43))
	emb := &embedder{err: errors.New("model unavailable")}
	worker := processing.NewWorker(db, imageSource{data: []byte("image")}, emb)
	for attempt := 1; attempt <= 5; attempt++ {
		worked, err := worker.RunOne(context.Background())
		require.NoError(t, err)
		require.True(t, worked)
		var job processing.Job
		require.NoError(t, db.Where("image_id = 43").First(&job).Error)
		require.Equal(t, uint(attempt), job.Attempts)
		if attempt >= 5 {
			require.Equal(t, processing.Failed, job.Status)
		} else {
			require.Equal(t, processing.Pending, job.Status)
		}
		require.NoError(t, db.Model(&job).Update("available_at", time.Now().Add(-time.Second)).Error)
	}
	var status string
	require.NoError(t, db.Table("image_metadata").Select("embedding_status").Where("id = 43").Scan(&status).Error)
	require.Equal(t, processing.Failed, status)

	emb.err = nil
	worked, err := worker.RunOne(context.Background())
	require.NoError(t, err)
	require.True(t, worked)
	var job processing.Job
	require.NoError(t, db.Where("image_id = 43").First(&job).Error)
	require.Equal(t, uint(6), job.Attempts)
	require.Equal(t, processing.Ready, job.Status)
	require.Empty(t, job.LastError)
	require.NoError(t, db.Table("image_metadata").Select("embedding_status").Where("id = 43").Scan(&status).Error)
	require.Equal(t, processing.Ready, status)
}

func TestCancelledJobIsNotProcessed(t *testing.T) {
	db := testDB(t)
	require.NoError(t, (processing.GormQueue{}).Enqueue(db, 44))
	require.NoError(t, (processing.GormQueue{}).Cancel(db, []uint{44}))
	emb := &embedder{}
	worked, err := processing.NewWorker(db, imageSource{data: []byte("image")}, emb).RunOne(context.Background())
	require.NoError(t, err)
	require.False(t, worked)
	require.Zero(t, emb.id)
}
