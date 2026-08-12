package image

import (
	"net/http"
	"server/internal/image"
	"server/internal/tag"
	"time"

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

func NewImageResponse(img *image.ImageMetadata) image.ImageResponse {
	return image.ImageResponse{
		ID:       img.ID,
		Hash:     img.Hash,
		Filename: img.Filename,
		Tags:     tag.TagsToStrings(img.Tags),
		Format:   img.Format,
		Width:    img.Width,
		Height:   img.Height,
		UserID:   img.UserID,
		Uploaded: img.CreatedAt.Format(time.RFC3339),
	}
}
