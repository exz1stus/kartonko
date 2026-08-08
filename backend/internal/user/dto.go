package user

type UserDataResponse struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Privilege  string `json:"privilege"`
	PictureURL string `json:"picture_url"`
	JoinedAt   string `json:"joined_at"`
	LastSeen   string `json:"last_seen"`
	Online     bool   `json:"online"`
}
