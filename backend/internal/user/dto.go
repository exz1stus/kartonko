package user

import "time"

func NewUserData(u *User) UserResponce {
	res := UserResponce{
		ID:        u.ID,
		Username:  u.Username,
		Privilege: u.Privilege.String(),
		JoinedAt:  u.CreatedAt.Format(time.DateOnly),
		LastSeen:  u.LastSeen.Format(time.DateTime),
	}

	if u.IsOauth() {
		res.PictureURL = u.PictureURL
	}

	return res
}

type UserResponce struct {
	ID         uint
	Username   string
	Privilege  string
	PictureURL string
	JoinedAt   string
	LastSeen   string
	Online     bool
}
