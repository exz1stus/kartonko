package board

import (
	"net/http"
	"server/internal/api/helpers"
	"server/internal/board"
	"server/internal/errors"
	"server/internal/user"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	boardService board.BoardService
}

func NewBoardHandler(
	boardService board.BoardService,
) *Handler {
	return &Handler{
		boardService: boardService,
	}
}

func (h *Handler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.GET("", h.ListBoards)

	protected.POST("", h.PostBoard)
	public.GET("/:id", h.GetBoard)
	protected.PATCH("/:id", h.PatchBoard)
	protected.DELETE("/:id", h.DeleteBoard)
	public.GET("/:id/image", h.ListBoardImages)

	protected.POST("/:id/image", h.PostBoardImage)
	protected.GET("/:id/image/:imageId", h.GetBoardImage)
	// protected.PATCH("/:id/image/:imageId", h.PatchBoardImage)
	protected.DELETE("/:id/image/:imageId", h.DeleteBoardImage)
}

// ListBoards godoc
// @Summary List boards
// @Description Returns a slice of list of boards
// @Tags boards
// @Produce json
// @Param cursor query int false "Pagination cursor"
// @Param limit query int false "Limit results" default(20)
// @Success 200 {array} BoardResponse
// @Failure 400 {object} errors.ErrorResponse
// @Router /board [get]
func (h *Handler) ListBoards(c *gin.Context) {
	cursor, limit, err := helpers.GetCursorLimit(c)
	if err != nil {
		errors.RespondError(c, errors.WrapBadRequest(err))
		return
	}

	boards, err := h.boardService.List(c, cursor, limit)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	responses := make([]BoardResponse, 0, len(boards))
	for _, board := range boards {
		responses = append(responses, NewBoardResponse(&board))
	}

	helpers.RespondJSON(c, http.StatusOK, responses)
}

// GetBoard godoc
// @Summary Gets board metadata by id
// @Tags boards
// @Produce json
// @Param id path uint64 true "Board ID"
// @Success 200 {object} BoardResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /board/{id} [get]
func (h *Handler) GetBoard(c *gin.Context) {
	id, err := helpers.ParseID(c)
	if err != nil {
		errors.RespondError(c, errors.WrapBadRequest(err))
		return
	}

	board, err := h.boardService.Get(c, id)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	helpers.RespondJSON(c, http.StatusOK, NewBoardResponse(board))
}

// PostBoard godoc
// @Summary Creates new board
// @Tags boards
// @Produce json
// @Param request body BoardCreateRequest true "Board creation request"
// @Success 201 {object} BoardResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /board [post]
func (h *Handler) PostBoard(c *gin.Context) {
	helpers.WithUser(c, func(user *user.User) error {
		var req board.BoardCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.WrapBadRequest(err)
		}

		board, err := h.boardService.Create(c, user.ID, &req)
		if err != nil {
			return err
		}

		helpers.RespondJSON(c, http.StatusCreated, NewBoardResponse(board))
		return nil
	})
}

// PatchBoard godoc
// @Summary Patches existing board by id
// @Tags boards
// @Produce json
// @Param id path uint64 true "Board ID"
// @Param request body BoardPatchRequest true "Board patch request"
// @Success 200 {object} BoardResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /board/{id} [patch]
func (h *Handler) PatchBoard(c *gin.Context) {
	helpers.WithUser(c, func(user *user.User) error {
		id, err := helpers.ParseID(c)
		if err != nil {
			return err
		}

		var req board.BoardPatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.WrapBadRequest(err)
		}

		board, err := h.boardService.Update(c, user.ID, id, &req)
		if err != nil {
			return err
		}

		helpers.RespondJSON(c, http.StatusOK, NewBoardResponse(board))
		return nil
	})
}

// DeleteBoard godoc
// @Summary Delete board by id
// @Tags boards
// @Produce json
// @Param id path uint64 true "Board ID"
// @Success 204
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /board/{id} [delete]
func (h *Handler) DeleteBoard(c *gin.Context) {
	helpers.WithUser(c, func(user *user.User) error {
		id, err := helpers.ParseID(c)
		if err != nil {
			return err
		}

		err = h.boardService.Delete(c, user.ID, id)
		if err != nil {
			return err
		}

		c.Status(http.StatusNoContent)
		return nil
	})
}

