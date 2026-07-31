package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"server/internal/api/dto"
	"server/internal/errors"
	"server/internal/models"
	"server/pkg/image"
	"server/pkg/tag"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const TimeFormat = time.RFC3339

func ConstructImageResponse(img *models.ImageMetadata) dto.ImageResponse {
	tags := models.TagsToStrings(img.Tags)

	return dto.ImageResponse{
		ID:       img.ID,
		Hash:     img.Hash,
		Filename: img.Filename,
		Tags:     tags,
		Format:   img.Format,
		Width:    img.Width,
		Height:   img.Height,
		UserID:   img.UserID,
		Uploaded: img.CreatedAt.Format(TimeFormat),
	}
}

func (api *api) GetImageByName(c *gin.Context) {
	req := c.Param("name")
	img, err := api.imageService.GetByName(req)

	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed getting image: " + err.Error()})
		return
	}

	response := ConstructImageResponse(img)
	c.JSON(http.StatusOK, response)
}

func (api *api) GetImageByHash(c *gin.Context) {
	req := c.Param("hash")
	img, err := api.imageService.GetByHash(req)

	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed getting image: " + err.Error()})
		return
	}

	response := ConstructImageResponse(img)
	c.JSON(http.StatusOK, response)
}

func (api *api) GetImageByID(c *gin.Context) {
	idStr := c.Param("id")

	id64, err := strconv.ParseUint(idStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad id"})
		return
	}

	id := uint(id64)
	img, err := api.imageService.GetByID(id)

	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed getting image: " + err.Error()})
		return
	}

	response := ConstructImageResponse(img)
	c.JSON(http.StatusOK, response)
}

func (api *api) GetRawImageByName(c *gin.Context) {
	api.streamObject(c, func(img *models.ImageMetadata) (string, string) {
		format, _ := image.ParseFormat(img.Format)
		return image.ImageKey(img.Hash, format), format.MIMEType()
	})
}

func (api *api) GetRawThumbnailByName(c *gin.Context) {
	api.streamObject(c, func(img *models.ImageMetadata) (string, string) {
		format, _ := image.ParseFormat(img.Format)
		return image.ThumbnailKey(img.Hash, format), format.MIMEType()
	})
}

// streamObject resolves an image by name and streams the object at the key
// returned by keyFn through the response writer. It is storage-agnostic: the
// key derivation is delegated to the caller, and bytes flow straight from
// storage to the client with no intermediate buffering.
func (api *api) streamObject(c *gin.Context, keyFn func(*models.ImageMetadata) (string, string)) {
	name := c.Param("name")
	img, err := api.imageService.GetByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	key, defaultContentType := keyFn(img)
	body, err := api.storage.Download(c, key)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	defer body.Close()

	contentType := defaultContentType
	if info, err := api.storage.Stat(c, key); err == nil && info.ContentType != "" {
		contentType = info.ContentType
	}
	c.Header("Content-Type", contentType)
	if _, err := io.Copy(c.Writer, body); err != nil {
		// client likely disconnected; nothing else to do
	}
}

func (api *api) GetImagesByQuery(c *gin.Context) {
	query, err := api.newQueryFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	images, err := api.imageService.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	response := make([]dto.ImageResponse, len(images))
	for i, img := range images {
		response[i] = ConstructImageResponse(&img)
	}

	c.JSON(http.StatusOK, response)
}

func (api *api) DeleteImageByName(c *gin.Context) {
	name := c.Param("name")
	name = strings.ToLower(name)

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	if err := api.imageService.DeleteByName(c.Request.Context(), user, name); err != nil {
		if err == errors.ErrPermissionDenied {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "permission denied: " + err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed getting image: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "image deleted"})
}

func (api *api) DeleteImagesByQuery(c *gin.Context) {
	query, err := api.newQueryFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("failed parsing query: %v", err)})
		return
	}

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	errs := api.imageService.DeleteByQuery(c.Request.Context(), user, query)
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
			c.JSON(http.StatusForbidden, gin.H{
				"errors": messages,
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"errors": messages,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "images deleted"})
}

func (api *api) PostImage(c *gin.Context) {
	formData := c.PostForm("metadata")

	var postRequest dto.ImagePostRequest
	if err := json.Unmarshal([]byte(formData), &postRequest); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("invalid JSON metadata: %v", err)})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "no file received"})
		return
	}

	if err := isImageRequestValid(&postRequest, fileHeader); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: fmt.Sprintf("failed to get user: %v", err)})
		return
	}

	img, err := api.imageService.Upload(c.Request.Context(), user, &postRequest, fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	response := ConstructImageResponse(img)
	c.JSON(http.StatusOK, response)
}

func (api *api) PostImagesBatch(c *gin.Context) {
	formData := c.PostForm("metadata")

	var batch dto.ImagePostBatchRequest
	if err := json.Unmarshal([]byte(formData), &batch); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("invalid JSON metadata: %s", err.Error())})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("invalid form data: %v", err)})
		return
	}

	files := form.File["files"]
	if files == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "no files provided in form"})
		return
	}

	if len(batch.Data) != len(files) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("files: %d and metadata: %d quantities are different", len(files), len(batch.Data))})
	}

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: fmt.Sprintf("failed to get user: %v", err)})
		return
	}

	response := &dto.ImagePostBatchResponse{}

	for i, metadata := range batch.Data {
		if i >= len(files) {
			response.Failures = append(response.Failures, dto.ImageError{
				Name:  metadata.Name,
				Error: "No file provided for this metadata",
			})
			continue
		}

		if err := isImageRequestValid(&metadata, files[i]); err != nil {
			response.Failures = append(response.Failures, dto.ImageError{
				Name: metadata.Name, Error: err.Error(),
			})
			continue
		}

		img, err := api.imageService.Upload(c.Request.Context(), user, &metadata, files[i])
		if err != nil {
			response.Failures = append(response.Failures, dto.ImageError{
				Name: metadata.Name, Error: err.Error(),
			})
			continue
		}

		response.Successes = append(response.Successes, ConstructImageResponse(img))
	}

	if len(response.Failures) > 0 {
		c.JSON(http.StatusMultiStatus, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

func isImageRequestValid(metadata *dto.ImagePostRequest, fileHeader *multipart.FileHeader) error {
	if len(metadata.Name) == 0 {
		return fmt.Errorf("empty name provided")
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

func (api *api) newQueryFromContext(c *gin.Context) (*models.ImageQuery, error) {
	cursor, limit, err := parseCursorLimit(c)
	if err != nil {
		return nil, err
	}

	prefix := c.Query("prefix")
	tagsString := c.Query("tags")
	username := c.Query("username")
	userIDStr := c.Query("user_id")

	builder := models.NewImageQueryBuilder().
		Prefix(prefix).
		Cursor(cursor).
		Limit(limit)

	if tagsString != "" {
		tags, err := tag.ParseTagsFromJSONString(tagsString)
		if err != nil {
			return nil, err
		}
		builder.Tags(tags)
	}

	if userIDStr != "" {
		userID64, err := strconv.ParseUint(userIDStr, 10, 64)
		userID := uint(userID64)
		if err != nil {
			return nil, err
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
