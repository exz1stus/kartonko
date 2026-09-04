package image

// swagger:model
type ImagePostRequest struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// swagger:model
type ImageResponse struct {
	ID       uint     `json:"id"`
	Hash     string   `json:"hash"`
	Filename string   `json:"filename"`
	Tags     []string `json:"tags"`
	Format   string   `json:"format"`
	Width    uint     `json:"width"`
	Height   uint     `json:"height"`
	UserID   uint     `json:"user_id"`
	Uploaded string   `json:"uploaded_at"`
}

// swagger:model
type ImagePostBatchRequest struct {
	Data       []ImagePostRequest `json:"data"`
	CommonTags []string           `json:"common_tags"`
}

// swagger:model
type ImageError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

// swagger:model
type ImagePostBatchResponse struct {
	Successes []ImageResponse `json:"successes"`
	Failures  []ImageError    `json:"failures"`
}
