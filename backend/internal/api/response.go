package api

import (
	"errors"
	"net/http"
	"time"

	"server/internal/image"
	"server/internal/tag"
	serverrors "server/internal/errors"

	"github.com/gin-gonic/gin"
)

func RespondImage(c *gin.Context, img *image.ImageMetadata) {
	c.JSON(http.StatusOK, NewImageResponse(img))
}

func RespondImages(c *gin.Context, images []image.ImageMetadata) {
	response := make([]image.ImageResponse, len(images))
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

func NewImageResponse(img *image.ImageMetadata) image.ImageResponse {
	return image.ImageResponse{
		ID:       img.ID,
		Hash:     img.Hash,
		Filename: img.Filename,
		Tags:     tagNames(img.Tags),
		Format:   img.Format,
		Width:    img.Width,
		Height:   img.Height,
		UserID:   img.UserID,
		Uploaded: img.CreatedAt.Format(time.RFC3339),
	}
}

func tagNames(tags []tag.Tag) []string {
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names
}