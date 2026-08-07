package api

import (
	"errors"
	"net/http"

	"server/internal/api/dto"
	serverrors "server/internal/errors"
	"server/internal/models"

	"github.com/gin-gonic/gin"
)

func RespondImage(c *gin.Context, img *models.ImageMetadata) {
	c.JSON(http.StatusOK, NewImageResponse(img))
}

func RespondImages(c *gin.Context, images []models.ImageMetadata) {
	response := make([]dto.ImageResponse, len(images))
	for i, img := range images {
		response[i] = NewImageResponse(&img)
	}
	c.JSON(http.StatusOK, response)
}

func RespondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	msg := err.Error()

	switch {
	case errors.Is(err, serverrors.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, serverrors.ErrPermissionDenied):
		status = http.StatusForbidden
	case errors.Is(err, serverrors.ErrBadRequest):
		status = http.StatusBadRequest
	case errors.Is(err, serverrors.ErrUnauthorized):
		status = http.StatusUnauthorized
	}

	c.JSON(status, ErrorResponse{Error: msg})
}
