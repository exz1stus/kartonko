package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/internal/env"
	"server/internal/models"
	"server/pkg/image"

	"github.com/gin-gonic/gin"
)

// GetImageByNameRequest godoc
// @Summary Returns imageData by it's unique name
// @Tags Image
// @Produce  application/json
// @Param   name query string true "name"
// @Failure 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Router /image/{name} [get]
func (api *api) GetImageByNameRequest(c *gin.Context) {
	req := c.Query("name")

	img, err := api.models.Images.GetImageByName(req)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"imageData": img.Filename, "tags": img.Tags})
}

// GetImagesByQueryRequest godoc
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
func (api *api) GetImagesByQueryRequest(c *gin.Context) {
	cursor, limit, err := parseCursorLimit(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	nameContains := c.Query("name")
	tagsString := c.Query("tags")
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

	images, err := api.models.Images.SearchImages(query, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"imageData": images})
}

// GetRawImageByNameRequest godoc
// @Summary Returns image file by its unique name
// @Tags Image
// @Produce  application/octet-stream
// @Param   name query string true "name"
// @Failure 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Router /raw-image/{name} [get]
func (api *api) GetRawImageByNameRequest(c *gin.Context) {
	req := c.Param("name")
	img, err := api.models.Images.GetImageByName(req)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	c.File(env.GetEnvString("UPLOADS_PATH") + "/" + img.Filename + "." + img.Format)
}

// PostImageRequest godoc
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
func (api *api) PostImageRequest(c *gin.Context) {
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	imgFormat, err := image.MIMETypeToFormat(file.Header.Get("Content-Type"))

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	imgWidth, imgHeight, err := image.GetDimensions(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("error getting image dimensions: %s", err.Error())})
		return
	}

	img := api.models.Images.ConstructImage(request.Name, request.Tags, imgFormat, imgWidth, imgHeight)

	if err := api.models.Images.SaveImage(img, file, c); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("error saving the image: %s", err.Error())})
		return
	}

	err = api.models.Log.OnImageCreated(img, user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload successful",
		"img":     img,
	})
}
