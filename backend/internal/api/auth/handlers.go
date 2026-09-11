package auth

import (
	"fmt"
	"net/http"
	"server/internal/env"
	"server/internal/errors"
	"server/internal/user"
	userpkg "server/internal/user"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtCookieMaxAge     time.Duration
	jwtCookieMaxAgeOnce sync.Once
)

type Handler struct {
	userService user.UserService
	jwtSecret   string
}

func NewAuthHandler(
	jwtSecret string,
	userService user.UserService,
) *Handler {
	return &Handler{
		userService: userService,
		jwtSecret:   jwtSecret,
	}
}

func (h *Handler) RegisterRoutes(public *gin.RouterGroup, protected *gin.RouterGroup) {
	public.POST("/login", h.PostLogin)
	public.POST("/register", h.PostRegister)
	public.POST("/logout", h.PostLogout)

	public.GET("/google", h.GetGoogleLogin)
	public.GET("/google/callback", h.GetGoogleCallback)
}

func GetUserFromContext(c *gin.Context) (*user.User, error) {
	userInter, exists := c.Get("user")
	if !exists {
		return nil, fmt.Errorf("user is not passed in context")
	}

	user, ok := userInter.(*user.User)
	if !ok {
		return nil, fmt.Errorf("context user is not of type *user.User")
	}

	return user, nil
}

// PostLogin godoc
// @Summary Login a user
// @Description Logs in a user, generating a JWT token.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   body body AuthRequest true "username and password"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) PostLogin(c *gin.Context) {
	var input AuthRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{Error: fmt.Sprint("invalid input: ", err.Error())})
		return
	}

	user, err := h.userService.GetByUsername(input.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{Error: "invalid username"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password)); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{Error: "invalid password"})
		return
	}

	tokenString, err := GenerateJwtToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{Error: "failed to generate token"})
		return
	}

	setTokenCookie(tokenString, &c.Writer)

	res := &LoginResponse{
		Token: tokenString,
		User:  userpkg.NewUserData(user),
	}

	c.JSON(http.StatusOK, res)
}

func GetJWTCookieMaxAge() time.Duration {
	jwtCookieMaxAgeOnce.Do(func() {
		jwtCookieMaxAge = time.Duration(env.GetEnvInt("JWT_COOKIE_MAX_AGE_HOURS")) * time.Hour
	})
	return jwtCookieMaxAge
}

// PostRegister godoc
// @Summary Register a user
// @Description Registers a new user, generating a JWT token.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   body body AuthRequest true "username and password"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /auth/register [post]
func (h *Handler) PostRegister(c *gin.Context) {
	var input AuthRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{Error: "Invalid input"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{Error: "failed to hash password"})
		return
	}

	user, err := h.userService.CreateByRegistration(input.Username, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{Error: fmt.Sprint("failed to create user: ", err.Error())})
		return
	}

	tokenString, err := GenerateJwtToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{Error: "failed to generate token"})
		return
	}

	setTokenCookie(tokenString, &c.Writer)

	res := &LoginResponse{
		Token: tokenString,
		User:  userpkg.NewUserData(user),
	}

	c.JSON(http.StatusOK, res)
}

// PostLogout godoc
// @Summary Logout a user
// @Description Logs out a user, deleting the JWT token.
// @Tags auth
// @Accept  json
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *Handler) PostLogout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		Domain:   env.GetEnvString("DOMAIN"),
		SameSite: http.SameSiteDefaultMode,
	})

	c.Status(http.StatusOK)
}

func GenerateJwtToken(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(GetJWTCookieMaxAge()).Unix(),
	})

	tokenString, err := token.SignedString([]byte(env.GetEnvString("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	return tokenString, nil
}

func setTokenCookie(token string, writer *gin.ResponseWriter) {
	http.SetCookie(*writer, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		MaxAge:   int(GetJWTCookieMaxAge().Seconds()),
		Domain:   env.GetEnvString("DOMAIN"),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	})
}
