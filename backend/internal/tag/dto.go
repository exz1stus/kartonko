package tag

import (
	"encoding/json"
	"fmt"
)

type TagPostRequest struct {
	Name string `json:"name" binding:"required"`
} // @name TagPostRequest

type TagPostBatchRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
} // @name TagPostBatchRequest

type TagPatchRequest struct {
	Name *string `json:"name,omitempty"`
} // @name TagPatchRequest

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

func NewTagResponse(tag *Tag) TagResponse {
	return TagResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}
}

func ParseTagsFromJSONString(tagsString string) ([]string, error) {
	var tags []string

	err := json.Unmarshal([]byte(tagsString), &tags)
	if err != nil {
		return nil, fmt.Errorf("Failed parsing tags from json: %w", err)
	}

	return tags, nil
}
