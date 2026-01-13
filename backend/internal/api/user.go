package api

import (
	"net/http"
	"server/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

type UserDataResponce struct {
	Username   string `json:"username"`
	Privileage string `json:"privileage"`
	PictureURL string `json:"picture_url"`
	JoinedAt   string `json:"joined_at"`
	LastSeen   string `json:"last_seen"`
	Online     bool   `json:"online"`
}

func buildUserResponce(user *models.User) UserDataResponce {
	res := UserDataResponce{
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
// @Param   username path string true "username"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /user/{username} [get]
func (api *api) GetUserRequest(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "empty username provided"})
	}

	user, err := api.models.Users.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	res := buildUserResponce(user)

	c.JSON(http.StatusOK, res)
}

func (api *api) GetMeRequest(c *gin.Context) {
	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	res := buildUserResponce(user)
	c.JSON(http.StatusOK, res)
}
