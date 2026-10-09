package image_test

import (
	"context"
	stderrors "errors"
	"server/internal/api/transaction"
	"server/internal/errors"
	"server/internal/image"
	imageMocks "server/internal/image/mocks"
	"server/internal/tag"
	"server/internal/user"
	userMocks "server/internal/user/mocks"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUpdate_RejectsDuplicateFilename(t *testing.T) {
	ctx := context.Background()
	owner := &user.User{Model: gorm.Model{ID: 7}}
	current := &image.ImageMetadata{Model: gorm.Model{ID: 11}, Filename: "current", UserID: owner.ID}
	existing := &image.ImageMetadata{Model: gorm.Model{ID: 12}, Filename: "taken", UserID: owner.ID}

	images := imageMocks.NewMockImageRepository(t)
	images.EXPECT().Get(mock.Anything, current.ID).Return(current, nil).Once()
	images.EXPECT().GetByName(mock.Anything, existing.Filename).Return(existing, nil).Once()

	users := userMocks.NewMockUserRepository(t)
	users.EXPECT().Get(mock.Anything, owner.ID).Return(owner, nil).Once()

	service := image.NewImageService(images, users, nil, nil, nil, transaction.TestRunner{})
	_, err := service.Update(ctx, owner.ID, current.ID, &image.ImagePatchRequest{Filename: &existing.Filename})

	require.Error(t, err)
	require.True(t, stderrors.Is(err, errors.ErrDuplicateName))
}

func TestUpdate_ReplacesTagsInTransaction(t *testing.T) {
	ctx := context.Background()
	owner := &user.User{Model: gorm.Model{ID: 7}}
	current := &image.ImageMetadata{Model: gorm.Model{ID: 11}, Filename: "current", UserID: owner.ID}
	filename := "renamed"
	tagNames := []string{"cat", "animal"}

	images := imageMocks.NewMockImageRepository(t)
	images.EXPECT().Get(mock.Anything, current.ID).Return(current, nil).Once()
	images.EXPECT().GetByName(mock.Anything, filename).Return(nil, gorm.ErrRecordNotFound).Once()

	transactionImages := imageMocks.NewMockImageRepository(t)
	images.EXPECT().WithTx((*gorm.DB)(nil)).Return(transactionImages).Once()
	transactionImages.EXPECT().
		Update(mock.Anything, mock.MatchedBy(func(updated *image.ImageMetadata) bool {
			return updated.ID == current.ID && updated.Filename == filename && updated.Tags == nil
		})).
		Return(nil).Once()
	transactionImages.EXPECT().
		ReplaceTags(mock.Anything, current.ID, tag.TagsByNames(tagNames)).
		Return(nil).Once()

	users := userMocks.NewMockUserRepository(t)
	users.EXPECT().Get(mock.Anything, owner.ID).Return(owner, nil).Once()

	service := image.NewImageService(images, users, nil, nil, nil, transaction.TestRunner{})
	updated, err := service.Update(ctx, owner.ID, current.ID, &image.ImagePatchRequest{
		Filename: &filename,
		Tags:     tagNames,
	})

	require.NoError(t, err)
	require.Equal(t, filename, updated.Filename)
	require.Equal(t, tagNames, tag.TagsToStrings(updated.Tags))
}
