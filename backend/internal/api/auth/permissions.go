package auth

import (
	"server/internal/user"
)

func CanEdit(userID uint, privilege user.Privilege, ownerID uint) bool {
	return privilege == user.Moderator ||
		userID == ownerID
}
