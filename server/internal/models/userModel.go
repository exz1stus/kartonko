package models

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Privileage int

const (
	Unprivileaged Privileage = iota
	Moderator
)

func (p Privileage) String() string {
	return [...]string{"Unprivileaged", "Moderator"}[p]
}

type User struct {
	gorm.Model
	Username       string     `json:"username" gorm:"unique;not null"`
	Email          string     `json:"email" gorm:"unique"`
	HashedPassword string     `json:"hashedPassword"`
	Privileage     Privileage `json:"privileage" gorm:"not null"`
	Provider       string     `json:"provider"`
	ProviderID     string     `json:"provider_id" gorm:"unique"`
	PictureURL     string     `json:"picture_url"`
}

type UserModel struct {
	db *gorm.DB
}

func (model *UserModel) GetOrRegisterGoogle(username string, email string, googleID string, pictureURL string) (*User, error) {
	user, err := model.GetUserByEmail(email)

	if err == nil {
		if user.IsOauth() && user.ProviderID == googleID {
			return user, nil
		}

		return nil, fmt.Errorf("user with email %s already registered manually", email)
	}

	if !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to retrieve user with email %s: %v", email, err)
	}

	user, err = model.CreateUserGoogle(username, email, googleID, pictureURL)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (model *UserModel) CreateUserGoogle(username string, email string, googleID string, pictureURL string) (*User, error) {
	user := &User{
		Username:   username,
		Email:      email,
		Privileage: Unprivileaged,
		Provider:   "google",
		ProviderID: googleID,
		PictureURL: pictureURL,
	}

	user, err := model.CreateUser(user)
	return user, err
}

func (model *UserModel) CreateUserRegistration(username string, hashedPassword string) (*User, error) {
	user := &User{
		Username:       username,
		HashedPassword: hashedPassword,
		Privileage:     Unprivileaged,
	}

	user, err := model.CreateUser(user)
	return user, err
}

func (model *UserModel) CreateUser(user *User) (*User, error) {
	if user.Username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	res, err := model.UsernameExists(user.Username)
	if err != nil {
		return nil, err
	}

	if res {
		return nil, fmt.Errorf("user with username %s already exists", user.Username)
	}

	result := model.db.Create(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %v", result.Error)
	}

	return user, nil
}

func (model *UserModel) UsernameExists(username string) (bool, error) {
	var user User

	err := model.db.Model(&User{}).Where("username = ?", username).First(&user).Error
	if err == nil {
		return true, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("failed to check for existing user: %v", err)
	}

	return false, nil
}

func (model *UserModel) EmailExists(email string) (bool, error) {
	var user User

	err := model.db.Where("email = ?", email).First(&user).Error
	if err == nil {
		return true, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("failed to check for existing user: %v", err)
	}

	return false, nil
}

func (model *UserModel) GetUserById(id float64) (*User, error) {
	var user User
	result := model.db.Model(&User{}).Where("id = ?", id).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user with id %f not found", id)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve user with id %f: %v", id, result.Error)
	}

	return &user, nil
}

func (model *UserModel) GetUserByUsername(name string) (*User, error) {
	var user User
	result := model.db.Where("username = ?", name).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user with username %s not found", name)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve user with username %s: %v", name, result.Error)
	}

	return &user, nil
}

func (model *UserModel) GetUserByEmail(email string) (*User, error) {
	var user User
	err := model.db.Where("email = ?", email).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user with email %s not found", email)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user with email %s: %v", email, err)
	}

	return &user, nil
}

func (user *User) IsOauth() bool {
	return user.Provider != "" && user.ProviderID != ""
}
