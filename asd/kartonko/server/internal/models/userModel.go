package models

import (
	"errors"
	"fmt"

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
	HashedPassword string     `json:"hashedPassword" gorm:"not null"`
	Privileage     Privileage `json:"privileage" gorm:"not null"`
}

type UserModel struct {
	db *gorm.DB
}

func (users *UserModel) CreateUser(username string, hashedPassword string) (*User, error) {
	var existingUser User

	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	err := users.db.Where("username = ?", username).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("username '%s' is already taken", username)
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check for existing user: %v", err)
	}

	user := &User{
		Username:       username,
		HashedPassword: hashedPassword,
		Privileage:     Unprivileaged,
	}

	result := users.db.Create(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %v", result.Error)
	}
	return user, nil
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
	result := model.db.Model(&User{}).Where("username = ?", name).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user with username %s not found", name)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve user with username %s: %v", name, result.Error)
	}
	return &user, nil
}
