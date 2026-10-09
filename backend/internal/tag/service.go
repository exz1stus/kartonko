package tag

import (
	"context"
	"fmt"
	"server/internal/api/transaction"
	"server/internal/errors"
	"server/internal/log"
	userpkg "server/internal/user"
	"strings"

	"gorm.io/gorm"
)

type TagService interface {
	Create(ctx context.Context, req *TagPostRequest, userID uint) (*Tag, error)
	Get(ctx context.Context, tagID uint) (*Tag, error)
	Update(ctx context.Context, userID, tagID uint, req *TagPatchRequest) (*Tag, error)
	Delete(ctx context.Context, userID, tagID uint) error

	SearchPrefix(ctx context.Context, prefix string, cursor int, limit int) ([]Tag, error)

	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistMany(ctx context.Context, tags []Tag) ([]bool, error)
}

type tagService struct {
	tags         TagRepository
	users        userpkg.UserRepository
	logs         log.LogService
	transactions transaction.Runner
}

func NewTagService(tags TagRepository, logs log.LogService, users userpkg.UserRepository, transactions transaction.Runner) TagService {
	return &tagService{tags, users, logs, transactions}
}

func (s *tagService) checkUserPermission(ctx context.Context, tagID uint, userID uint) error {
	user, err := s.users.Get(ctx, userID)
	if err != nil {
		return err
	}

	tag, err := s.tags.Get(ctx, tagID)
	if err != nil {
		return err
	}

	if !userpkg.CanEdit(user.ID, user.Privilege, tag.UserID) {
		return errors.ErrPermissionDenied
	}

	return nil
}

func (s *tagService) Create(ctx context.Context, req *TagPostRequest, userID uint) (*Tag, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, errors.ErrBadRequest
	}
	exists, err := s.tags.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("tag already exists: %s", req.Name)
	}

	var createdTag *Tag

	err = s.transactions.Within(ctx, func(tx *gorm.DB) error {
		tagsRepoTX := s.tags.WithTx(tx)
		tag := &Tag{Name: req.Name, UserID: userID}

		if err := tagsRepoTX.Create(ctx, tag); err != nil {
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

func (s *tagService) Get(ctx context.Context, tagID uint) (*Tag, error) {
	return s.tags.Get(ctx, tagID)
}

func (s *tagService) Delete(ctx context.Context, userID, tagID uint) error {
	if err := s.checkUserPermission(ctx, tagID, userID); err != nil {
		return err
	}

	return s.tags.Delete(ctx, tagID)
}

func (s *tagService) Update(ctx context.Context, userID, tagID uint, req *TagPatchRequest) (*Tag, error) {
	if err := s.checkUserPermission(ctx, tagID, userID); err != nil {
		return nil, err
	}

	board, err := s.Get(ctx, tagID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errors.ErrBadRequest
		}
		if name != board.Name {
			exists, err := s.tags.ExistsByName(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("check duplicate: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("%w: %s", errors.ErrDuplicateName, name)
			}
		}
		board.Name = name
	}

	if err := s.tags.Update(ctx, board); err != nil {
		return nil, err
	}

	return board, nil
}

func (s *tagService) SearchPrefix(ctx context.Context, prefix string, cursor int, limit int) ([]Tag, error) {
	return s.tags.SearchPrefix(ctx, prefix, cursor, limit)
}

func (s *tagService) ExistsByName(ctx context.Context, name string) (bool, error) {
	return s.tags.ExistsByName(ctx, name)
}

func (s *tagService) ExistMany(ctx context.Context, tags []Tag) ([]bool, error) {
	return s.tags.ExistMany(ctx, tags)
}
