package api

import (
	"net/http"
	"server/internal/errors"
	userpkg "server/internal/user"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func NewUserResponse(u *userpkg.User) userpkg.UserDataResponse {
	res := userpkg.UserDataResponse{
		ID:        u.ID,
		Username:  u.Username,
		Privilege: u.Privilege.String(),
		JoinedAt:  u.CreatedAt.Format(time.DateOnly),
		LastSeen:  u.LastSeen.Format(time.DateTime),
	}

	if u.IsOauth() {
		res.PictureURL = u.PictureURL
	}

	return res
}

func (api *api) GetUserByName(c *gin.Context) {
	HandleGet(c, func() (*userpkg.User, error) {
		return api.userService.GetByUsername(c.Param("name"))
	}, func(u *userpkg.User) any { return NewUserResponse(u) })
}

func (api *api) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		RespondError(c, errors.ErrBadRequest)
		return
	}
	HandleGet(c, func() (*userpkg.User, error) {
		return api.userService.GetByID(uint(id64))
	}, func(u *userpkg.User) any { return NewUserResponse(u) })
}

func (api *api) GetMe(c *gin.Context) {
	WithUser(c, api, func(u *userpkg.User) error {
		RespondJSON(c, http.StatusOK, NewUserResponse(u))
		return nil
	})
}
