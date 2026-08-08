package api

import (
	"net/http"
	"server/internal/errors"
	"server/internal/tag"
	"server/internal/user"

	"github.com/gin-gonic/gin"
)

func (api *api) GetTags(c *gin.Context) {
	HandleList(c, func(cursor, limit int) ([]tag.TagResponse, error) {
		tags, err := api.tagService.SearchPrefix("", cursor, limit)
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

func (api *api) PostTag(c *gin.Context) {
	WithUser(c, api, func(usr *user.User) error {
		var req tag.PostTagRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.ErrBadRequest
		}

		t, err := api.tagService.Create(c.Request.Context(), req.Name, usr)
		if err != nil {
			return err
		}

		RespondCreated(c, tag.TagResponse{ID: t.ID, Name: t.Name})
		return nil
	})
}

func (api *api) PostTagsBatch(c *gin.Context) {
	WithUser(c, api, func(usr *user.User) error {
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
			t, err := api.tagService.Create(c.Request.Context(), name, usr)
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
			RespondJSON(c, http.StatusBadRequest, response)
			return nil
		}

		if len(failures) > 0 {
			RespondMultiStatus(c, response)
			return nil
		}

		RespondCreated(c, response)
		return nil
	})
}