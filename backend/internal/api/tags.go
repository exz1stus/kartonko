package api

import (
	"net/http"
	"server/internal/api/dto"
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
	tags, err := api.tagService.SearchPrefix(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

func (api *api) PostTag(c *gin.Context) {
	var req dto.PostTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	tag, err := api.tagService.Create(c.Request.Context(), req.Name, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tag": tag})
}

func (api *api) PostTagsBatch(c *gin.Context) {
	var req dto.PostTagsBatchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body or empty names list",
		})
		return
	}

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	created := make([]string, 0, len(req.Names))
	errs := make([]string, 0)

	for _, name := range req.Names {
		tag, err := api.tagService.Create(
			c.Request.Context(),
			name,
			user,
		)

		if err != nil {
			errs = append(errs, err.Error())
			continue
		}

		created = append(created, tag.Name)
	}

	response := gin.H{
		"created": created,
	}

	if len(errs) > 0 {
		response["errors"] = errs
		c.JSON(http.StatusMultiStatus, response)
		return
	}

	c.JSON(http.StatusCreated, response)
}
