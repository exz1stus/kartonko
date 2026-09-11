package image

import "server/internal/image"

type ImagePostRequest struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
} // @name ImagePostRequest

func (req ImagePostRequest) ToUploadRequest() image.UploadRequest {
	return image.UploadRequest{
		Name: req.Name,
		Tags: req.Tags,
	}
}

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
	Successes []ImageResponse `json:"successes"`
	Failures  []ImageError    `json:"failures"`
} // @name ImagePostBatchResponse
