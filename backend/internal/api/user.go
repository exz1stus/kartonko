package api

import (
	"net/http"
	"server/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UserDataResponce struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Privileage string `json:"privileage"`
	PictureURL string `json:"picture_url"`
	JoinedAt   string `json:"joined_at"`
	LastSeen   string `json:"last_seen"`
	Online     bool   `json:"online"`
}

func constructUserResponce(user *models.User) UserDataResponce {
	res := UserDataResponce{
		ID:         user.ID,
		Username:   user.Username,
		Privileage: user.Privileage.String(),
		JoinedAt:   user.CreatedAt.Format(time.DateOnly),
		LastSeen:   user.LastSeen.Format(time.DateTime),
	}

	if user.IsOauth() {
		res.PictureURL = user.PictureURL
	}

	return res
}

// GetUserRequest godoc
// @Summary Returns user's data by username
// @Tags User
// @Produce  json
// @Query   username path string true "username"
// @Query   id path number true "id"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /user [get]
func (api *api) GetUserRequest(c *gin.Context) {
	username := c.Query("username")
	idStr := c.Query("id")

	if username == "" && idStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "username or id must be provided"})
		return
	}

	var user *models.User
	var err error

	if username != "" {
		user, err = api.models.Users.GetUserByUsername(username)
	} else {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad id provided"})
			return
		}
		user, err = api.models.Users.GetUserById(id)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, constructUserResponce(user))
}

func (api *api) GetMeRequest(c *gin.Context) {
	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	res := constructUserResponce(user)
	c.JSON(http.StatusOK, res)
}
