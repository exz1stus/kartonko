package log

import (
	"net/http"
	"server/internal/api/helpers"
	"server/internal/errors"
	logpkg "server/internal/log"
	"server/internal/user"
	"strconv"

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
	protected.GET("", h.GetAuditLogEntries)
	protected.POST("/:id/revert", h.RevertAuditLogEntry)
}

// RevertAuditLogEntry godoc
// @Summary Reverts a reversible audit entry
// @Tags log
// @Produce json
// @Param id path uint true "Entry ID"
// @Success 200 {object} log.EntryResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Router /log/{id}/revert [post]
// @ID RevertAuditLogEntry
func (h *Handler) RevertAuditLogEntry(c *gin.Context) {
	helpers.WithUser(c, func(actor *user.User) error {
		if !actor.IsModerator() {
			return errors.ErrPermissionDenied
		}
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			return errors.ErrBadRequest
		}
		entry, err := h.logService.Revert(uint(id), actor.ID)
		if err != nil {
			return err
		}
		c.JSON(http.StatusOK, entry)
		return nil
	})
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
// @ID GetAuditLogEntries
func (h *Handler) GetAuditLogEntries(c *gin.Context) {
	helpers.WithUser(c, func(actor *user.User) error {
		if !actor.IsModerator() {
			return errors.ErrPermissionDenied
		}
		helpers.HandleList(c, func(cursor, limit int) ([]logpkg.EntryResponse, error) {
			entries, err := h.logService.GetEntries(cursor, limit)
			if err != nil {
				return nil, err
			}
			return entries, nil
		}, func(entry *logpkg.EntryResponse) any { return entry })
		return nil
	})
}