// ListBoardImages godoc
// @Summary List board items
// @Description Returns a slice of board's items
// @Tags boards
// @Produce json
// @Param id path uint64 true "Board ID"
// @Param cursor query int false "Pagination cursor"
// @Param limit query int false "Limit results" default(20)
// @Success 200 {array} BoardItemResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /board/{id}/image [get]
func (h *Handler) ListBoardImages(c *gin.Context) {
	boardID, err := helpers.ParseID(c)
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	cursor, limit, err := helpers.GetCursorLimit(c)
	if err != nil {
		errors.RespondError(c, err)
	}

	items, err := h.boardService.ListImages(c, boardID, cursor, limit)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	responses := make([]BoardItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, NewBoardItemResponse(&item))
	}

	helpers.RespondJSON(c, http.StatusOK, responses)
}

func parseBoardImageIDs(c *gin.Context) (uint, uint, error) {
	boardID, err := helpers.ParseID(c)
	if err != nil {
		return 0, 0, errors.WrapBadRequest(err)
	}

	imageID, err := strconv.ParseUint(c.Param("imageId"), 10, 64)
	if err != nil {
		return 0, 0, errors.WrapBadRequest(err)
	}

	return boardID, uint(imageID), nil
}

// PostBoardImage godoc
// @Summary Adds image to a board
// @Tags boards
// @Produce json
// @Param id path uint64 true "Board ID"
// @Param request body BoardItemPostRequest true "Board item creation request"
// @Success 201 {object} BoardResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /board/{id}/image [post]
func (h *Handler) PostBoardImage(c *gin.Context) {
	helpers.WithUser(c, func(user *user.User) error {
		id, err := helpers.ParseID(c)
		if err != nil {
			return err
		}

		var req BoardItemPostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.WrapBadRequest(err)
		}

		item, err := h.boardService.AddImage(c, id, uint(req.ImageID), user.ID)
		if err != nil {
			return err
		}

		helpers.RespondJSON(c, http.StatusCreated, NewBoardItemResponse(item))
		return nil
	})
}

// GetBoardImage godoc
// @Summary Gets image from the board
// @Tags boards
// @Produce json
// @Param id path uint64 true "Board ID"
// @Param imageID path uint64 true "Image ID"
// @Success 200 {object} BoardItemResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /board/{id}/image/{imageID} [get]
func (h *Handler) GetBoardImage(c *gin.Context) {
	id, imageID, err := parseBoardImageIDs(c)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	item, err := h.boardService.GetImage(c, id, imageID)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	helpers.RespondJSON(c, http.StatusOK, NewBoardItemResponse(item))
}

// PatchBoardImage godoc
// @Summary Patches existing board image
// @Tags boards
// @Produce json
// @Param id path uint64 true "Board ID"
// @Param imageID path uint64 true "Image ID"
// @Param request body BoardItemPatchRequest true "Board item patch request"
// @Success 200 {object} BoardResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /board/{id}/image/{imageID} [patch]
func (h *Handler) PatchBoardImage(c *gin.Context) {
	errors.RespondError(c, errors.ErrNotFound)
	// helpers.WithUser(c, func(user *user.User) error {
	// id, imageID, err := parseBoardImageIDs(c)
	// if err != nil {
	// 	return err
	// }
	// item, err := h.boardService.UpdateImage(c, id, imageID, user.ID)
	// if err != nil {
	// 	return err
	// }

	// helpers.RespondJSON(c, http.StatusOK, item)
	// return nil
	// })
}

// DeleteBoardImage godoc
// @Summary Removes image from board by id
// @Tags boards
// @Produce json
// @Param id path uint64 true "Board ID"
// @Param imageID path uint64 true "Image ID"
// @Success 204
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /board/{id}/image/{imageID} [delete]
func (h *Handler) DeleteBoardImage(c *gin.Context) {
	helpers.WithUser(c, func(user *user.User) error {
		id, imageID, err := parseBoardImageIDs(c)
		if err != nil {
			return err
		}

		if err := h.boardService.RemoveImage(c, id, imageID, user.ID); err != nil {
			return err
		}

		c.Status(http.StatusNoContent)
		return nil
	})
}
