package api

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"server/internal/errors"
	"server/internal/image"
	"server/internal/tag"
	"server/internal/user"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const TimeFormat = time.RFC3339

func getRawContentType(meta any) string {
	if img, ok := meta.(*image.ImageMetadata); ok {
		return img.ParseFormat().MIMEType()
	}
	return "application/octet-stream"
}

func (api *api) GetImageByName(c *gin.Context) {
	HandleGet(c, func() (*image.ImageMetadata, error) {
		return api.imageService.GetByName(c.Param("name"))
	}, func(img *image.ImageMetadata) any { return NewImageResponse(img) })
}

func (api *api) GetImageByHash(c *gin.Context) {
	HandleGet(c, func() (*image.ImageMetadata, error) {
		return api.imageService.GetByHash(c.Param("hash"))
	}, func(img *image.ImageMetadata) any { return NewImageResponse(img) })
}

func (api *api) GetImageByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, strconv.IntSize)
	if err != nil {
		RespondError(c, errors.ErrBadRequest)
		return
	}
	HandleGet(c, func() (*image.ImageMetadata, error) {
		return api.imageService.GetByID(uint(id64))
	}, func(img *image.ImageMetadata) any { return NewImageResponse(img) })
}

func (api *api) GetRawImageByName(c *gin.Context) {
	name := c.Param("name")
	HandleStream(c, func(ctx context.Context) (io.ReadCloser, any, error) {
		img, err := api.imageService.GetByName(name)
		if err != nil {
			return nil, nil, err
		}
		return api.objectService.GetRawImageByHash(ctx, img.Hash)
	}, getRawContentType)
}

func (api *api) GetRawThumbnailByName(c *gin.Context) {
	name := c.Param("name")
	HandleStream(c, func(ctx context.Context) (io.ReadCloser, any, error) {
		img, err := api.imageService.GetByName(name)
		if err != nil {
			return nil, nil, err
		}
		return api.objectService.GetRawThumbnailByHash(ctx, img.Hash)
	}, getRawContentType)
}

func (api *api) GetRawImageByHash(c *gin.Context) {
	hash := c.Param("hash")
	HandleStream(c, func(ctx context.Context) (io.ReadCloser, any, error) {
		return api.objectService.GetRawImageByHash(ctx, hash)
	}, getRawContentType)
}

func (api *api) GetRawThumbnailByHash(c *gin.Context) {
	hash := c.Param("hash")
	HandleStream(c, func(ctx context.Context) (io.ReadCloser, any, error) {
		return api.objectService.GetRawThumbnailByHash(ctx, hash)
	}, getRawContentType)
}

func (api *api) GetImagesByQuery(c *gin.Context) {
	WithQuery(c, api.newQueryFromContext, func(query *image.Query) error {
		images, err := api.imageService.Search(query)
		if err != nil {
			return err
		}
		RespondImages(c, images)
		return nil
	})
}

func (api *api) DeleteImageByName(c *gin.Context) {
	name := strings.ToLower(c.Param("name"))
	WithUser(c, api, func(usr *user.User) error {
		return api.imageService.DeleteByName(c.Request.Context(), usr, name)
	})
}

func (api *api) DeleteImagesByQuery(c *gin.Context) {
	WithUserAndQuery(c, api, api.newQueryFromContext, func(usr *user.User, query *image.Query) error {
		errs := api.imageService.DeleteByQuery(c.Request.Context(), usr, query)
		if errs != nil {
			hasPermissionError := false
			messages := make([]string, len(errs))
			for i, err := range errs {
				messages[i] = err.Error()
				if err == errors.ErrPermissionDenied {
					hasPermissionError = true
				}
			}
			if hasPermissionError {
				c.JSON(http.StatusForbidden, gin.H{"errors": messages})
				return nil
			}
			c.JSON(http.StatusBadRequest, gin.H{"errors": messages})
		}
		return nil
	})
}

func (api *api) PostImage(c *gin.Context) {
	formData := c.PostForm("metadata")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		RespondError(c, errors.ErrBadRequest)
		return
	}

	img, err := HandleUpload(c, api, formData, fileHeader)
	if err != nil {
		RespondError(c, err)
		return
	}
	RespondImage(c, img)
}

func (api *api) PostImagesBatch(c *gin.Context) {
	formData := c.PostForm("metadata")
	form, err := c.MultipartForm()
	if err != nil {
		RespondError(c, errors.ErrBadRequest)
		return
	}

	response, err := HandleBatchUpload(c, api, formData, form)
	if err != nil {
		RespondError(c, err)
		return
	}

	if len(response.Failures) > 0 {
		c.JSON(http.StatusMultiStatus, response)
		return
	}
	c.JSON(http.StatusOK, response)
}

func isImageRequestValid(metadata *image.ImagePostRequest, fileHeader *multipart.FileHeader) error {
	if len(metadata.Name) == 0 {
		return errors.ErrBadRequest
	}

	imgFormat, err := image.FormatFromMIME(fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return fmt.Errorf("image format parsing error: %v", err)
	}

	if !imgFormat.IsSupported() {
		return fmt.Errorf("unsupported image format: %s", imgFormat)
	}

	return nil
}

func (api *api) newQueryFromContext(c *gin.Context) (*image.Query, error) {
	cursor, limit, err := GetCursorLimit(c)
	if err != nil {
		return nil, err
	}

	prefix := c.Query("prefix")
	tagsString := c.Query("tags")
	username := c.Query("username")
	userIDStr := c.Query("user_id")

	builder := image.NewQueryBuilder().
		Prefix(prefix).
		Cursor(cursor).
		Limit(limit)

	if tagsString != "" {
		tags, err := tag.ParseTagsFromJSONString(tagsString)
		if err != nil {
			return nil, wrapBadRequest(err)
		}
		builder.Tags(tags)
	}

	if userIDStr != "" {
		userID64, err := strconv.ParseUint(userIDStr, 10, 64)
		userID := uint(userID64)
		if err != nil {
			return nil, wrapBadRequest(err)
		}
		user, err := api.userService.GetByID(userID)
		if err != nil {
			return nil, err
		}
		builder.User(user)
	} else if username != "" {
		user, err := api.userService.GetByUsername(username)
		if err != nil {
			return nil, err
		}
		builder.User(user)
	}

	return builder.Build(), nil
}