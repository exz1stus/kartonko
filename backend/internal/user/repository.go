package user

import (
	"fmt"

	"gorm.io/gorm"
)

type UserRepository interface {
	WithTx(tx *gorm.DB) UserRepository

	Create(user *User) error

	GetByID(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByProviderID(id string) (*User, error)

	SetPrivilege(userID uint, privilege Privilege) error

	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) WithTx(tx *gorm.DB) UserRepository {
	return NewUserRepository(tx)
}

func (r *userRepository) Create(user *User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(id uint) (*User, error) {
	var user User
	err := r.db.Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByUsername(username string) (*User, error) {
	var user User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByProviderID(id string) (*User, error) {
	var user User
	err := r.db.Where("provider_id = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) SetPrivilege(userID uint, privilege Privilege) error {
	return r.db.Model(&User{}).
		Where("id = ?", userID).
		Update("privilege", privilege).Error
}

func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
