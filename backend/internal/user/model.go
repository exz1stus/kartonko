package user

import (
	"time"

	"gorm.io/gorm"
)

type Privilege int

const (
	Unprivileged Privilege = iota
	Moderator
)

func (p Privilege) String() string {
	return [...]string{"Unprivileged", "Moderator"}[p]
}

type User struct {
	gorm.Model
	Username       string    `json:"username" gorm:"unique;not null"`
	Email          *string   `json:"email" gorm:"unique"`
	HashedPassword string    `json:"hashed_password"`
	Privilege      Privilege `json:"privilege" gorm:"not null"`
	Provider       string    `gorm:"uniqueIndex:idx_provider_provider_id"`
	ProviderID     *string   `gorm:"uniqueIndex:idx_provider_provider_id"`
	PictureURL     string    `json:"picture_url"`
	LastSeen       time.Time `json:"last_seen"`
}

func (user *User) IsOauth() bool {
	return user.Provider != "" && user.ProviderID != nil
}

func (user *User) IsModerator() bool {
	return user.Privilege == Moderator
}
