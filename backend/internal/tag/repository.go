package tag

import (
	"fmt"

	"gorm.io/gorm"
)

type TagRepository interface {
	WithTx(tx *gorm.DB) TagRepository

	Create(tag string) (*Tag, error)

	SearchPrefix(prefix string, cursor int, limit int) ([]Tag, error)

	Exists(name string) (bool, error)
	ExistMany(tags []Tag) ([]bool, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) WithTx(tx *gorm.DB) TagRepository {
	return NewTagRepository(tx)
}

func (r *tagRepository) Create(tag string) (*Tag, error) {
	newTag := &Tag{Name: tag}
	if err := r.db.Create(newTag).Error; err != nil {
		return nil, fmt.Errorf("failed to insert tag: %w", err)
	}

	return newTag, nil
}

func (r *tagRepository) SearchPrefix(prefix string, cursor int, limit int) ([]Tag, error) {
	var tags []Tag
	if err := r.db.Model(&Tag{}).
		Where("name LIKE ?", prefix+"%").
		Offset(cursor).
		Limit(limit).
		Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

func (r *tagRepository) Exists(name string) (bool, error) {
	var count int64
	err := r.db.Model(&Tag{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *tagRepository) ExistMany(tags []Tag) ([]bool, error) {
	tagNames := make([]string, len(tags))

	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	var retrievedTags []Tag
	if err := r.db.
		Where("name IN ?", tagNames).
		Find(&retrievedTags).
		Error; err != nil {
		return nil, fmt.Errorf("failed to check tags in database: %w", err)
	}

	exists := make(map[string]bool, len(retrievedTags))
	for _, tag := range retrievedTags {
		exists[tag.Name] = true
	}

	result := make([]bool, len(tags))
	for i, tag := range tags {
		result[i] = exists[tag.Name]
	}

	return result, nil
}
