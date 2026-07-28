package api

import (
	"net/http"
	"server/internal/api/dto"
	"server/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func constructUserResponse(user *models.User) dto.UserDataResponse {
	res := dto.UserDataResponse{
		ID:        user.ID,
		Username:  user.Username,
		Privilege: user.Privilege.String(),
		JoinedAt:  user.CreatedAt.Format(time.DateOnly),
		LastSeen:  user.LastSeen.Format(time.DateTime),
	}

	if user.IsOauth() {
		res.PictureURL = user.PictureURL
	}

	return res
}

func (api *api) GetUserByName(c *gin.Context) {
	username := c.Param("name")
	user, err := api.userService.GetByUsername(username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, constructUserResponse(user))
}

func (api *api) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad id"})
		return
	}

	user, err := api.userService.GetByID(uint(id))

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, constructUserResponse(user))
}

func (api *api) GetMe(c *gin.Context) {
	user, err := api.GetUserFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	res := constructUserResponse(user)
	c.JSON(http.StatusOK, res)
}
