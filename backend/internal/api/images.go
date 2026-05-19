package api

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"server/internal/env"
	"server/internal/models"
	"server/pkg/image"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ImageResponse struct {
	ID       uint
	Filename string   `json:"filename"`
	Tags     []string `json:"tags"`
	Format   string   `json:"format"`
	Width    uint     `json:"width"`
	Height   uint     `json:"height"`
	UserID   uint     `json:"user_id"`
	Uploaded string   `json:"uploaded_at"`
}

const TimeFormat = time.RFC3339

func ConstructImageResponse(img *models.Image) ImageResponse {
	tags := models.TagsToStrings(img.Tags)

	return ImageResponse{
		ID:       img.ID,
		Filename: img.Filename,
		Tags:     tags,
		Format:   img.Format,
		Width:    img.Width,
		Height:   img.Height,
		UserID:   img.UserID,
		Uploaded: img.CreatedAt.Format(TimeFormat),
	}
}

// GetImageByName godoc
// @Summary Returns imageData by it's unique name
// @Tags Image
// @Produce  application/json
// @Param   name path string true "name"
// @Failure 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Router /image/{name} [get]
func (api *api) GetImageByName(c *gin.Context) {
	req := c.Param("name")
	img, err := api.models.Images.GetImageByName(req)

	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed getting image: " + err.Error()})
		return
	}

	response := ConstructImageResponse(img)
	c.JSON(http.StatusOK, response)
}

func (api *api) GetImageByHash(c *gin.Context) {
	req := c.Param("hash")
	img, err := api.models.Images.GetImageByHash(req)

	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed getting image: " + err.Error()})
		return
	}

	response := ConstructImageResponse(img)
	c.JSON(http.StatusOK, response)
}

func (api *api) GetImageByID(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad id"})
		return
	}

	img, err := api.models.Images.GetImageByID(id)

	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed getting image: " + err.Error()})
		return
	}

	response := ConstructImageResponse(img)
	c.JSON(http.StatusOK, response)
}

// GetRawImageByName godoc
// @Summary Returns image file by its unique name
// @Tags Image
// @Produce  application/octet-stream
// @Param   name path string true "name"
// @Failure 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Router /image/raw/{name} [get]
func (api *api) GetRawImageByName(c *gin.Context) {
	req := c.Param("name")
	img, err := api.models.Images.GetImageByName(req)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.Header("Content-Type", "image/"+img.Format)
	c.File(env.GetEnvString("UPLOADS_PATH") + "/" + img.Filename + "." + img.Format)
}

// GetImagesByQuery godoc
// @Summary Searches images by given query
// @Description Returns a list of images at "cursor + limit" matching the "query"
// @Tags Image
// @Produce  application/json
// @Param   name query string false "name" "name contains"
// @Param   tags query string false "tags" "with tags"
// @Param   cursor query int false "cursor" "cursor for pagination"
// @Param   limit query int false "limit" "limit for pagination"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /images [get]
func (api *api) GetImagesByQuery(c *gin.Context) {
	cursor, limit, err := parseCursorLimit(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	query, err := api.parseQueryFromContext(c)

	images, err := api.models.Images.SearchImages(*query, cursor, limit)
	response := make([]ImageResponse, len(images))
	for i, img := range images {
		response[i] = ConstructImageResponse(&img)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
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

	img, err := api.models.Images.GetImageByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed getting image: " + err.Error()})
		return
	}

	if err = api.models.Images.DeleteImage(img, user.ID); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed deleting image: " + err.Error()})
		return
	}

	uploadPath := env.GetEnvString("UPLOADS_PATH")
	if err := image.DeleteImagesWithName(img.Filename, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed deleting image: " + err.Error()})
		return
	}

	thumbPath := env.GetEnvString("THUMBNAILS_PATH")
	if err := image.DeleteImagesWithName(img.Filename, thumbPath); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed deleting image thumbnail: " + err.Error()})
		return
	}

	if err = api.models.Log.AddImageDeleted(img, user); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed adding log entry: " + err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "image deleted"})
}

