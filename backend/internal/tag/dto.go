package tag

// swagger:model
type TagPostBatchRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
}

// swagger:model
type TagPostRequest struct {
	Name string `json:"name"`
}

// swagger:model
type TagResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// swagger:model
type TagBatchResponse struct {
	Successes []TagResponse `json:"successes"`
	Failures  []struct {
		Name  string `json:"name"`
		Error string `json:"error"`
	} `json:"failures,omitempty"`
}
