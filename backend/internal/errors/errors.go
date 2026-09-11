package errors

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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

	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound

	case errors.Is(err, ErrPermissionDenied):
		status = http.StatusForbidden

	case errors.Is(err, ErrUnsupportedFormat):
	case errors.Is(err, ErrDuplicateHash):
	case errors.Is(err, ErrDuplicateName):
	case errors.Is(err, ErrBadRequest):
		status = http.StatusBadRequest

	case errors.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
	}

	c.JSON(status, ErrorResponse{Error: msg})
}
