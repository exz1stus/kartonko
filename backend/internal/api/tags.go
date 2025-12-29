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
func (api *api) GetTagsRequest(c *gin.Context) {
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
