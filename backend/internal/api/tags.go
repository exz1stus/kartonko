package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary Searches tags by given query
// @Description Returns a list of tags at "cursor + limit" matching the "query"
// @Tags Tags
// @Produce  json
// @Param   query query string true "query"
// @Param   limit query int true "limit"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tags [get]
func (api *api) GetTags(c *gin.Context) {
	query := c.Query("query")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse limit"})
		return
	}
	tags, err := api.models.Tags.SearchTags(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

type CreateTagRequest struct {
	Name string `json:"name"`
}

func (api *api) PostTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	tag, err := api.models.Tags.AddTag(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if err = api.models.Log.AddTagCreated(tag, user); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tag": tag})
}

type BatchTagsRequest struct {
	Names []string `json:"names" binding:"required,min=1"`
}

func (api *api) PostTagsBatch(c *gin.Context) {
	var req BatchTagsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body or empty names list"})
		return
	}

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	tags, err := api.models.Tags.AddTagsBatch(req.Names)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	for _, tag := range tags {
		if err = api.models.Log.AddTagCreated(&tag, user); err != nil {
			log.Printf("Failed to log tag creation for %s: %v", tag.Name, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}
