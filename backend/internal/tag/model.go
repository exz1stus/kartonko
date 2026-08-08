package tag

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type Tag struct {
	gorm.Model
	Name string `json:"name" gorm:"unique;not null"`
}

func ConstructTagsByNames(names []string) []Tag {
	var tags []Tag
	for _, name := range names {
		tags = append(tags, Tag{Name: name})
	}

	return tags
}

func TagsToStrings(tags []Tag) []string {
	var names []string
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return names
}

func parseTagsFromJSONString(tagsString string) ([]string, error) {
	var tags []string

	err := json.Unmarshal([]byte(tagsString), &tags)
	if err != nil {
		return nil, fmt.Errorf("Failed parsing tags from json: %w", err)
	}

	return tags, nil
}
