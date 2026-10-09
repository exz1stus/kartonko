package user

func CanEdit(userID uint, privilege Privilege, ownerID uint) bool {
	return privilege == Moderator ||
		userID == ownerID
}
