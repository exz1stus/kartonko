package user

import (
	"net/http"
	"server/internal/api/helpers"
	"server/internal/errors"
	"server/internal/user"
	userpkg "server/internal/user"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	userService userpkg.UserService
}

func NewUserHandler(
	userService user.UserService,
) *Handler {
	return &Handler{
		userService: userService,
	}
}

func (h *Handler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.GET("/:name", h.GetUserByName)
	public.GET("/id/:id", h.GetUserByID)

	protected.GET("/me", h.GetMe)
}

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

func (h *Handler) GetUserByName(c *gin.Context) {
	helpers.HandleGet(c, func() (*userpkg.User, error) {
		return h.userService.GetByUsername(c.Param("name"))
	}, func(u *userpkg.User) any { return NewUserResponse(u) })
}

func (h *Handler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		errors.RespondError(c, errors.ErrBadRequest)
		return
	}
	helpers.HandleGet(c, func() (*userpkg.User, error) {
		return h.userService.GetByID(uint(id64))
	}, func(u *userpkg.User) any { return NewUserResponse(u) })
}

func (h *Handler) GetMe(c *gin.Context) {
	helpers.WithUser(c, func(u *userpkg.User) error {
		helpers.RespondJSON(c, http.StatusOK, NewUserResponse(u))
		return nil
	})
}
