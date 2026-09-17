package board

type BoardCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BoardPatchRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
