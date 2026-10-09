package tag

import (
	"net/http"
	"server/internal/api/helpers"
	"server/internal/errors"
	tagpkg "server/internal/tag"
	"server/internal/user"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	tagService tagpkg.TagService
}

func NewTagHandler(
	tagService tagpkg.TagService,
) *Handler {
	return &Handler{
		tagService: tagService,
	}
}

func (h *Handler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.GET("", h.GetTags)
	protected.POST("", h.PostTag)
	protected.PATCH("/:id", h.PatchTag)
	protected.DELETE("/:id", h.DeleteTag)

	protected.POST("/batch", h.PostTagsBatch)
}

// GetTags godoc
// @Summary Gets all tags
// @Description Returns a paginated list of all tags
// @Tags tags
// @Produce json
// @Param prefix query string false "Tag name prefix"
// @Param cursor query int false "Pagination cursor"
// @Param limit query int false "Limit results" default(20)
// @Success 200 {array} TagResponse
// @Failure 400 {object} errors.ErrorResponse
// @Router /tags [get]
// @ID GetTags
func (h *Handler) GetTags(c *gin.Context) {
	helpers.HandleList(c, func(cursor, limit int) ([]tagpkg.Tag, error) {
		prefix := c.Query("prefix")
		tags, err := h.tagService.SearchPrefix(c, prefix, cursor, limit)
		if err != nil {
			return nil, err
		}

		return tags, nil
	}, func(t *tagpkg.Tag) any { return tagpkg.NewTagResponse(t) })
}

// PostTag godoc
// @Summary Creates a new tag
// @Description Creates a new tag (requires authentication)
// @Tags tags
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body TagPostRequest true "Tag name"
// @Success 201 {object} TagResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /tags [post]
// @ID PostTag
func (h *Handler) PostTag(c *gin.Context) {
	helpers.WithUser(c, func(usr *user.User) error {
		var req tagpkg.TagPostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.WrapBadRequest(err)
		}

		t, err := h.tagService.Create(c, &req, usr.ID)
		if err != nil {
			return err
		}

		helpers.RespondJSON(c, http.StatusCreated, tagpkg.NewTagResponse(t))
		return nil
	})
}

// PostTagsBatch godoc
// @Summary Creates multiple tags in batch
// @Description Creates multiple tags in a single request (requires authentication)
// @Tags tags
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body TagPostBatchRequest true "Tag names"
// @Success 200 {object} TagBatchResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /tags/batch [post]
// @ID PostTagsBatch
func (h *Handler) PostTagsBatch(c *gin.Context) {
	helpers.WithUser(c, func(usr *user.User) error {
		var req tagpkg.TagPostBatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.WrapBadRequest(err)
		}

		names := req.Names
		var results []tagpkg.TagResponse
		var failures []struct {
			Name  string `json:"name"`
			Error string `json:"error"`
		}

		for _, name := range names {
			tagReq := &tagpkg.TagPostRequest{Name: name}
			t, err := h.tagService.Create(c, tagReq, usr.ID)
			if err != nil {
				failures = append(failures, struct {
					Name  string `json:"name"`
					Error string `json:"error"`
				}{Name: name, Error: err.Error()})
				continue
			}
			results = append(results, tagpkg.NewTagResponse(t))
		}

		response := tagpkg.TagBatchResponse{
			Successes: results,
			Failures:  failures,
		}

		if len(failures) > 0 && len(results) == 0 {
			helpers.RespondJSON(c, http.StatusBadRequest, response)
			return nil
		}

		if len(failures) > 0 {
			helpers.RespondJSON(c, http.StatusMultiStatus, response)
			return nil
		}

		helpers.RespondJSON(c, http.StatusCreated, response)
		return nil
	})
}

// PatchTag godoc
// @Summary Patches existing tag by id
// @Tags tags
// @Produce json
// @Param id path uint64 true "Tag ID"
// @Param request body TagPatchRequest true "Tag patch request"
// @Success 200 {object} TagResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /tags/{id} [patch]
// @ID PatchTag
func (h *Handler) PatchTag(c *gin.Context) {
	helpers.WithUser(c, func(user *user.User) error {
		id, err := helpers.ParseID(c)
		if err != nil {
			return err
		}

		var req tagpkg.TagPatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.WrapBadRequest(err)
		}

		tag, err := h.tagService.Update(c, user.ID, id, &req)
		if err != nil {
			return err
		}

		helpers.RespondJSON(c, http.StatusOK, tagpkg.NewTagResponse(tag))
		return nil
	})
}

// DeleteTag godoc
// @Summary Delete tag by id
// @Tags tags
// @Produce json
// @Param id path uint64 true "Tag ID"
// @Success 204
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /tags/{id} [delete]
// @ID DeleteTag
func (h *Handler) DeleteTag(c *gin.Context) {
	helpers.WithUser(c, func(user *user.User) error {
		id, err := helpers.ParseID(c)
		if err != nil {
			return err
		}

		h.tagService.Delete(c, user.ID, id)
		c.Status(http.StatusNoContent)
		return nil
	})
}
