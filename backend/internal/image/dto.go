package image

type UploadRequest struct {
	Name string
	Tags []string
}

type UploadBatchRequest struct {
	Data       []UploadRequest
	CommonTags []string
}

type ImageError struct {
	Name  string
	Error string
}

type ImagePatchRequest struct {
	Filename *string  `json:"filename"`
	Tags     []string `json:"tags"`
} // @name ImagePatchRequest
