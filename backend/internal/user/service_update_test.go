package user_test

import (
	"context"
	stderrors "errors"
	"server/internal/errors"
	"server/internal/user"
	userMocks "server/internal/user/mocks"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUpdate_RejectsDuplicateUsername(t *testing.T) {
	ctx := context.Background()
	current := &user.User{Model: gorm.Model{ID: 7}, Username: "current"}
	existing := &user.User{Model: gorm.Model{ID: 8}, Username: "taken"}

	users := userMocks.NewMockUserRepository(t)
	users.EXPECT().Get(mock.Anything, current.ID).Return(current, nil).Twice()
	users.EXPECT().GetByUsername(mock.Anything, existing.Username).Return(existing, nil).Once()

	service := user.NewUserService(users)
	_, err := service.Update(ctx, current.ID, current.ID, &user.UserPatchRequest{Username: &existing.Username})

	require.Error(t, err)
	require.True(t, stderrors.Is(err, errors.ErrDuplicateName))
}
