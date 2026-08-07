package dto

type UserDataResponse struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Privilege  string `json:"privilege"`
	PictureURL string `json:"picture_url"`
	JoinedAt   string `json:"joined_at"`
	LastSeen   string `json:"last_seen"`
	Online     bool   `json:"online"`
}

type PostTagsBatchRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
}

type PostTagRequest struct {
	Name string `json:"name"`
}
