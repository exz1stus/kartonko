package log

import (
	"server/internal/api/helpers"
	"server/internal/log"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	logService log.LogService
}

func NewLogHandler(
	logService log.LogService,
) *Handler {
	return &Handler{
		logService: logService,
	}
}

func (h *Handler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.GET("/log", h.GetAuditLogEntries)
}

func (h *Handler) GetAuditLogEntries(c *gin.Context) {
	helpers.HandleList(c, func(cursor, limit int) ([]log.EntryResponse, error) {
		return h.logService.GetEntries(cursor, limit)
	}, func(entry log.EntryResponse) any { return entry })
}
