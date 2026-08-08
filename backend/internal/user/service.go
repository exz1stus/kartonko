package user

import (
	"fmt"
)

type UserService interface {
	Create(user *User) error
	CreateByGoogle(username string, email string, googleID string, pictureURL string) (*User, error)
	CreateByRegistration(username string, hashedPassword string) (*User, error)

	GetByID(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByProviderID(id string) (*User, error)

	SetPrivilege(userID uint, privilege Privilege) error

	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
}

type userService struct {
	users UserRepository
}

func NewUserService(users UserRepository) UserService {
	return &userService{users}
}

func (s *userService) Create(user *User) error {
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

func (s *userService) CreateByGoogle(username string, email string, googleID string, pictureURL string) (*User, error) {
	if email == "" {
		return nil, fmt.Errorf("email is empty")
	}

	if googleID == "" {
		return nil, fmt.Errorf("googleID is empty")
	}

	user := &User{
		Username:   username,
		Email:      email,
		Privilege:  Unprivileged,
		Provider:   "google",
		ProviderID: googleID,
		PictureURL: pictureURL,
	}

	err := s.Create(user)
	return user, err
}

func (s *userService) CreateByRegistration(username string, hashedPassword string) (*User, error) {
	if hashedPassword == "" {
		return nil, fmt.Errorf("hashed password is empty")
	}

	user := &User{
		Username:       username,
		HashedPassword: hashedPassword,
		Privilege:      Unprivileged,
	}

	err := s.Create(user)
	return user, err
}

func (s *userService) GetByID(id uint) (*User, error) {
	return s.users.GetByID(id)
}

func (s *userService) GetByUsername(username string) (*User, error) {
	return s.users.GetByUsername(username)
}

func (s *userService) GetByEmail(email string) (*User, error) {
	return s.users.GetByEmail(email)
}

func (s *userService) GetByProviderID(id string) (*User, error) {
	return s.users.GetByProviderID(id)
}

func (s *userService) SetPrivilege(userID uint, privilege Privilege) error {
	return s.users.SetPrivilege(userID, privilege)
}

func (s *userService) ExistsByUsername(username string) (bool, error) {
	return s.users.ExistsByUsername(username)
}

func (s *userService) ExistsByEmail(email string) (bool, error) {
	return s.users.ExistsByEmail(email)
}
