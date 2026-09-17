package user

import (
	"net/http"
	"server/internal/api/helpers"
	"server/internal/errors"
	userpkg "server/internal/user"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	userService userpkg.UserService
}

func NewUserHandler(
	userService userpkg.UserService,
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

// GetUserByName godoc
// @Summary Gets user by username
// @Description Returns user profile by username
// @Tags user
// @Produce json
// @Param name path string true "Username"
// @Success 200 {object} user.UserDataResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /user/{name} [get]
func (h *Handler) GetUserByName(c *gin.Context) {
	user, err := h.userService.GetByName(c.Param("name"))
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	helpers.RespondJSON(c, http.StatusOK, NewUserDataResponse(user))
}

// GetUserByID godoc
// @Summary Gets user by ID
// @Description Returns user profile by numeric ID
// @Tags user
// @Produce json
// @Param id path uint64 true "User ID"
// @Success 200 {object} user.UserDataResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /user/id/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	id, err := helpers.ParseID(c)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	user, err := h.userService.GetByID(c, id)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	helpers.RespondJSON(c, http.StatusOK, NewUserDataResponse(user))
}

// GetMe godoc
// @Summary Gets current user profile
// @Description Returns the authenticated user's profile (requires authentication)
// @Tags user
// @Security BearerAuth
// @Produce json
// @Success 200 {object} user.UserDataResponse
// @Failure 401 {object} errors.ErrorResponse
// @Router /user/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	helpers.WithUser(c, func(u *userpkg.User) error {
		helpers.RespondJSON(c, http.StatusOK, NewUserDataResponse(u))
		return nil
	})
}
