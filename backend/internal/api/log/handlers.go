package log

import (
	"server/internal/api/helpers"
	logpkg "server/internal/log"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	logService logpkg.LogService
}

func NewLogHandler(
	logService logpkg.LogService,
) *Handler {
	return &Handler{
		logService: logService,
	}
}

func (h *Handler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.GET("", h.GetAuditLogEntries)
}

// GetAuditLogEntries godoc
// @Summary Gets audit log entries
// @Description Returns a paginated list of audit log entries
// @Tags log
// @Produce json
// @Param cursor query int false "Pagination cursor"
// @Param limit query int false "Limit results" default(20)
// @Success 200 {array} log.EntryResponse
// @Failure 400 {object} errors.ErrorResponse
// @Router /log [get]
func (h *Handler) GetAuditLogEntries(c *gin.Context) {
	helpers.HandleList(c, func(cursor, limit int) ([]EntryResponse, error) {
		entries, err := h.logService.GetEntries(cursor, limit)
		if err != nil {
			return nil, err
		}
		return FromServiceEntries(entries), nil
	}, func(entry EntryResponse) any { return entry })
}
