package tag

import "server/internal/tag"

type TagPostRequest struct {
	Name string `json:"name"`
} // @name TagPostRequest

type TagPostBatchRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
} // @name TagPostBatchRequest

type TagResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
} // @name TagResponse

type TagBatchResponse struct {
	Successes []TagResponse `json:"successes"`
	Failures  []struct {
		Name  string `json:"name"`
		Error string `json:"error"`
	} `json:"failures,omitempty"`
} // @name TagBatchResponse

func (req TagPostRequest) ToServiceCreate() string {
	return req.Name
}

func (req TagPostBatchRequest) ToServiceCreate() []string {
	return req.Names
}

func FromServiceTag(t *tag.Tag) TagResponse {
	return TagResponse{ID: t.ID, Name: t.Name}
}

func FromServiceTagValue(t tag.Tag) TagResponse {
	return TagResponse{ID: t.ID, Name: t.Name}
}

func FromServiceTags(tags []tag.Tag) []TagResponse {
	resp := make([]TagResponse, len(tags))
	for i, t := range tags {
		resp[i] = FromServiceTagValue(t)
	}
	return resp
}