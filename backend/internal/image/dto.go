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
