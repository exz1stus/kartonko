package tag

type TagResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type PostTagsBatchRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
}

type PostTagRequest struct {
	Name string `json:"name"`
}
