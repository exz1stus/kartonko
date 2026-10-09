package board

import (
	"net/http"
	"server/internal/api/helpers"
	"server/internal/board"
	"server/internal/errors"
	"server/internal/user"
	"strconv"
	"strings"

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
	public.GET("/image/:imageId/board-ids", h.ListBoardIDsForImage)
	public.GET("/slug/:slug", h.GetBoardBySlug)

	protected.POST("", h.PostBoard)
	public.GET("/:id", h.GetBoard)
	protected.PATCH("/:id", h.PatchBoard)
	protected.DELETE("/:id", h.DeleteBoard)
	public.GET("/:id/image", h.ListBoardImages)
	public.GET("/:id/image-ids", h.ListBoardImageIDs)

	protected.POST("/:id/image", h.PostBoardImage)
	public.GET("/:id/image/:imageId", h.GetBoardImage)
	protected.DELETE("/:id/image/:imageId", h.DeleteBoardImage)
}

// GetBoardBySlug godoc
// @Summary Gets board metadata by its readable slug
// @Tags boards
// @Produce json
// @Param slug path string true "Board slug"
// @Success 200 {object} BoardResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /board/slug/{slug} [get]
// @ID GetBoardBySlug
func (h *Handler) GetBoardBySlug(c *gin.Context) {
	board, err := h.boardService.GetBySlug(c, c.Param("slug"))
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	helpers.RespondJSON(c, http.StatusOK, NewBoardResponse(board))
}

// ListBoardIDsForImage godoc
// @Summary List IDs of boards containing an image
// @Tags boards
// @Produce json
// @Param imageId path uint64 true "Image ID"
// @Success 200 {array} integer
// @Router /board/image/{imageId}/board-ids [get]
// @ID ListBoardIDsForImage
func (h *Handler) ListBoardIDsForImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("imageId"), 10, 64)
	if err != nil || id == 0 {
		errors.RespondError(c, errors.ErrBadRequest)
		return
	}
	ids, err := h.boardService.BoardIDsForImage(c, uint(id))
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	helpers.RespondJSON(c, http.StatusOK, ids)
}

// ListBoardImageIDs godoc
// @Summary List IDs of images on a board
// @Tags boards
// @Produce json
// @Param id path uint64 true "Board ID"
// @Success 200 {array} integer
// @Router /board/{id}/image-ids [get]
// @ID ListBoardImageIDs
func (h *Handler) ListBoardImageIDs(c *gin.Context) {
	id, err := helpers.ParseID(c)
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	ids, err := h.boardService.ImageIDs(c, id)
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	helpers.RespondJSON(c, http.StatusOK, ids)
}

// ListBoards godoc
// @Summary List boards
// @Description Returns a slice of list of boards
// @Tags boards
// @Produce json
// @Param cursor query int false "Pagination cursor"
// @Param limit query int false "Limit results" default(20)
// @Param name query string false "Board name contains"
// @Param user_id query int false "Board owner ID"
// @Success 200 {array} BoardResponse
// @Failure 400 {object} errors.ErrorResponse
// @Router /board [get]
// @ID ListBoards
func (h *Handler) ListBoards(c *gin.Context) {
	cursor, limit, err := helpers.GetCursorLimit(c)
	if err != nil {
		errors.RespondError(c, errors.WrapBadRequest(err))
		return
	}

	var ownerID uint64
	if raw := c.Query("user_id"); raw != "" {
		ownerID, err = strconv.ParseUint(raw, 10, 64)
		if err != nil || ownerID == 0 {
			errors.RespondError(c, errors.ErrBadRequest)
			return
		}
	}

	boards, err := h.boardService.List(c, c.Query("name"), uint(ownerID), cursor, limit)
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
// @ID GetBoard
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
// @ID PostBoard
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
// @ID PatchBoard
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
// @ID DeleteBoard
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
// @Param prefix query string false "Image filename prefix"
// @Param tags query []string false "Image tags" collectionFormat(csv)
// @Success 200 {array} BoardItemResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /board/{id}/image [get]
// @ID ListBoardImages
func (h *Handler) ListBoardImages(c *gin.Context) {
	boardID, err := helpers.ParseID(c)
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	cursor, limit, err := helpers.GetCursorLimit(c)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	var tags []string
	for _, value := range c.QueryArray("tags") {
		for _, tag := range strings.Split(value, ",") {
			if tag = strings.TrimSpace(tag); tag != "" {
				tags = append(tags, tag)
			}
		}
	}
	items, err := h.boardService.ListImages(c, boardID, c.Query("prefix"), tags, cursor, limit)
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
// @Success 201 {object} BoardItemResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /board/{id}/image [post]
// @ID PostBoardImage
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
// @ID GetBoardImage
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
// @ID DeleteBoardImage
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
