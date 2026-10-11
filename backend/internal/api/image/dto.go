package image

import "server/internal/image"

type ImagePostRequest struct {
	Name     string   `json:"name"`
	Tags     []string `json:"tags"`
	ClientID string   `json:"client_id,omitempty"`
} // @name ImagePostRequest

func (req ImagePostRequest) ToUploadRequest() image.UploadRequest {
	return image.UploadRequest{
		Name: req.Name,
		Tags: req.Tags,
	}
}

type ImageResponse struct {
	ID              uint     `json:"id"`
	Hash            string   `json:"hash"`
	Filename        string   `json:"filename"`
	Tags            []string `json:"tags"`
	Format          string   `json:"format"`
	Width           uint     `json:"width"`
	Height          uint     `json:"height"`
	UserID          uint     `json:"user_id"`
	Uploaded        string   `json:"uploaded_at"`
	EmbeddingStatus string   `json:"embedding_status"`
} // @name ImageMetadata

type ImagePostBatchRequest struct {
	Data       []ImagePostRequest `json:"data"`
	CommonTags []string           `json:"common_tags"`
} // @name ImagePostBatchRequest

type ImageError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
} // @name ImageError

type ImagePostBatchResponse struct {
	Successes []ImageResponse        `json:"successes"`
	Failures  []ImageError           `json:"failures"`
	Results   []ImagePostBatchResult `json:"results"`
} // @name ImagePostBatchResponse

type ImagePostBatchResult struct {
	ClientID string         `json:"client_id"`
	Image    *ImageResponse `json:"image,omitempty"`
	Error    string         `json:"error,omitempty"`
} // @name ImagePostBatchResult
