package tag

import (
	"context"
	"fmt"
	"server/internal/api/transaction"
	"server/internal/log"

	"gorm.io/gorm"
)

type TagService interface {
	Create(ctx context.Context, tag string, userID uint) (*Tag, error)

	SearchPrefix(prefix string, cursor int, limit int) ([]Tag, error)

	Exists(name string) (bool, error)
	ExistMany(tags []Tag) ([]bool, error)
}

type tagService struct {
	tags         TagRepository
	logs         log.LogService
	transactions transaction.Runner
}

func NewTagService(tags TagRepository, logs log.LogService, transactions transaction.Runner) TagService {
	return &tagService{tags, logs, transactions}
}

func (s *tagService) Create(ctx context.Context, tag string, userID uint) (*Tag, error) {
	exists, err := s.tags.Exists(tag)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("tag already exists: %s", tag)
	}

	var createdTag *Tag

	err = s.transactions.Within(ctx, func(tx *gorm.DB) error {
		tagsRepoTX := s.tags.WithTx(tx)
		tag, err := tagsRepoTX.Create(tag)
		if err != nil {
			return err
		}

		if err := s.logs.Log(tx, "create", "tag", userID, tag.ID, nil); err != nil {
			return err
		}

		createdTag = tag
		return nil
	})

	return createdTag, nil
}

func (s *tagService) SearchPrefix(prefix string, cursor int, limit int) ([]Tag, error) {
	return s.tags.SearchPrefix(prefix, cursor, limit)
}

func (s *tagService) Exists(name string) (bool, error) {
	return s.tags.Exists(name)
}

func (s *tagService) ExistMany(tags []Tag) ([]bool, error) {
	return s.tags.ExistMany(tags)
}
