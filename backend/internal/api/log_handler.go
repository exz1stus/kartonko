package api

import (
	"github.com/gin-gonic/gin"
	"server/internal/log"
)

func (api *api) GetAuditLogEntries(c *gin.Context) {
	HandleList(c, func(cursor, limit int) ([]log.EntryResponse, error) {
		return api.logService.GetEntries(cursor, limit)
	}, func(entry log.EntryResponse) any { return entry })
}