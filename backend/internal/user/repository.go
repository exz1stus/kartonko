package user

import (
	"context"
	stderrors "errors"
	apperrors "server/internal/errors"

	"gorm.io/gorm"
)

type UserRepository interface {
	WithTx(tx *gorm.DB) UserRepository

	Create(ctx context.Context, user *User) error

	Get(ctx context.Context, id uint) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByProviderID(ctx context.Context, providerID string) (*User, error)

	Update(ctx context.Context, user *User) error

	Delete(ctx context.Context, userID uint) error

	SetPrivilege(ctx context.Context, userID uint, privilege Privilege) error

	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
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

func (r *userRepository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Get(ctx context.Context, id uint) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByProviderID(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("provider_id = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) Delete(ctx context.Context, userID uint) error {
	return r.db.
		WithContext(ctx).
		Where("id = ?", userID).
		Delete(&User{}).
		Error
}

func (r *userRepository) Update(ctx context.Context, user *User) error {
	result := r.db.
		WithContext(ctx).
		Updates(user)

	if result.Error != nil {
		if stderrors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return apperrors.ErrDuplicateName
		}
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *userRepository) SetPrivilege(ctx context.Context, userID uint, privilege Privilege) error {
	return r.db.WithContext(ctx).Model(&User{}).
		Where("id = ?", userID).
		Update("privilege", privilege).Error
}

func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}
