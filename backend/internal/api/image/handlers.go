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

func getFormatFromHeader(fileHeader *multipart.FileHeader) (image.Format, error) {
	format, err := image.FormatFromMIME(fileHeader.Header.Get("Content-Type"))
	if err != nil || !format.IsSupported() {
		return image.FormatInvalid, errors.ErrUnsupportedFormat
	}

	return format, nil
}

func getImageBytesFromHeader(fileHeader *multipart.FileHeader) ([]byte, error) {
	f, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening uploaded file: %v", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading uploaded file: %v", err)
	}

	return data, nil
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
	format, err := getFormatFromHeader(fileHeader)
	if err != nil {
		return nil, err
	}
	data, err := getImageBytesFromHeader(fileHeader)
	if err != nil {
		return nil, err
	}

	return images.Upload(c.Request.Context(), user.ID, &postRequest, format, data)
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

		format, err := getFormatFromHeader(files[i])
		if err != nil {
			return nil, err
		}
		data, err := getImageBytesFromHeader(files[i])
		if err != nil {
			return nil, err
		}

		img, err := images.Upload(c.Request.Context(), user.ID, &meta, format, data)
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

// GetImageByName godoc
// @Summary Gets image metadata by name
// @Description Returns image metadata by its name
// @Tags images
// @Produce json
// @Param name path string true "Image name"
// @Success 200 {object} image.ImageResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /image/{name} [get]
func (h *Handler) GetImageByName(c *gin.Context) {
	helpers.HandleGet(c, func() (*image.ImageMetadata, error) {
		return h.imageService.GetByName(c.Param("name"))
	}, func(img *image.ImageMetadata) any { return NewImageResponse(img) })
}

// GetImageByHash godoc
// @Summary Gets image metadata by its unique hash
// @Description Returns image metadata by its hash
// @Tags images
// @Produce json
// @Param hash path string true "Image hash"
// @Success 200 {object} image.ImageResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /image/hash/{hash} [get]
func (h *Handler) GetImageByHash(c *gin.Context) {
	helpers.HandleGet(c, func() (*image.ImageMetadata, error) {
		return h.imageService.GetByHash(c.Param("hash"))
	}, func(img *image.ImageMetadata) any { return NewImageResponse(img) })
}

// GetImageByID godoc
// @Summary Gets image metadata by ID
// @Description Returns image metadata by its numeric ID
// @Tags images
// @Produce json
// @Param id path uint64 true "Image ID"
// @Success 200 {object} image.ImageResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /image/id/{id} [get]
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

// GetRawImageByName godoc
// @Summary Gets raw image by name
// @Description Returns the raw image file by its name
// @Tags images
// @Produce application/octet-stream
// @Param name path string true "Image name"
// @Success 200 {file} binary "Raw image file"
// @Failure 404 {object} errors.ErrorResponse
// @Router /image/raw/{name} [get]
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

// GetRawThumbnailByName godoc
// @Summary Gets raw image thumbnail by name
// @Description Returns the raw thumbnail image file by its name
// @Tags images
// @Produce application/octet-stream
// @Param name path string true "Image thumbnail name"
// @Success 200 {file} binary "Raw thumbnail image file"
// @Failure 404 {object} errors.ErrorResponse
// @Router /image/thumb/{name} [get]
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

// GetRawImageByHash godoc
// @Summary Gets raw image by hash
// @Description Returns the raw image file by its hash
// @Tags images
// @Produce application/octet-stream
// @Param hash path string true "Image hash"
// @Success 200 {file} binary "Raw image file"
// @Failure 404 {object} errors.ErrorResponse
// @Router /image/raw/hash/{hash} [get]
func (h *Handler) GetRawImageByHash(c *gin.Context) {
	hash := c.Param("hash")
	helpers.HandleStream(c, func(ctx context.Context) (io.ReadCloser, any, error) {
		return h.objectService.GetRawImageByHash(ctx, hash)
	}, getRawContentType)
}

// GetRawThumbnailByHash godoc
// @Summary Gets raw image thumbnail by hash
// @Description Returns the raw thumbnail image file by its hash
// @Tags images
// @Produce application/octet-stream
// @Param hash path string true "Image hash"
// @Success 200 {file} binary "Raw thumbnail image file"
// @Failure 404 {object} errors.ErrorResponse
// @Router /image/thumb/hash/{hash} [get]
func (h *Handler) GetRawThumbnailByHash(c *gin.Context) {
	hash := c.Param("hash")
	helpers.HandleStream(c, func(ctx context.Context) (io.ReadCloser, any, error) {
		return h.objectService.GetRawThumbnailByHash(ctx, hash)
	}, getRawContentType)
}

// GetImagesByQuery godoc
// @Summary Gets images by query
// @Description Returns a list of images matching the query parameters
// @Tags images
// @Produce json
// @Param prefix query string false "Filter by name prefix"
// @Param tags query string false "JSON array of tags"
// @Param username query string false "Filter by username"
// @Param user_id query uint false "Filter by user ID"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Limit results" default(20)
// @Success 200 {array} image.ImageResponse
// @Failure 400 {object} errors.ErrorResponse
// @Router /image [get]
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

// DeleteImageByName godoc
// @Summary Deletes an image by name
// @Description Deletes an image by its name (requires authentication)
// @Tags images
// @Security BearerAuth
// @Produce json
// @Param name path string true "Image name"
// @Success 204 "No Content"
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /image/{name} [delete]
func (h *Handler) DeleteImageByName(c *gin.Context) {
	name := strings.ToLower(c.Param("name"))
	helpers.WithUser(c, func(usr *user.User) error {
		return h.imageService.DeleteByName(c.Request.Context(), usr, name)
	})
}

// DeleteImagesByQuery godoc
// @Summary Deletes images by query
// @Description Deletes multiple images matching the query parameters (requires authentication)
// @Tags images
// @Security BearerAuth
// @Produce json
// @Param prefix query string false "Filter by name prefix"
// @Param tags query string false "JSON array of tags"
// @Param username query string false "Filter by username"
// @Param user_id query uint false "Filter by user ID"
// @Success 204 "No Content"
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Router /image [delete]
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

// PostImage godoc
// @Summary Uploads a single image
// @Description Uploads an image with metadata (requires authentication)
// @Tags images
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param metadata formData string true "Image metadata (JSON)" Example({"name": "my-image", "tags": ["tag1", "tag2"]})
// @Param file formData file true "Image file"
// @Success 200 {object} image.ImageResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /image/upload [post]
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

// PostImagesBatch godoc
// @Summary Uploads multiple images in batch
// @Description Uploads multiple images with metadata in a single request (requires authentication)
// @Tags images
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param metadata formData string true "Batch metadata (JSON)" Example({"data": [{"name": "img1", "tags": ["tag1"]}, {"name": "img2", "tags": ["tag2"]}], "common_tags": ["common"]})
// @Param files formData file true "Image files (multiple)"
// @Success 200 {object} image.ImagePostBatchResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /image/upload/batch [post]
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
