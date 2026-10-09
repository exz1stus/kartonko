package user

import "time"

type UserPatchRequest struct {
	Username   *string `json:"username,omitempty"`
	PictureURL *string `json:"picture_url,omitempty"`
}

func NewUserData(u *User) UserResponce {
	res := UserResponce{
		ID:        u.ID,
		Username:  u.Username,
		Privilege: u.Privilege.String(),
		JoinedAt:  u.CreatedAt.Format(time.DateOnly),
		LastSeen:  u.LastSeen.Format(time.DateTime),
	}

	res.PictureURL = u.PictureURL

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
