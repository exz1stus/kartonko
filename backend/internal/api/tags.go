package api

import (
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

func (api *api) PostTag(c *gin.Context) {
	req := c.Query("name")
	if req == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "name is required"})
		return
	}
	tag, err := api.models.Tags.AddTag(req)
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
