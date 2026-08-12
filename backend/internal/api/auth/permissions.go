package auth

import (
	"server/internal/image"
	userpkg "server/internal/user"
)

func CanEdit(user *userpkg.User, img *image.ImageMetadata) bool {
	return user.Privilege == userpkg.Moderator ||
		user.ID == img.UserID
}
