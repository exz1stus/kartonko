package image

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"server/internal/api/auth"
	"server/internal/api/helpers"
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

type Handler struct {
	imageService  image.ImageService
	objectService image.ObjectService
	userService   user.UserService
	tagService    tag.TagService
}

func NewImageHandler(
	imageService image.ImageService,
	objectService image.ObjectService,
	userService user.UserService,
	tagService tag.TagService,
) *Handler {
	return &Handler{
		imageService:  imageService,
		objectService: objectService,
		userService:   userService,
		tagService:    tagService,
	}
}

func getRawContentType(meta any) string {
	if img, ok := meta.(*image.ImageMetadata); ok {
		return img.ParseFormat().MIMEType()
	}
	return "application/octet-stream"
}

func HandleUpload(c *gin.Context, images image.ImageService, metadata string, fileHeader *multipart.FileHeader) (*image.ImageMetadata, error) {
	var postRequest image.ImagePostRequest
	if err := json.Unmarshal([]byte(metadata), &postRequest); err != nil {
		return nil, errors.ErrBadRequest
	}

	if err := isImageRequestValid(&postRequest, fileHeader); err != nil {
		return nil, helpers.WrapBadRequest(err)
	}

	user, err := auth.GetUserFromContext(c)
	if err != nil {
		return nil, errors.ErrUnauthorized
	}

	return images.Upload(c.Request.Context(), user, &postRequest, fileHeader)
}

func HandleBatchUpload(c *gin.Context, images image.ImageService, metadata string, form *multipart.Form) (*image.ImagePostBatchResponse, error) {
	var batch image.ImagePostBatchRequest
	if err := json.Unmarshal([]byte(metadata), &batch); err != nil {
		return nil, helpers.WrapBadRequest(err)
	}

	files := form.File["files"]
	if files == nil {
		return nil, errors.ErrBadRequest
	}

	if len(batch.Data) != len(files) {
		return nil, errors.ErrBadRequest
	}

	user, err := auth.GetUserFromContext(c)
	if err != nil {
		return nil, errors.ErrUnauthorized
	}

	response := &image.ImagePostBatchResponse{}

	for i, meta := range batch.Data {
		if i >= len(files) {
			response.Failures = append(response.Failures, image.ImageError{
				Name:  meta.Name,
				Error: "No file provided for this metadata",
			})
			continue
		}

		if err := isImageRequestValid(&meta, files[i]); err != nil {
			response.Failures = append(response.Failures, image.ImageError{
				Name:  meta.Name,
				Error: helpers.WrapBadRequest(err).Error(),
			})
			continue
		}

		img, err := images.Upload(c.Request.Context(), user, &meta, files[i])
		if err != nil {
			response.Failures = append(response.Failures, image.ImageError{
				Name:  meta.Name,
				Error: err.Error(),
			})
			continue
		}

		response.Successes = append(response.Successes, NewImageResponse(img))
	}

	return response, nil
}

func (h *Handler) GetImageByName(c *gin.Context) {
	helpers.HandleGet(c, func() (*image.ImageMetadata, error) {
		return h.imageService.GetByName(c.Param("name"))
	}, func(img *image.ImageMetadata) any { return NewImageResponse(img) })
}

func (h *Handler) GetImageByHash(c *gin.Context) {
	helpers.HandleGet(c, func() (*image.ImageMetadata, error) {
		return h.imageService.GetByHash(c.Param("hash"))
	}, func(img *image.ImageMetadata) any { return NewImageResponse(img) })
}

func (h *Handler) GetImageByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, strconv.IntSize)
	if err != nil {
		errors.RespondError(c, errors.ErrBadRequest)
		return
	}
	helpers.HandleGet(c, func() (*image.ImageMetadata, error) {
		return h.imageService.GetByID(uint(id64))
	}, func(img *image.ImageMetadata) any { return NewImageResponse(img) })
}

func (h *Handler) GetRawImageByName(c *gin.Context) {
	name := c.Param("name")
	helpers.HandleStream(c, func(ctx context.Context) (io.ReadCloser, any, error) {
		img, err := h.imageService.GetByName(name)
		if err != nil {
			return nil, nil, err
		}
		return h.objectService.GetRawImageByHash(ctx, img.Hash)
	}, getRawContentType)
}

func (h *Handler) GetRawThumbnailByName(c *gin.Context) {
	name := c.Param("name")
	helpers.HandleStream(c, func(ctx context.Context) (io.ReadCloser, any, error) {
		img, err := h.imageService.GetByName(name)
		if err != nil {
			return nil, nil, err
		}
		return h.objectService.GetRawThumbnailByHash(ctx, img.Hash)
	}, getRawContentType)
}

func (h *Handler) GetRawImageByHash(c *gin.Context) {
	hash := c.Param("hash")
	helpers.HandleStream(c, func(ctx context.Context) (io.ReadCloser, any, error) {
		return h.objectService.GetRawImageByHash(ctx, hash)
	}, getRawContentType)
}

func (h *Handler) GetRawThumbnailByHash(c *gin.Context) {
	hash := c.Param("hash")
	helpers.HandleStream(c, func(ctx context.Context) (io.ReadCloser, any, error) {
		return h.objectService.GetRawThumbnailByHash(ctx, hash)
	}, getRawContentType)
}

func (h *Handler) GetImagesByQuery(c *gin.Context) {
	helpers.WithQuery(c, h.newQueryFromContext, func(query *image.Query) error {
		images, err := h.imageService.Search(query)
		if err != nil {
			return err
		}
		RespondImages(c, images)
		return nil
	})
}

func (h *Handler) DeleteImageByName(c *gin.Context) {
	name := strings.ToLower(c.Param("name"))
	helpers.WithUser(c, func(usr *user.User) error {
		return h.imageService.DeleteByName(c.Request.Context(), usr, name)
	})
}

func (h *Handler) DeleteImagesByQuery(c *gin.Context) {
	helpers.WithUserAndQuery(c, h.newQueryFromContext, func(usr *user.User, query *image.Query) error {
		errs := h.imageService.DeleteByQuery(c.Request.Context(), usr, query)
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

func (h *Handler) PostImage(c *gin.Context) {
	formData := c.PostForm("metadata")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		errors.RespondError(c, errors.ErrBadRequest)
		return
	}

	img, err := HandleUpload(c, h.imageService, formData, fileHeader)
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	RespondImage(c, img)
}

func (h *Handler) PostImagesBatch(c *gin.Context) {
	formData := c.PostForm("metadata")
	form, err := c.MultipartForm()
	if err != nil {
		errors.RespondError(c, errors.ErrBadRequest)
		return
	}

	response, err := HandleBatchUpload(c, h.imageService, formData, form)
	if err != nil {
		errors.RespondError(c, err)
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

func (h *Handler) newQueryFromContext(c *gin.Context) (*image.Query, error) {
	cursor, limit, err := helpers.GetCursorLimit(c)
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
			return nil, helpers.WrapBadRequest(err)
		}
		builder.Tags(tags)
	}

	if userIDStr != "" {
		userID64, err := strconv.ParseUint(userIDStr, 10, 64)
		userID := uint(userID64)
		if err != nil {
			return nil, helpers.WrapBadRequest(err)
		}
		user, err := h.userService.GetByID(userID)
		if err != nil {
			return nil, err
		}
		builder.User(user)
	} else if username != "" {
		user, err := h.userService.GetByUsername(username)
		if err != nil {
			return nil, err
		}
		builder.User(user)
	}

	return builder.Build(), nil
}
