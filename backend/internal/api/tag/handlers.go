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
	public.GET("/tags", h.GetTags)

	protected.POST("/tag", h.PostTag)
	protected.POST("/tags/batch", h.PostTagsBatch)
}

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

func (h *Handler) PostTag(c *gin.Context) {
	helpers.WithUser(c, func(usr *user.User) error {
		var req tag.PostTagRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.ErrBadRequest
		}

		t, err := h.tagService.Create(c.Request.Context(), req.Name, usr)
		if err != nil {
			return err
		}

		helpers.RespondCreated(c, tag.TagResponse{ID: t.ID, Name: t.Name})
		return nil
	})
}

func (h *Handler) PostTagsBatch(c *gin.Context) {
	helpers.WithUser(c, func(usr *user.User) error {
		var req tag.PostTagsBatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.ErrBadRequest
		}

		var results []tag.TagResponse
		var failures []struct {
			Name  string `json:"name"`
			Error string `json:"error"`
		}

		for _, name := range req.Names {
			t, err := h.tagService.Create(c.Request.Context(), name, usr)
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
