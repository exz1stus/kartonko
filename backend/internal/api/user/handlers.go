package user

import (
	"net/http"
	"server/internal/api/helpers"
	"server/internal/errors"
	"server/internal/user"
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
	public.GET("/:id", h.GetUser)
	public.GET("/name/:name", h.GetUserByName)

	protected.PATCH("/:id", h.PatchUser)
	protected.DELETE("/:id", h.DeleteUser)

	protected.GET("/me", h.GetMe)
}

// GetUser godoc
// @Summary Gets user by ID
// @Description Returns user profile by numeric ID
// @Tags user
// @Produce json
// @Param id path uint64 true "User ID"
// @Success 200 {object} user.UserDataResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /user/{id} [get]
// @ID GetUser
func (h *Handler) GetUser(c *gin.Context) {
	id, err := helpers.ParseID(c)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	user, err := h.userService.Get(c, id)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	helpers.RespondJSON(c, http.StatusOK, NewUserDataResponse(user))
}

// GetUserByName godoc
// @Summary Gets user by username
// @Description Returns user profile by username
// @Tags user
// @Produce json
// @Param name path string true "Username"
// @Success 200 {object} user.UserDataResponse
// @Failure 404 {object} errors.ErrorResponse
// @Router /user/name/{name} [get]
// @ID GetUserByName
func (h *Handler) GetUserByName(c *gin.Context) {
	user, err := h.userService.GetByUsername(c, c.Param("name"))
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
// @ID GetMe
func (h *Handler) GetMe(c *gin.Context) {
	helpers.WithUser(c, func(u *userpkg.User) error {
		helpers.RespondJSON(c, http.StatusOK, NewUserDataResponse(u))
		return nil
	})
}

// PatchUser godoc
// @Summary Patches existing user by id
// @Tags user
// @Produce json
// @Param id path uint64 true "User ID"
// @Param request body UserPatchRequest true "User patch request"
// @Success 200 {object} UserDataResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /user/{id} [patch]
// @ID PatchUser
func (h *Handler) PatchUser(c *gin.Context) {
	helpers.WithUser(c, func(user *user.User) error {
		id, err := helpers.ParseID(c)
		if err != nil {
			return err
		}

		var req userpkg.UserPatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return errors.WrapBadRequest(err)
		}

		tag, err := h.userService.Update(c, user.ID, id, &req)
		if err != nil {
			return err
		}

		helpers.RespondJSON(c, http.StatusOK, NewUserDataResponse(tag))
		return nil
	})
}

// DeleteUser godoc
// @Summary Delete user by id
// @Tags user
// @Produce json
// @Param id path uint64 true "User ID"
// @Success 204
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /user/{id} [delete]
// @ID DeleteUser
func (h *Handler) DeleteUser(c *gin.Context) {
	helpers.WithUser(c, func(user *user.User) error {
		id, err := helpers.ParseID(c)
		if err != nil {
			return err
		}

		if err := h.userService.Delete(c, user.ID, id); err != nil {
			return err
		}
		c.Status(http.StatusNoContent)
		return nil
	})
}