func (api *api) DeleteImagesByQuery(c *gin.Context) {
	query, err := api.parseQueryFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("failed parsing query: %v", err)})
		return
	}

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	deletedImages, err := api.models.Images.DeleteImagesByQuery(*query, user.ID)

	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed deleting image: " + err.Error()})
		return
	}

	response := make([]ImageResponse, len(deletedImages))
	for i, img := range deletedImages {
		response[i] = ConstructImageResponse(&img)
	}

	for _, img := range deletedImages {
		uploadPath := env.GetEnvString("UPLOADS_PATH")
		if err := image.DeleteImagesWithName(img.Filename, uploadPath); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed deleting image: " + err.Error()})
			return
		}

		thumbPath := env.GetEnvString("THUMBNAILS_PATH")
		if err := image.DeleteImagesWithName(img.Filename, thumbPath); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed deleting image thumbnail: " + err.Error()})
			return
		}

		if err = api.models.Log.AddImageDeleted(&img, user); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed adding log entry: " + err.Error()})
		}
	}

	c.JSON(http.StatusOK, response)
}

type ImageMetadata struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// PostImage godoc
// @Summary Uploads image
// @Description Uploads image, name and hash must be unique
// @Tags Image
// @Accept  application/json, multipart/form-data
// @Produce  json
// @Security BearerToken
// @Failure 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /upload [post]
func (api *api) PostImage(c *gin.Context) {
	formData := c.PostForm("metadata")

	var metadata ImageMetadata
	if err := json.Unmarshal([]byte(formData), &metadata); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("invalid JSON metadata: %v", err)})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "no file received"})
		return
	}

	if err := isImageRequestValid(&metadata, fileHeader); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("%v", err)})
	}

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: fmt.Sprintf("failed to get user: %v", err)})
		return
	}

	img, err := api.ImageUploadPipeline(c, &metadata, fileHeader, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("%v", err)})
		return
	}

	response := ConstructImageResponse(img)
	c.JSON(http.StatusOK, response)
}

type ImageMetadataBatchRequest struct {
	Data       []ImageMetadata `json:"data"`
	CommonTags []string        `json:"common_tags"`
}

type ImageBatchResponse struct {
	Successes []ImageResponse    `json:"successes"`
	Failures  []image.ImageError `json:"failures"`
}

func (api *api) PostImagesBatch(c *gin.Context) {
	formData := c.PostForm("metadata")

	var batch ImageMetadataBatchRequest
	if err := json.Unmarshal([]byte(formData), &batch); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("invalid JSON metadata: %s", err.Error())})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("invalid form data: %v", err.Error())})
	}

	files := form.File["files"]
	if files == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("no files provided in form")})
	}

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: fmt.Sprintf("failed to get user: %v", err)})
		return
	}

	response := &ImageBatchResponse{}

	for i, metadata := range batch.Data {
		if i >= len(files) {
			response.Failures = append(response.Failures, image.ImageError{
				Name:  metadata.Name,
				Error: "No file provided for this metadata",
			})
			continue
		}

		if err := isImageRequestValid(&metadata, files[i]); err != nil {
			response.Failures = append(response.Failures, image.ImageError{
				Name: metadata.Name, Error: err.Error(),
			})
			continue
		}

		img, err := api.ImageUploadPipeline(c, &metadata, files[i], user)

		if err != nil {
			response.Failures = append(response.Failures, image.ImageError{
				Name: metadata.Name, Error: err.Error(),
			})
			continue
		}

		response.Successes = append(response.Successes, ConstructImageResponse(img))
	}

	c.JSON(http.StatusOK, response)
}

