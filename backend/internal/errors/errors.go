package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrNotFound         = errors.New("not found")
	ErrBadRequest       = errors.New("bad request")
	ErrUnauthorized     = errors.New("unauthorized")
)

// swagger:model
type ErrorResponse struct {
	Error string `json:"error"`
}

func RespondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	msg := err.Error()

	switch {
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrPermissionDenied):
		status = http.StatusForbidden
	case errors.Is(err, ErrBadRequest):
		status = http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
	}

	c.JSON(status, ErrorResponse{Error: msg})
}
