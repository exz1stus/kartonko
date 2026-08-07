package services

import (
	"context"
	"fmt"
	"server/internal/models"
	"server/internal/repositories"

	"gorm.io/gorm"
)

type TagService interface {
	Create(ctx context.Context, tag string, user *models.User) (*models.Tag, error)

	SearchPrefix(prefix string, limit int) ([]models.Tag, error)

	Exists(name string) (bool, error)
	ExistMany(tags []models.Tag) ([]bool, error)
}

type tagService struct {
	tags repositories.TagRepository
	logs LogService
	db   *gorm.DB
}

func NewTagService(db *gorm.DB, tags repositories.TagRepository, logs LogService) TagService {
	return &tagService{tags, logs, db}
}

func (s *tagService) Create(ctx context.Context, tag string, user *models.User) (*models.Tag, error) {
	if user == nil {
		return nil, fmt.Errorf("creating tag: received nil user")
	}
	exists, err := s.tags.Exists(tag)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("tag already exists: %s", tag)
	}

	var createdTag *models.Tag

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tagsRepoTX := s.tags.WithTx(tx)
		tag, err := tagsRepoTX.Create(tag)
		if err != nil {
			return err
		}

		if err := s.logs.Log(tx, "create", "tag", user.ID, tag.ID, nil); err != nil {
			return err
		}

		createdTag = tag
		return nil
	})

	return createdTag, nil
}

func (s *tagService) SearchPrefix(prefix string, limit int) ([]models.Tag, error) {
	return s.tags.SearchPrefix(prefix, limit)
}

func (s *tagService) Exists(name string) (bool, error) {
	return s.tags.Exists(name)
}

func (s *tagService) ExistMany(tags []models.Tag) ([]bool, error) {
	return s.tags.ExistMany(tags)
}
