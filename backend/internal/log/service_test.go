package log_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	apperrors "server/internal/errors"
	"server/internal/log"
	"server/internal/tag"
	"server/internal/user"
)

func TestLogRequiresActionData(t *testing.T) {
	service := log.NewLogService(nil, nil, nil, nil)
	require.ErrorIs(t, service.Log(nil, "delete", "image", 1, 1, nil), apperrors.ErrBadRequest)
}

func TestRevertTagChanges(t *testing.T) {
	for _, action := range []string{"create", "delete", "update"} {
		t.Run(action, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			require.NoError(t, err)
			require.NoError(t, db.AutoMigrate(&user.User{}, &tag.Tag{}, &log.AuditEntry{}))
			actor := user.User{Username: "moderator", Privilege: user.Moderator}
			require.NoError(t, db.Create(&actor).Error)
			entryTag := tag.Tag{Name: "original", UserID: actor.ID}
			require.NoError(t, db.Create(&entryTag).Error)
			service := log.NewLogService(log.NewLogRepository(db), db, nil, nil)
			var data any = map[string]string{"name": "original"}
			switch action {
			case "delete":
				require.NoError(t, db.Delete(&entryTag).Error)
			case "update":
				require.NoError(t, db.Model(&entryTag).Update("name", "renamed").Error)
				data = map[string]string{"before": "original", "after": "renamed"}
			}
			require.NoError(t, service.Log(db, action, "tag", actor.ID, entryTag.ID, data))
			entries, err := service.GetEntries(0, 10)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			require.NotNil(t, entries[0].Actor)
			require.Equal(t, "moderator", entries[0].Actor.Username)
			_, err = service.Revert(entries[0].ID, actor.ID)
			require.NoError(t, err)
			var actual tag.Tag
			require.NoError(t, db.Unscoped().First(&actual, entryTag.ID).Error)
			if action == "create" {
				require.True(t, actual.DeletedAt.Valid)
			} else {
				require.False(t, actual.DeletedAt.Valid)
			}
			if action == "update" {
				require.Equal(t, "original", actual.Name)
			}
			_, err = service.Revert(entries[0].ID, actor.ID)
			require.Error(t, err)
			entries, err = service.GetEntries(0, 10)
			require.NoError(t, err)
			require.Len(t, entries, 2)
			require.Equal(t, "revert", entries[0].Action)
			require.NotNil(t, entries[1].RevertedAt)
			var payload map[string]any
			require.NoError(t, json.Unmarshal(entries[0].Data, &payload))
			require.Equal(t, float64(entries[1].ID), payload["entry_id"])
		})
	}
}

func TestRevertRejectsChangedTag(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.User{}, &tag.Tag{}, &log.AuditEntry{}))
	actor := user.User{Username: "moderator"}
	require.NoError(t, db.Create(&actor).Error)
	entryTag := tag.Tag{Name: "newer", UserID: actor.ID}
	require.NoError(t, db.Create(&entryTag).Error)
	service := log.NewLogService(log.NewLogRepository(db), db, nil, nil)
	require.NoError(t, service.Log(db, "update", "tag", actor.ID, entryTag.ID, map[string]string{"before": "old", "after": "intermediate"}))
	entries, err := service.GetEntries(0, 10)
	require.NoError(t, err)
	_, err = service.Revert(entries[0].ID, actor.ID)
	require.Error(t, err)
	var actual tag.Tag
	require.NoError(t, db.First(&actual, entryTag.ID).Error)
	require.Equal(t, "newer", actual.Name)
}

