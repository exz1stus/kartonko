package repository_integration_tests

import (
	"context"
	"server/internal/image"
	"server/internal/tag"
	"server/internal/user"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageRepository_AttachTags_TagsExist(t *testing.T) {
	ctx, cleanup := setupPostgreContext(t)
	defer cleanup()
	db := ctx.Db

	user := user.User{
		Username:  "testuser",
		Privilege: user.Unprivileged,
	}
	require.NoError(t, db.Create(&user).Error)

	repo := image.NewImageRepository(db)

	img := image.ImageMetadata{
		Filename: "cat.png",
		Hash:     "some hash",
		Format:   image.FormatPNG.String(),
		Width:    10,
		Height:   10,
		UserID:   1,
	}
	require.NoError(t, db.Create(&img).Error)

	tags := []tag.Tag{
		{Name: "cat", UserID: 1},
		{Name: "animal", UserID: 1},
	}
	require.NoError(t, db.Create(&tags).Error)

	err := repo.AttachTags(context.Background(), img.ID, tags)

	require.NoError(t, err)

	var result image.ImageMetadata
	require.NoError(t,
		db.Preload("Tags").
			First(&result, img.ID).
			Error,
	)

	require.ElementsMatch(t,
		[]string{"cat", "animal"},
		tag.TagsToStrings(result.Tags),
	)
}

func TestImageRepository_AttachTags_RejectsTagDoesntExist(t *testing.T) {
	ctx, cleanup := setupPostgreContext(t)
	defer cleanup()
	db := ctx.Db

	user := user.User{
		Username:  "testuser",
		Privilege: user.Unprivileged,
	}
	require.NoError(t, db.Create(&user).Error)

	repo := image.NewImageRepository(db)

	img := image.ImageMetadata{
		Filename: "cat.png",
		Hash:     "some hash",
		Format:   image.FormatPNG.String(),
		Width:    10,
		Height:   10,
		UserID:   1,
	}
	require.NoError(t, db.Create(&img).Error)

	tags := []tag.Tag{
		{Name: "animal", UserID: 1},
	}
	require.NoError(t, db.Create(&tags).Error)

	err := repo.AttachTags(context.Background(), img.ID, append(tags, tag.Tag{Name: "cat", UserID: 1}))

	require.Error(t, err)
}
