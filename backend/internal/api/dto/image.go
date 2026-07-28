package dto

import "server/pkg/image"

type ImagePostRequest struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
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
}

type ImagePostBatchRequest struct {
	Data       []ImagePostRequest `json:"data"`
	CommonTags []string           `json:"common_tags"`
}

type ImagePostBatchResponse struct {
	Successes []ImageResponse    `json:"successes"`
	Failures  []image.ImageError `json:"failures"`
}
