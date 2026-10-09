package user

import (
	"context"
	stderrors "errors"
	"fmt"
	"server/internal/errors"
	"strings"

	"gorm.io/gorm"
)

type UserService interface {
	Create(ctx context.Context, user *User) error
	CreateByGoogle(ctx context.Context, username string, email string, googleID string, pictureURL string) (*User, error)
	CreateByRegistration(ctx context.Context, username string, hashedPassword string) (*User, error)

	Get(ctx context.Context, userID uint) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByProviderID(ctx context.Context, providerID string) (*User, error)

	Update(ctx context.Context, userID, targetID uint, req *UserPatchRequest) (*User, error)

	Delete(ctx context.Context, userID, targetID uint) error

	SetPrivilege(ctx context.Context, userID uint, privilege Privilege) error

	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type userService struct {
	users UserRepository
}

func NewUserService(users UserRepository) UserService {
	return &userService{users}
}

func (s *userService) Create(ctx context.Context, user *User) error {
	user.Username = strings.TrimSpace(user.Username)
	if user.Username == "" {
		return errors.ErrBadRequest
	}

	exists, err := s.ExistsByUsername(ctx, user.Username)
	if err != nil {
		return fmt.Errorf("check duplicate username error: %w", err)
	}
	if exists {
		return fmt.Errorf("%w: %s", errors.ErrDuplicateName, user.Username)
	}

	if err := s.users.Create(ctx, user); err != nil {
		return fmt.Errorf("failed creating user: %w", err)
	}

	return nil
}

func (s *userService) CreateByGoogle(ctx context.Context, username string, email string, googleID string, pictureURL string) (*User, error) {
	if email == "" {
		return nil, fmt.Errorf("email is empty")
	}

	if googleID == "" {
		return nil, fmt.Errorf("googleID is empty")
	}

	user := &User{
		Username:   username,
		Email:      &email,
		Privilege:  Unprivileged,
		Provider:   "google",
		ProviderID: &googleID,
		PictureURL: pictureURL,
	}

	err := s.Create(ctx, user)
	return user, err
}

func (s *userService) CreateByRegistration(ctx context.Context, username string, hashedPassword string) (*User, error) {
	if hashedPassword == "" {
		return nil, fmt.Errorf("hashed password is empty")
	}

	user := &User{
		Username:       username,
		HashedPassword: hashedPassword,
		Privilege:      Unprivileged,
		Provider:       "registration",
	}

	err := s.Create(ctx, user)
	return user, err
}

func (s *userService) checkUserPermission(ctx context.Context, targetID uint, userID uint) error {
	user, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}

	if !CanEdit(user.ID, user.Privilege, targetID) {
		return errors.ErrPermissionDenied
	}

	return nil
}

func (s *userService) Get(ctx context.Context, id uint) (*User, error) {
	return s.users.Get(ctx, id)
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*User, error) {
	return s.users.GetByUsername(ctx, username)
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*User, error) {
	return s.users.GetByEmail(ctx, email)
}

func (s *userService) GetByProviderID(ctx context.Context, providerID string) (*User, error) {
	return s.users.GetByProviderID(ctx, providerID)
}

func (s *userService) Update(ctx context.Context, userID uint, targetID uint, req *UserPatchRequest) (*User, error) {
	if err := s.checkUserPermission(ctx, targetID, userID); err != nil {
		return nil, err
	}

	user, err := s.Get(ctx, targetID)
	if err != nil {
		return nil, err
	}

	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if username == "" {
			return nil, errors.ErrBadRequest
		}
		if username != user.Username {
			existing, lookupErr := s.GetByUsername(ctx, username)
			switch {
			case lookupErr == nil && existing.ID != user.ID:
				return nil, fmt.Errorf("%w: %s", errors.ErrDuplicateName, username)
			case lookupErr != nil && !stderrors.Is(lookupErr, gorm.ErrRecordNotFound):
				return nil, fmt.Errorf("check duplicate username: %w", lookupErr)
			}
		}
		user.Username = username
	}

	if req.PictureURL != nil {
		user.PictureURL = *req.PictureURL
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Delete(ctx context.Context, userID uint, targetID uint) error {
	if err := s.checkUserPermission(ctx, targetID, userID); err != nil {
		return err
	}

	return s.users.Delete(ctx, targetID)
}

func (s *userService) SetPrivilege(ctx context.Context, userID uint, privilege Privilege) error {
	return s.users.SetPrivilege(ctx, userID, privilege)
}

func (s *userService) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return s.users.ExistsByUsername(ctx, username)
}

func (s *userService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return s.users.ExistsByEmail(ctx, email)
}
