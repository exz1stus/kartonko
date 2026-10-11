package image

import (
	"testing"
	"time"

	apperrors "server/internal/errors"
	"server/internal/user"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestEnforceUploadRate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.User{}))
	require.NoError(t, db.Exec("CREATE TABLE image_metadata (id integer primary key, user_id integer, created_at datetime, deleted_at datetime)").Error)

	regular := user.User{Username: "regular"}
	moderator := user.User{Username: "moderator", Privilege: user.Moderator}
	require.NoError(t, db.Create(&regular).Error)
	require.NoError(t, db.Create(&moderator).Error)
	require.NoError(t, enforceUploadRate(db, regular.ID))
	for i := 0; i < 49; i++ {
		require.NoError(t, db.Exec("INSERT INTO image_metadata (user_id, created_at) VALUES (?, ?)", regular.ID, time.Now()).Error)
	}
	require.NoError(t, db.Exec("INSERT INTO image_metadata (user_id, created_at, deleted_at) VALUES (?, ?, ?)", regular.ID, time.Now(), time.Now()).Error)
	require.ErrorIs(t, enforceUploadRate(db, regular.ID), apperrors.ErrUploadRateLimited)
	for i := 0; i < 50; i++ {
		require.NoError(t, db.Exec("INSERT INTO image_metadata (user_id, created_at) VALUES (?, ?)", moderator.ID, time.Now()).Error)
	}
	require.NoError(t, enforceUploadRate(db, moderator.ID))
	require.NoError(t, db.Exec("UPDATE image_metadata SET created_at = ? WHERE id = 1", time.Now().Add(-2*time.Hour)).Error)
	require.NoError(t, enforceUploadRate(db, regular.ID))
}
