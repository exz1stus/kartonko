package tag

import (
	"encoding/json"
	"fmt"
)

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
