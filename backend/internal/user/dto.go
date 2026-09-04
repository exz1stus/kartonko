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

// swagger:model
type UserDataResponse struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Privilege  string `json:"privilege"`
	PictureURL string `json:"picture_url"`
	JoinedAt   string `json:"joined_at"`
	LastSeen   string `json:"last_seen"`
	Online     bool   `json:"online"`
}
