package board

type BoardCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
} // @name BoardPostRequest

type BoardPatchRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
} // @name BoardPatchRequest
