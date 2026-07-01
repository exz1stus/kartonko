package api

import (
	"fmt"
	"net/http"
	"server/internal/env"
	"server/internal/models"
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

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (rh *api) GetUserFromContext(c *gin.Context) (*models.User, error) {
	userInter, exists := c.Get("user")
	if !exists {
		return nil, fmt.Errorf("user is not passed in context")
	}

	user, ok := userInter.(*models.User)
	if !ok {
		return nil, fmt.Errorf("context user is not of type *models.User")
	}

	return user, nil
}

// PostLogin godoc
// @Summary Login a user
// @Description Logs in a user, generating a JWT token.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   body body authRequest true "username and password"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (rh *api) PostLogin(c *gin.Context) {
	var input authRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprint("invalid input: ", err.Error())})
		return
	}

	user, err := rh.models.Users.GetUserByUsername(input.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid username"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(input.Password)); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid password"})
		return
	}

	tokenString, err := GenerateJwtToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
		return
	}

	setTokenCookie(tokenString, &c.Writer)

	res := &LoginResponse{
		Token: tokenString,
		User:  *user,
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
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   body body authRequest true "username and password"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (rh *api) PostRegister(c *gin.Context) {
	var input authRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid input"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to hash password"})
		return
	}

	user, err := rh.models.Users.CreateUserRegistration(input.Username, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprint("failed to create user: ", err.Error())})
		return
	}

	tokenString, err := GenerateJwtToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
		return
	}

	setTokenCookie(tokenString, &c.Writer)

	c.JSON(http.StatusOK, gin.H{"register": true})
}

// PostLogout godoc
// @Summary Logout a user
// @Description Logs out a user, deleting the JWT token.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func (rh *api) PostLogout(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{"logout": true})
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
