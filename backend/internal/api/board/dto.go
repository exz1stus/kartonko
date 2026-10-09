package board

import (
	"server/internal/api/image"
	"server/internal/board"
	"time"
)

type BoardResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	UserID      uint   `json:"user_id"`
	ImageCount  int64  `json:"image_count"`
} // @name BoardResponse

type BoardItemResponse struct {
	ID    uint                `json:"id"`
	Added time.Time           `json:"added"`
	Image image.ImageResponse `json:"image_metadata"`
} // @name BoardItemResponse

type BoardItemPostRequest struct {
	ImageID uint64 `json:"image_id" binding:"required"`
} // @name BoardItemPostRequest

func NewBoardItemResponse(item *board.BoardItem) BoardItemResponse {
	return BoardItemResponse{
		ID:    item.ID,
		Added: item.CreatedAt,
		Image: image.NewImageResponse(&item.Image),
	}
}

func NewBoardResponse(board *board.Board) BoardResponse {
	return BoardResponse{
		ID:          board.ID,
		Name:        board.Name,
		Slug:        board.Slug,
		Description: board.Description,
		UserID:      board.UserID,
		ImageCount:  board.ImageCount,
	}
}
