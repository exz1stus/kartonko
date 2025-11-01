package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/image"
	"server/internal/env"
	"server/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// swagger:model
type ErrorResponse struct {
	Error string `json:"error"`
}

type RequestHandler struct {
	models *models.Models
}

func MustInitReqHandler(models *models.Models) *RequestHandler {
	if models == nil {
		panic("storage is nil")
	}

	rh := RequestHandler{models: models}
	rh.models = models
	return &rh
}

// @Summary Health check
// @Description Returns "ok" if server is up and running
// @Tags HealthCheck
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} ErrorResponse
// @Router /health [get]
func (rh *RequestHandler) GetHealthCheckRequest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// @Summary Searches tags by given query
// @Description Returns a list of tags at "cursor + limit" matching the "query"
// @Tags Tags
// @Produce  json
// @Param   query query string true "query"
// @Param   limit query int true "limit"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tags [get]
func (rh *RequestHandler) GetTagsRequest(c *gin.Context) {
	query := c.Query("query")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "failed to parse limit"})
		return
	}
	tags, err := rh.models.Tags.SearchTags(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// GetImageByNameRequest godoc
// @Summary Returns imageData by it's unique name
// @Tags Image
// @Produce  application/json
// @Param   name query string true "name"
// @Failure 200 {object} map[string]interface{}
// @Failure 404 {object} ErrorResponse
// @Router /image/{name} [get]
func (rh *RequestHandler) GetImageByNameRequest(c *gin.Context) {
	req := c.Query("name")

	img, err := rh.models.Images.GetImageByName(req)
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
func (rh *RequestHandler) GetImagesByQueryRequest(c *gin.Context) {
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

	images, err := rh.models.Images.SearchImages(query, cursor, limit)
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
func (rh *RequestHandler) GetRawImageByNameRequest(c *gin.Context) {
	req := c.Param("name")
	img, err := rh.models.Images.GetImageByName(req)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}
	c.File(env.GetEnvString("UPLOADS_PATH") + "/" + img.Filename + "." + img.Format)
}

// GetAuditLogEntriesRequest godoc
// @Summary Gets audit log entries by given range
// @Description Returns a list of log entries at "cursor + limit" matching the "query"
// @Tags AuditLog
// @Produce  json
// @Param   cursor query string true "cursor"
// @Param   limit query string true "limit"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /log [get]
func (rh *RequestHandler) GetAuditLogEntriesRequest(c *gin.Context) {
	cursor, limit, err := parseCursorLimit(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	entries, err := rh.models.Log.GetEntries(cursor, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

type UserDataResponce struct {
	Username   string `json:"username"`
	Privileage string `json:"privileage"`
	PictureURL string `json:"picture_url"`
	JoinedAt   string `json:"joined_at"`
	LastSeen   string `json:"last_seen"`
	Online     bool   `json:"online"`
}

// GetUserRequest godoc
// @Summary Returns user's data by username
// @Tags User
// @Produce  json
// @Param   username path string true "username"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /user/{username} [get]
func (rh *RequestHandler) GetUserRequest(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "empty username provided"})
	}

	user, err := rh.models.Users.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	res := UserDataResponce{
		Username:   user.Username,
		Privileage: user.Privileage.String(),
		JoinedAt:   user.CreatedAt.Format(time.DateOnly),
		LastSeen:   user.LastSeen.Format(time.DateTime),
	}

	if user.IsOauth() {
		res.PictureURL = user.PictureURL
	}

	c.JSON(http.StatusOK, res)
}

func (rh *RequestHandler) GetMeRequest(c *gin.Context) {
	user, err := rh.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "user/"+user.Username)
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
func (rh *RequestHandler) PostImageRequest(c *gin.Context) {
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

	imgFormat, err := image.MIMETypeToFormat(file.Header.Get("Content-Type"))

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	img := rh.models.Images.ConstructImage(request.Name, request.Tags, imgFormat)

	if err := rh.models.Images.SaveImage(img, file, c); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("error saving the image: %s", err.Error())})
		return
	}

	user, err := rh.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	err = rh.models.Log.OnImageCreated(img, user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Upload successful",
		"img":     img,
	})
}

func parseCursorLimit(c *gin.Context) (int, int, error) {
	cursorStr := c.Query("cursor")
	limitStr := c.Query("limit")

	if cursorStr == "" || limitStr == "" {
		return 0, 0, fmt.Errorf("no cursor or limit parameter provided")
	}

	cursor, err := strconv.Atoi(cursorStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid cursor parameter: %w", err)
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit parameter: %w", err)
	}

	return cursor, limit, nil
}
