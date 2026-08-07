package api

import (
	"server/internal/api/dto"
	"server/internal/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func NewTagResponse(tag *models.Tag) dto.TagResponse {
	return dto.TagResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}
}

func (api *api) GetTags(c *gin.Context) {
	query := c.Query("query")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil || limit <= 0 {
		limit = 50
	}

	HandleList(c, func(cursor, limit int) ([]models.Tag, error) {
		return api.tagService.SearchPrefix(query, cursor, limit)
	}, func(tag models.Tag) any { return NewTagResponse(&tag) })
}

func (api *api) PostTag(c *gin.Context) {
	var req dto.PostTagRequest
	if err := GetJSON(c, &req, "Name"); err != nil {
		RespondError(c, err)
		return
	}

	WithUser(c, api, func(user *models.User) error {
		tag, err := api.tagService.Create(c.Request.Context(), req.Name, user)
		if err != nil {
			return err
		}
		RespondCreated(c, NewTagResponse(tag))
		return nil
	})
}

func (api *api) PostTagsBatch(c *gin.Context) {
	var req dto.PostTagsBatchRequest
	if err := GetJSON(c, &req, "Names"); err != nil {
		RespondError(c, err)
		return
	}

	WithUser(c, api, func(user *models.User) error {
		HandleBatch(c, req.Names, func(name string) (*models.Tag, error) {
			return api.tagService.Create(c.Request.Context(), name, user)
		}, func(tag *models.Tag) any { return NewTagResponse(tag) }, true)
		return nil
	})
}