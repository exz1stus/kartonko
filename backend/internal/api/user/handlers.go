package user

import (
	"bytes"
	"fmt"
	stdimage "image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"server/internal/api/helpers"
	"server/internal/errors"
	"server/internal/storage"
	userpkg "server/internal/user"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	userService userpkg.UserService
	storage     storage.Storage
}

func NewUserHandler(
	userService userpkg.UserService,
	objectStorage storage.Storage,
) *Handler {
	return &Handler{
		userService: userService,
		storage:     objectStorage,
	}
}

func (h *Handler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.GET("/:id", h.GetUser)
	public.GET("/:id/picture", h.GetPicture)
	public.GET("/name/:name", h.GetUserByName)

	protected.PATCH("/:id", h.PatchUser)
	protected.POST("/:id/picture", h.UploadPicture)
	protected.DELETE("/:id", h.DeleteUser)

	protected.GET("/me", h.GetMe)
}

const maxProfilePictureBytes = 5 << 20

func profilePictureKey(id uint) string { return fmt.Sprintf("profile/%d/picture", id) }
func profilePictureURL(id uint) string { return fmt.Sprintf("/apilocal/user/%d/picture", id) }

// UploadPicture stores a profile picture in object storage and updates the profile URL.
func (h *Handler) UploadPicture(c *gin.Context) {
	helpers.WithUser(c, func(actor *userpkg.User) error {
		id, err := helpers.ParseID(c)
		if err != nil {
			return err
		}
		if !userpkg.CanEdit(actor.ID, actor.Privilege, id) {
			return errors.ErrPermissionDenied
		}
		if c.Request.ContentLength > maxProfilePictureBytes+(1<<16) {
			return errors.ErrBadRequest
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxProfilePictureBytes+(1<<16))
		file, header, err := c.Request.FormFile("picture")
		if err != nil {
			return errors.WrapBadRequest(err)
		}
		defer file.Close()
		if header.Size > maxProfilePictureBytes {
			return errors.ErrBadRequest
		}
		data, err := io.ReadAll(io.LimitReader(file, maxProfilePictureBytes+1))
		if err != nil {
			return errors.WrapBadRequest(err)
		}
		if len(data) == 0 || len(data) > maxProfilePictureBytes {
			return errors.ErrBadRequest
		}
		config, format, err := stdimage.DecodeConfig(bytes.NewReader(data))
		if err != nil || config.Width < 1 || config.Height < 1 || config.Width > 8192 || config.Height > 8192 || int64(config.Width)*int64(config.Height) > 40_000_000 {
			return errors.ErrUnsupportedFormat
		}
		contentType := map[string]string{"jpeg": "image/jpeg", "png": "image/png", "gif": "image/gif"}[format]
		if contentType == "" {
			return errors.ErrUnsupportedFormat
		}
		if err := h.storage.Upload(c, profilePictureKey(id), bytes.NewReader(data), contentType); err != nil {
			return err
		}
		pictureURL := profilePictureURL(id)
		updated, err := h.userService.Update(c, actor.ID, id, &userpkg.UserPatchRequest{PictureURL: &pictureURL})
		if err != nil {
			return err
		}
		helpers.RespondJSON(c, http.StatusOK, NewUserDataResponse(updated))
		return nil
	})
}

// GetPicture streams the stored profile picture.
func (h *Handler) GetPicture(c *gin.Context) {
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
	if user.PictureURL != profilePictureURL(id) {
		errors.RespondError(c, errors.ErrNotFound)
		return
	}
	info, err := h.storage.Stat(c, profilePictureKey(id))
	if err != nil {
		errors.RespondError(c, errors.ErrNotFound)
		return
	}
	body, err := h.storage.Download(c, profilePictureKey(id))
	if err != nil {
		errors.RespondError(c, errors.ErrNotFound)
		return
	}
	defer body.Close()
	c.Header("Cache-Control", "public, max-age=86400, stale-while-revalidate=604800")
	contentType := info.ContentType
	if !strings.HasPrefix(contentType, "image/") {
		contentType = "application/octet-stream"
	}
	c.DataFromReader(http.StatusOK, info.Size, contentType, body, nil)
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
	helpers.WithUser(c, func(user *userpkg.User) error {
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
	helpers.WithUser(c, func(user *userpkg.User) error {
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
