package board

import (
	"server/internal/api/image"
	"server/internal/board"
	"time"
)

type BoardResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
} // @name BoardResponse

type BoardItemResponse struct {
	Added time.Time           `json:"added"`
	Image image.ImageResponse `json:"image_metadata"`
} // @name BoardItemResponse

type BoardItemPostRequest struct {
	ImageID uint64 `json:"image_id" binding:"required"`
} // @name BoardItemPostRequest

type BoardItemPatchRequest struct {
	//TODO: add request when image have some additional data
} // @name BoardItemPatchRequest

func NewBoardItemResponse(item *board.BoardItem) BoardItemResponse {
	return BoardItemResponse{
		Added: item.CreatedAt,
		Image: image.NewImageResponse(&item.Image),
	}
}

func NewBoardResponse(board *board.Board) BoardResponse {
	return BoardResponse{
		Name:        board.Name,
		Description: board.Description,
	}
}