func TestRevertBoardMembership(t *testing.T) {
	for _, action := range []string{"add_image", "remove_image"} {
		t.Run(action, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			require.NoError(t, err)
			require.NoError(t, db.AutoMigrate(&user.User{}, &log.AuditEntry{}))
			for _, statement := range []string{
				"CREATE TABLE boards (id integer primary key, deleted_at datetime)",
				"CREATE TABLE image_metadata (id integer primary key, deleted_at datetime)",
				"CREATE TABLE board_items (id integer primary key, board_id integer, image_id integer, created_at datetime)",
				"INSERT INTO boards (id) VALUES (4)",
				"INSERT INTO image_metadata (id) VALUES (7)",
			} {
				require.NoError(t, db.Exec(statement).Error)
			}
			if action == "add_image" {
				require.NoError(t, db.Exec("INSERT INTO board_items (board_id, image_id) VALUES (4, 7)").Error)
			}
			actor := user.User{Username: "moderator"}
			require.NoError(t, db.Create(&actor).Error)
			service := log.NewLogService(log.NewLogRepository(db), db, nil, nil)
			require.NoError(t, service.Log(db, action, "board", actor.ID, 4, map[string]any{"image_id": 7, "board_name": "Example"}))
			entries, err := service.GetEntries(0, 10)
			require.NoError(t, err)
			_, err = service.Revert(entries[0].ID, actor.ID)
			require.NoError(t, err)
			var count int64
			require.NoError(t, db.Table("board_items").Count(&count).Error)
			if action == "add_image" {
				require.EqualValues(t, 0, count)
			} else {
				require.EqualValues(t, 1, count)
			}
		})
	}
}

type imageObjectCleaner struct {
	hash, format string
}

func (c *imageObjectCleaner) DeleteImageObjects(_ context.Context, hash, format string) error {
	c.hash, c.format = hash, format
	return nil
}

type imageEmbeddingCleaner struct {
	id uint
}

func (c *imageEmbeddingCleaner) Delete(_ context.Context, id uint) error {
	c.id = id
	return nil
}

func TestRevertImageUpload(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.User{}, &log.AuditEntry{}))
	require.NoError(t, db.Exec("CREATE TABLE image_metadata (id integer primary key, filename text, hash text, format text, deleted_at datetime)").Error)
	require.NoError(t, db.Exec("INSERT INTO image_metadata (id, filename, hash, format) VALUES (7, 'photo.jpg', 'hash-value', 'jpeg')").Error)
	actor := user.User{Username: "moderator", Privilege: user.Moderator}
	require.NoError(t, db.Create(&actor).Error)
	objects := &imageObjectCleaner{}
	embeddings := &imageEmbeddingCleaner{}
	service := log.NewLogService(log.NewLogRepository(db), db, objects, embeddings)
	require.NoError(t, service.Log(db, "create", "image", actor.ID, 7, map[string]string{"name": "photo.jpg"}))
	entries, err := service.GetEntries(0, 10)
	require.NoError(t, err)
	result, err := service.Revert(entries[0].ID, actor.ID)
	require.NoError(t, err)
	require.NotNil(t, result.RevertedAt)
	var deletedAt *string
	require.NoError(t, db.Table("image_metadata").Select("deleted_at").Where("id = 7").Scan(&deletedAt).Error)
	require.NotNil(t, deletedAt)
	require.Equal(t, uint(7), embeddings.id)
	require.Equal(t, "hash-value", objects.hash)
	require.Equal(t, "jpeg", objects.format)
	_, err = service.Revert(entries[0].ID, actor.ID)
	require.ErrorIs(t, err, apperrors.ErrConflict)
}

func TestRevertImageUploadAfterRename(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.User{}, &log.AuditEntry{}))
	require.NoError(t, db.Exec("CREATE TABLE image_metadata (id integer primary key, filename text, hash text, format text, deleted_at datetime)").Error)
	require.NoError(t, db.Exec("INSERT INTO image_metadata (id, filename, hash, format) VALUES (7, 'renamed.jpg', 'hash-value', 'jpeg')").Error)
	actor := user.User{Username: "moderator"}
	require.NoError(t, db.Create(&actor).Error)
	objects := &imageObjectCleaner{}
	embeddings := &imageEmbeddingCleaner{}
	service := log.NewLogService(log.NewLogRepository(db), db, objects, embeddings)
	require.NoError(t, service.Log(db, "create", "image", actor.ID, 7, map[string]string{"name": "photo.jpg"}))
	entries, err := service.GetEntries(0, 10)
	require.NoError(t, err)
	_, err = service.Revert(entries[0].ID, actor.ID)
	require.NoError(t, err)
	require.Equal(t, uint(7), embeddings.id)
	require.Equal(t, "hash-value", objects.hash)
}
