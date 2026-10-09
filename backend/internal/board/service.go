package board

import (
	"context"
	"server/internal/errors"
	userpkg "server/internal/user"
	"strings"
)

type BoardService interface {
	Create(ctx context.Context, userID uint, req *BoardCreateRequest) (*Board, error)
	Get(ctx context.Context, boardID uint) (*Board, error)
	Update(ctx context.Context, userID, boardID uint, req *BoardPatchRequest) (*Board, error)
	Delete(ctx context.Context, userID, boardID uint) error

	List(ctx context.Context, cursor, limit int) ([]Board, error)

	AddImage(ctx context.Context, boardID, imageID, userID uint) (*BoardItem, error)
	GetImage(ctx context.Context, boardID, imageID uint) (*BoardItem, error)
	RemoveImage(ctx context.Context, boardID, imageID, userID uint) error

	ListImages(ctx context.Context, boardID uint, cursor, limit int) ([]BoardItem, error)
}

type boardService struct {
	boards BoardRepository
	users  userpkg.UserRepository
}

func NewBoardService(boards BoardRepository, users userpkg.UserRepository) BoardService {
	return &boardService{boards, users}
}

func (s *boardService) List(ctx context.Context, cursor, limit int) ([]Board, error) {
	return s.boards.List(ctx, cursor, limit)
}

func (s *boardService) Create(ctx context.Context, userID uint, req *BoardCreateRequest) (*Board, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, errors.ErrBadRequest
	}
	board := &Board{
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
	}

	if err := s.boards.Create(ctx, board); err != nil {
		return nil, err
	}

	return board, nil
}

func (s *boardService) Delete(ctx context.Context, userID uint, boardID uint) error {
	if err := s.checkUserPermission(ctx, boardID, userID); err != nil {
		return err
	}

	return s.boards.Delete(ctx, boardID)
}

func (s *boardService) Get(ctx context.Context, boardID uint) (*Board, error) {
	return s.boards.Get(ctx, boardID)
}

func (s *boardService) Update(ctx context.Context, userID uint, boardID uint, req *BoardPatchRequest) (*Board, error) {
	if err := s.checkUserPermission(ctx, boardID, userID); err != nil {
		return nil, err
	}

	board, err := s.Get(ctx, boardID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, errors.ErrBadRequest
		}
		board.Name = name
	}

	if req.Description != nil {
		board.Description = *req.Description
	}

	if err := s.boards.Update(ctx, board); err != nil {
		return nil, err
	}

	return board, nil
}

func (s *boardService) checkUserPermission(ctx context.Context, boardID uint, userID uint) error {
	user, err := s.users.Get(ctx, userID)
	if err != nil {
		return err
	}

	board, err := s.boards.Get(ctx, boardID)
	if err != nil {
		return err
	}

	if !userpkg.CanEdit(user.ID, user.Privilege, board.UserID) {
		return errors.ErrPermissionDenied
	}

	return nil
}

func (s *boardService) GetImage(ctx context.Context, boardID, imageID uint) (*BoardItem, error) {
	return s.boards.GetItem(ctx, boardID, imageID)
}

func (s *boardService) ListImages(ctx context.Context, boardID uint, cursor, limit int) ([]BoardItem, error) {
	return s.boards.ListItems(ctx, boardID, cursor, limit)
}

func (s *boardService) AddImage(ctx context.Context, boardID uint, imageID uint, userID uint) (*BoardItem, error) {
	if err := s.checkUserPermission(ctx, boardID, userID); err != nil {
		return nil, err
	}

	item := &BoardItem{
		BoardID: boardID,
		ImageID: imageID,
	}

	if err := s.boards.AddItem(ctx, boardID, item); err != nil {
		return nil, err
	}

	return s.boards.GetItem(ctx, boardID, imageID)
}

func (s *boardService) RemoveImage(ctx context.Context, boardID uint, imageID uint, userID uint) error {
	if err := s.checkUserPermission(ctx, boardID, userID); err != nil {
		return err
	}

	return s.boards.DeleteItem(ctx, boardID, imageID)
}
