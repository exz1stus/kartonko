package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAuditLogEntriesRequest godoc
// @Summary Gets audit log entries by given range
// @Description Returns a list of log entries at "cursor + limit" matching the "query"
// @Tags AuditLog
// @Produce  json
// @Param   cursor query string true "cursor"
// @Param   limit query string true "limit"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /log [get]
func (api *api) GetAuditLogEntriesRequest(c *gin.Context) {
	cursor, limit, err := parseCursorLimit(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	entries, err := api.models.Log.GetEntries(cursor, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries})
}
