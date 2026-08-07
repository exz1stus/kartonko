package api

import (
	"server/internal/api/dto"

	"github.com/gin-gonic/gin"
)

func (api *api) GetAuditLogEntries(c *gin.Context) {
	HandleList(c, func(cursor, limit int) ([]dto.EntryResponse, error) {
		return api.logService.GetEntries(cursor, limit)
	}, func(entry dto.EntryResponse) any { return entry })
}
