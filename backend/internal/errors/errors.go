package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ErrInternalServer   = errors.New("internal server error")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNotFound         = errors.New("not found")
	ErrBadRequest       = errors.New("bad request")
	ErrUnauthorized     = errors.New("unauthorized")

	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrDuplicateName     = errors.New("name already exists")
	ErrDuplicateHash     = errors.New("hash already exists")
)

type ErrorResponse struct {
	Error string `json:"error"`
} // @name ErrorResponse

func RespondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	msg := err.Error()

	switch {
	case errors.Is(err, ErrInternalServer):
		status = http.StatusInternalServerError

	case errors.Is(err, ErrNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		status = http.StatusNotFound

	case errors.Is(err, ErrPermissionDenied):
		status = http.StatusForbidden

	case errors.Is(err, ErrUnsupportedFormat),
		errors.Is(err, ErrDuplicateHash),
		errors.Is(err, ErrDuplicateName),
		errors.Is(err, ErrBadRequest):
		status = http.StatusBadRequest

	case errors.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
	}

	c.JSON(status, ErrorResponse{Error: msg})
}
