package auth

import "server/internal/models"

func CanEdit(user *models.User, img *models.ImageMetadata) bool {
	return user.Privilege == models.Moderator ||
		user.ID == img.UserID
}
