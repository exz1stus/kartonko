package board

import (
	"context"
	"server/internal/api/auth"
	"server/internal/errors"
	"server/internal/user"
)

type BoardService interface {
	Create(ctx context.Context, userID uint, req *BoardCreateRequest) (*Board, error)
	Get(ctx context.Context, boardID uint) (*Board, error)
	Update(ctx context.Context, userID, boardID uint, req *BoardPatchRequest) (*Board, error)
	Delete(ctx context.Context, userID, boardID uint) error

	List(ctx context.Context, cursor, limit int) ([]Board, error)

	AddImage(ctx context.Context, boardID, imageID, userID uint) (*BoardItem, error)
	GetImage(ctx context.Context, boardID, imageID uint) (*BoardItem, error)
	// UpdateImage(ctx context.Context, boardID, ..., userID uint) (*BoardItem, error)
	RemoveImage(ctx context.Context, boardID, imageID, userID uint) error

	ListImages(ctx context.Context, boardID uint, cursor, limit int) ([]BoardItem, error)
}

type boardService struct {
	boards BoardRepository
	users  user.UserRepository
}

func NewBoardService(boards BoardRepository, users user.UserRepository) BoardService {
	return &boardService{boards, users}
}

func (s *boardService) List(ctx context.Context, cursor, limit int) ([]Board, error) {
	return s.boards.List(cursor, limit)
}

func (s *boardService) Create(ctx context.Context, userID uint, req *BoardCreateRequest) (*Board, error) {
	board := &Board{
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
	}

	if err := s.boards.Create(board); err != nil {
		return nil, err
	}

	return board, nil
}

func (s *boardService) Delete(ctx context.Context, userID uint, boardID uint) error {
	if err := s.checkUserPermission(ctx, boardID, userID); err != nil {
		return err
	}

	return s.boards.Delete(boardID)
}

func (s *boardService) Get(ctx context.Context, boardID uint) (*Board, error) {
	return s.boards.Get(boardID)
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
		board.Name = *req.Name
	}

	if req.Description != nil {
		board.Description = *req.Description
	}

	if err := s.boards.Update(board); err != nil {
		return nil, err
	}

	return board, nil
}

func (s *boardService) checkUserPermission(ctx context.Context, boardID uint, userID uint) error {
	user, err := s.users.GetByID(userID)
	if err != nil {
		return err
	}

	board, err := s.boards.Get(boardID)
	if err != nil {
		return err
	}

	if !auth.CanEdit(user.ID, user.Privilege, board.UserID) {
		return errors.ErrPermissionDenied
	}

	return nil
}

func (s *boardService) GetImage(ctx context.Context, boardID, imageID uint) (*BoardItem, error) {
	return s.boards.GetItem(boardID, imageID)
}

func (s *boardService) ListImages(ctx context.Context, boardID uint, cursor, limit int) ([]BoardItem, error) {
	return s.boards.ListItems(boardID, cursor, limit)
}

func (s *boardService) AddImage(ctx context.Context, boardID uint, imageID uint, userID uint) (*BoardItem, error) {
	if err := s.checkUserPermission(ctx, boardID, userID); err != nil {
		return nil, errors.WrapPermissionDenied(err)
	}

	item := &BoardItem{
		BoardID: boardID,
		ImageID: imageID,
	}

	if err := s.boards.AddItem(boardID, item); err != nil {
		return nil, err
	}

	return item, nil
}

func (s *boardService) RemoveImage(ctx context.Context, boardID uint, imageID uint, userID uint) error {
	if err := s.checkUserPermission(ctx, boardID, userID); err != nil {
		return err
	}

	return s.boards.DeleteItem(boardID, imageID)
}

// func (s *boardService) UpdateImage(ctx context.Context, boardID uint, item *BoardItemPatchRequest, userID uint) (*BoardItem, error) {
// 	if err := s.checkUserPermission(ctx, boardID, userID); err != nil {
// 		return nil, err
// 	}

// 	//TODO

// 	if err := s.boards.UpdateItem(boardID, item); err != nil {
// 		return nil, err
// 	}

// 	return item, nil
// }
