package user

import "time"

func NewUserResponse(u *User) UserDataResponse {
	res := UserDataResponse{
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

type UserDataResponse struct {
	ID         uint
	Username   string
	Privilege  string
	PictureURL string
	JoinedAt   string
	LastSeen   string
	Online     bool
} // @name UserInternalResponse