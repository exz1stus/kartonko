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
	protected.POST("/batch", h.PostTagsBatch)
}

// GetTags godoc
// @Summary Gets all tags
// @Description Returns a paginated list of all tags
// @Tags tags
// @Produce json
// @Param cursor query int false "Pagination cursor"
// @Param limit query int false "Limit results" default(20)
// @Success 200 {array} TagResponse
// @Failure 400 {object} errors.ErrorResponse
// @Router /tags [get]
func (h *Handler) GetTags(c *gin.Context) {
	helpers.HandleList(c, func(cursor, limit int) ([]TagResponse, error) {
		tags, err := h.tagService.SearchPrefix("", cursor, limit)
		if err != nil {
			return nil, err
		}
		return FromServiceTags(tags), nil
	}, func(t TagResponse) any { return t })
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
func (h *Handler) PostTag(c *gin.Context) {
	helpers.WithUser(c, func(usr *user.User) error {
		var req TagPostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.ErrBadRequest
		}

		t, err := h.tagService.Create(c.Request.Context(), req.ToServiceCreate(), usr.ID)
		if err != nil {
			return err
		}

		helpers.RespondCreated(c, FromServiceTag(t))
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
func (h *Handler) PostTagsBatch(c *gin.Context) {
	helpers.WithUser(c, func(usr *user.User) error {
		var req TagPostBatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.ErrBadRequest
		}

		names := req.ToServiceCreate()
		var results []TagResponse
		var failures []struct {
			Name  string `json:"name"`
			Error string `json:"error"`
		}

		for _, name := range names {
			t, err := h.tagService.Create(c.Request.Context(), name, usr.ID)
			if err != nil {
				failures = append(failures, struct {
					Name  string `json:"name"`
					Error string `json:"error"`
				}{Name: name, Error: err.Error()})
				continue
			}
			results = append(results, FromServiceTag(t))
		}

		response := TagBatchResponse{
			Successes: results,
			Failures:  failures,
		}

		if len(failures) > 0 && len(results) == 0 {
			helpers.RespondJSON(c, http.StatusBadRequest, response)
			return nil
		}

		if len(failures) > 0 {
			helpers.RespondMultiStatus(c, response)
			return nil
		}

		helpers.RespondCreated(c, response)
		return nil
	})
}
