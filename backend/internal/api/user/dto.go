package user

import (
	"fmt"
	"time"

	"server/internal/user"
)

type UserDataResponse struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Privilege  string `json:"privilege"`
	PictureURL string `json:"picture_url"`
	JoinedAt   string `json:"joined_at"`
	LastSeen   string `json:"last_seen"`
	Online     bool   `json:"online"`
} // @name UserDataResponse

func NewUserDataResponse(u *user.User) UserDataResponse {
	res := UserDataResponse{
		ID:        u.ID,
		Username:  u.Username,
		Privilege: u.Privilege.String(),
		JoinedAt:  u.CreatedAt.Format(time.DateOnly),
		LastSeen:  u.LastSeen.Format(time.DateTime),
	}

	res.PictureURL = u.PictureURL
	if u.PictureURL == profilePictureURL(u.ID) {
		res.PictureURL = fmt.Sprintf("%s?v=%d", u.PictureURL, u.UpdatedAt.UnixNano())
	}

	return res
}
