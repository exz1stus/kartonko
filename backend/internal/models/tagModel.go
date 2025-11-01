package models

import (
	"bufio"
	"fmt"
	"os"

	"gorm.io/gorm"
)

type Tag struct {
	Name string `json:"name" gorm:"primaryKey;unique;not null"`
}

type TagModel struct {
	db *gorm.DB
}

func (model *TagModel) RegisterAutoTags() error {
	autoTagsFile, err := os.Open("autotags.txt")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(autoTagsFile)
	for scanner.Scan() {
		tagName := scanner.Text()
		if tagName != "" {
			model.AddTag(tagName)
		}
	}

	defer autoTagsFile.Close()
	return nil
}

func (model *TagModel) SearchTags(query string, limit int) ([]Tag, error) {
	var tags []Tag
	result := model.db.Model(&Tag{}).Where("name LIKE ?", query+"%").Limit(limit).Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

func ConstructTagsByNames(names []string) []Tag {
	var tags []Tag
	for _, name := range names {
		tags = append(tags, Tag{Name: name})
	}

	return tags
}

func (model *TagModel) AddTag(tag string) error {
	if model.TagExists(tag) {
		return fmt.Errorf("tag %s already exists", tag)
	}

	newTag := &Tag{Name: tag}
	result := model.db.Create(newTag)
	if result.Error != nil {
		return fmt.Errorf("failed to insert tag: %v", result.Error)
	}
	return nil
}

func (model *TagModel) CheckTags(tags []Tag) error {
	if len(tags) == 0 || len(tags) == 1 && tags[0].Name == "" {
		return nil
	}

	tagNames := make([]string, len(tags))
	tagNameSet := make(map[string]struct{}, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
		tagNameSet[tag.Name] = struct{}{}
	}

	var retrievedTags []Tag
	result := model.db.Model(&Tag{}).Where("name IN ?", tagNames).Find(&retrievedTags)
	if result.Error != nil {
		return fmt.Errorf("failed to check tags in database: %v", result.Error)
	}

	for _, tag := range retrievedTags {
		delete(tagNameSet, tag.Name)
	}

	if len(tagNameSet) > 0 {
		missing := make([]string, 0, len(tagNameSet))
		for name := range tagNameSet {
			missing = append(missing, name)
		}
		return fmt.Errorf("tags do not exist: %v", missing)
	}

	return nil
}

func (model *TagModel) TagExists(name string) bool {
	var count int64
	model.db.Model(&Tag{}).Where("name = ?", name).Count(&count)
	return count > 0
}
