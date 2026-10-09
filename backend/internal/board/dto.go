package board

type BoardCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
} // @name BoardPostRequest

type BoardPatchRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
} // @name BoardPatchRequest
