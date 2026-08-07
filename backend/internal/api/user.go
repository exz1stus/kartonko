package api

import (
	"net/http"
	"server/internal/api/dto"
	"server/internal/errors"
	"server/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func NewUserResponse(user *models.User) dto.UserDataResponse {
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
	HandleGet(c, func() (*models.User, error) {
		return api.userService.GetByUsername(c.Param("name"))
	}, func(u *models.User) any { return NewUserResponse(u) })
}

func (api *api) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		RespondError(c, errors.ErrBadRequest)
		return
	}
	HandleGet(c, func() (*models.User, error) {
		return api.userService.GetByID(uint(id64))
	}, func(u *models.User) any { return NewUserResponse(u) })
}

func (api *api) GetMe(c *gin.Context) {
	WithUser(c, api, func(user *models.User) error {
		RespondJSON(c, http.StatusOK, NewUserResponse(user))
		return nil
	})
}
