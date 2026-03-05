package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/internal/env"
	"server/internal/models"
	"server/pkg/image"
	"strconv"
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

func (api *api) DeleteImageByName(c *gin.Context) {
	req := c.Param("name")

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	img, err := api.models.Images.GetImageByName(req)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed getting image: " + err.Error()})
		return
	}

	if err = api.models.Images.DeleteImage(img, user.ID); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "failed deleting image: " + err.Error()})
		return
	}

	uploadPath := env.GetEnvString("UPLOADS_PATH")
	if err := image.DeleteImagesWithName(req, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed deleting image: " + err.Error()})
		return
	}

	thumbPath := env.GetEnvString("THUMBNAILS_PATH")
	if err := image.DeleteImagesWithName(req, thumbPath); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed deleting image thumbnail: " + err.Error()})
		return
	}

	if err = api.models.Log.AddImageDeleted(img, user); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed adding log entry: " + err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "image deleted"})
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

	nameContains := c.Query("name")
	tagsString := c.Query("tags")
	fromUsername := c.Query("username")
	fromUserID := c.Query("user_id")
	var tags []string
	var query models.ImageQuery = models.EmptyQuery()
	if nameContains != "" {
		query.NameContains = nameContains
	}
	if tagsString != "" {
		err = json.Unmarshal([]byte(tagsString), &tags)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		query.WithTags = tags
	}
	if fromUsername != "" {
		uploader, err := api.models.Users.GetUserByUsername(fromUsername)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		query.User = uploader
	}
	if fromUserID != "" {
		userID, err := strconv.ParseUint(fromUserID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		uploader, err := api.models.Users.GetUserById(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		query.User = uploader
	}

	images, err := api.models.Images.SearchImages(query, cursor, limit)
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
	type ImageRequest struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	imageData := c.PostForm("metadata")

	var request ImageRequest
	if err := json.Unmarshal([]byte(imageData), &request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("invalid JSON metadata: %s", err.Error())})
		return
	}

	if len(request.Name) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "empty name provided"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "no file received"})
		return
	}

	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "failed to get user: " + err.Error()})
		return
	}

	imgFormat, err := image.MIMETypeToFormat(file.Header.Get("Content-Type"))

	// if err != nil || !image.IsFormatSupported(imgFormat) {
	// 	c.JSON(http.StatusBadRequest, ErrorResponse{Error: "unsoported image format: " + err.Error()})
	// 	return
	// }
	//TODO fix . + format

	imgWidth, imgHeight, err := image.GetDimensions(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("error getting image dimensions: %s", err.Error())})
		return
	}

	img := api.models.Images.ConstructImage(request.Name, request.Tags, imgFormat, imgWidth, imgHeight, user.ID)

	hash, err := image.HashFile(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("error hashing the image: %s", err.Error())})
		return
	}
	img.Hash = hash

	if err := api.models.Images.AddImage(img); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("error saving the image: %s", err.Error())})
		return
	}

	file.Filename = img.Filename + "." + img.Format
	uploadDst := env.GetEnvString("UPLOADS_PATH") + "/" + file.Filename
	if err := c.SaveUploadedFile(file, uploadDst); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("error saving the image: %s", err.Error())})
		return
	}

	thumbDst := env.GetEnvString("THUMBNAILS_PATH") + "/" + file.Filename
	if err := image.GenerateThumbnail(uploadDst, thumbDst); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("error generating thumbnail: %s", err.Error())})
		return
	}

	if err := api.models.Log.AddImageCreated(img, user); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to log image creation: " + err.Error()})
		return
	}

	response := ConstructImageResponse(img)
	c.JSON(http.StatusOK, response)
}
