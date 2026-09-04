package tag

import (
	"net/http"
	"server/internal/api/helpers"
	"server/internal/errors"
	"server/internal/tag"
	"server/internal/user"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	tagService tag.TagService
}

func NewTagHandler(
	tagService tag.TagService,
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
// @Success 200 {array} tag.TagResponse
// @Failure 400 {object} errors.ErrorResponse
// @Router /tags [get]
func (h *Handler) GetTags(c *gin.Context) {
	helpers.HandleList(c, func(cursor, limit int) ([]tag.TagResponse, error) {
		tags, err := h.tagService.SearchPrefix("", cursor, limit)
		if err != nil {
			return nil, err
		}
		resp := make([]tag.TagResponse, len(tags))
		for i, t := range tags {
			resp[i] = tag.TagResponse{ID: t.ID, Name: t.Name}
		}
		return resp, nil
	}, func(t tag.TagResponse) any { return t })
}

// PostTag godoc
// @Summary Creates a new tag
// @Description Creates a new tag (requires authentication)
// @Tags tags
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body tag.TagPostRequest true "Tag name"
// @Success 201 {object} tag.TagResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /tags [post]
func (h *Handler) PostTag(c *gin.Context) {
	helpers.WithUser(c, func(usr *user.User) error {
		var req tag.TagPostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.ErrBadRequest
		}

		t, err := h.tagService.Create(c.Request.Context(), req.Name, usr.ID)
		if err != nil {
			return err
		}

		helpers.RespondCreated(c, tag.TagResponse{ID: t.ID, Name: t.Name})
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
// @Param body body tag.TagPostBatchRequest true "Tag names"
// @Success 200 {object} tag.TagBatchResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /tags/batch [post]
func (h *Handler) PostTagsBatch(c *gin.Context) {
	helpers.WithUser(c, func(usr *user.User) error {
		var req tag.TagPostBatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.ErrBadRequest
		}

		var results []tag.TagResponse
		var failures []struct {
			Name  string `json:"name"`
			Error string `json:"error"`
		}

		for _, name := range req.Names {
			t, err := h.tagService.Create(c.Request.Context(), name, usr.ID)
			if err != nil {
				failures = append(failures, struct {
					Name  string `json:"name"`
					Error string `json:"error"`
				}{Name: name, Error: err.Error()})
				continue
			}
			results = append(results, tag.TagResponse{ID: t.ID, Name: t.Name})
		}

		response := struct {
			Successes []tag.TagResponse `json:"successes"`
			Failures  []struct {
				Name  string `json:"name"`
				Error string `json:"error"`
			} `json:"failures,omitempty"`
		}{
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