func isImageRequestValid(metadata *ImageMetadata, fileHeader *multipart.FileHeader) error {
	if len(metadata.Name) == 0 {
		return fmt.Errorf("empty name provided")
	}

	imgFormat, err := image.MIMETypeToFormat(fileHeader.Header.Get("Content-Type"))

	if err != nil {
		return fmt.Errorf("image format parsing error: %v", err)
	}

	//TODO fix . + format
	if err != nil || !image.IsFormatSupported(imgFormat) {
		return fmt.Errorf("unsoported image format: %s", imgFormat)
	}

	return nil
}

func (api *api) ImageUploadPipeline(c *gin.Context, metadata *ImageMetadata, fileHeader *multipart.FileHeader, user *models.User) (*models.Image, error) {
	imgWidth, imgHeight, err := image.GetDimensions(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("error getting image dimensions: %v", err)
	}

	imgFormat, err := image.MIMETypeToFormat(fileHeader.Header.Get("Content-Type"))

	if err != nil {
		return nil, fmt.Errorf("image format parsing error: %v", err)
	}

	img := api.models.Images.ConstructImage(metadata.Name, metadata.Tags, imgFormat, imgWidth, imgHeight, user.ID)

	hash, err := image.HashFile(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("error hashing the image: %v", err)
	}
	img.Hash = hash

	if err := api.models.Images.AddImage(img); err != nil {
		return nil, fmt.Errorf("error saving the image to database: %v", err)
	}

	fileHeader.Filename = img.Filename + "." + img.Format
	uploadDst := env.GetEnvString("UPLOADS_PATH") + "/" + fileHeader.Filename
	if err := c.SaveUploadedFile(fileHeader, uploadDst); err != nil {
		return nil, fmt.Errorf("error saving the image file: %v", err)
	}

	thumbDst := env.GetEnvString("THUMBNAILS_PATH") + "/" + fileHeader.Filename
	if err := image.GenerateThumbnail(uploadDst, thumbDst); err != nil {
		return nil, fmt.Errorf("error generating thumbnail: %v", err)
	}

	if err := api.models.Log.AddImageCreated(img, user); err != nil {
		return nil, fmt.Errorf("failed to log image creation: %v", err)
	}

	return img, nil
}

func (api *api) ImageDeletePipeline(c *gin.Context, img *models.Image, user *models.User) error {
	uploadPath := env.GetEnvString("UPLOADS_PATH")
	if err := image.DeleteImagesWithName(img.Filename, uploadPath); err != nil {
		return fmt.Errorf("failed deleting image: %v", err)

	}

	thumbPath := env.GetEnvString("THUMBNAILS_PATH")
	if err := image.DeleteImagesWithName(img.Filename, thumbPath); err != nil {
		return fmt.Errorf("failed deleting image thumbnail: %v", err)
	}

	if err := api.models.Log.AddImageDeleted(img, user); err != nil {
		return fmt.Errorf("failed adding log entry: %v", err)
	}

	return nil
}

func (api *api) parseQueryFromContext(c *gin.Context) (*models.ImageQuery, error) {
	query := models.EmptyQuery()

	nameContains := c.Query("name")
	tagsString := c.Query("tags")
	fromUsername := c.Query("username")
	fromUserID := c.Query("user_id")

	var err error
	var tags []string

	if nameContains != "" {
		query.NameContains = nameContains
	}
	if tagsString != "" {
		err = json.Unmarshal([]byte(tagsString), &tags)
		if err != nil {
			return nil, err
		}
		query.WithTags = tags
	}
	if fromUsername != "" {
		uploader, err := api.models.Users.GetUserByUsername(fromUsername)
		if err != nil {
			return nil, err
		}
		query.User = uploader
	}
	if fromUserID != "" {
		userID, err := strconv.ParseUint(fromUserID, 10, 64)
		if err != nil {
			return nil, err
		}

		uploader, err := api.models.Users.GetUserById(userID)
		if err != nil {
			return nil, err
		}
		query.User = uploader
	}

	return &query, nil
}
