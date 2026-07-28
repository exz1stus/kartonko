package services

import (
	"fmt"
	"server/internal/models"
	"server/internal/repositories"
)

type UserService interface {
	Create(user *models.User) error
	CreateByGoogle(username string, email string, googleID string, pictureURL string) (*models.User, error)
	CreateByRegistration(username string, hashedPassword string) (*models.User, error)

	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByProviderID(id string) (*models.User, error)

	SetPrivilege(userID uint, privilege models.Privilege) error

	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
}

type userService struct {
	users  repositories.UserRepository
	images repositories.ImageRepository
}

func NewUserService(users repositories.UserRepository, images repositories.ImageRepository) UserService {
	return &userService{users, images}
}

func (s *userService) Create(user *models.User) error {
	if user.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	exists, err := s.ExistsByUsername(user.Username)
	if err != nil {
		return fmt.Errorf("check duplicate username error: %w", err)
	}
	if exists {
		return fmt.Errorf("user with username %s already exists", user.Username)
	}

	if err := s.users.Create(user); err != nil {
		return fmt.Errorf("failed creating user: %w", err)
	}

	return nil
}

func (s *userService) CreateByGoogle(username string, email string, googleID string, pictureURL string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email is empty")
	}

	if googleID == "" {
		return nil, fmt.Errorf("googleID is empty")
	}

	user := &models.User{
		Username:   username,
		Email:      email,
		Privilege:  models.Unprivileged,
		Provider:   "google",
		ProviderID: googleID,
		PictureURL: pictureURL,
	}

	err := s.Create(user)
	return user, err
}

func (s *userService) CreateByRegistration(username string, hashedPassword string) (*models.User, error) {
	if hashedPassword == "" {
		return nil, fmt.Errorf("hashed password is empty")
	}

	user := &models.User{
		Username:       username,
		HashedPassword: hashedPassword,
		Privilege:      models.Unprivileged,
	}

	err := s.Create(user)
	return user, err
}

func (s *userService) GetByID(id uint) (*models.User, error) {
	return s.users.GetByID(id)
}

func (s *userService) GetByUsername(username string) (*models.User, error) {
	return s.users.GetByUsername(username)
}

func (s *userService) GetByEmail(email string) (*models.User, error) {
	return s.users.GetByEmail(email)
}

func (s *userService) GetByProviderID(id string) (*models.User, error) {
	return s.users.GetByProviderID(id)
}

func (s *userService) SetPrivilege(userID uint, privilege models.Privilege) error {
	return s.users.SetPrivilege(userID, privilege)
}

func (s *userService) ExistsByUsername(username string) (bool, error) {
	return s.users.ExistsByUsername(username)
}

func (s *userService) ExistsByEmail(email string) (bool, error) {
	return s.users.ExistsByEmail(email)
}
