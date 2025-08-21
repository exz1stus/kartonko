package main

import (
	"fmt"
	"net/http"
	"server/internal/env"
	"server/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (rh *RequestHandler) GetUserFromContext(c *gin.Context) (*models.User, error) {
	tokenString, err := c.Cookie("jwt")
	if err != nil {
		return nil, fmt.Errorf("failed to get jwt token: %v", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(env.GetEnvString("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userId, ok := claims["userId"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	user, err := rh.models.Users.GetUserById(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

// LoginRequest godoc
// @Summary Login a user
// @Description Logs in a user, generating a JWT token.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   body body authRequest true "username and password"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (rh *RequestHandler) LoginRequest(c *gin.Context) {
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

	c.SetCookie("jwt", tokenString, 3600, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"login": true})
}

// RegisterRequest godoc
// @Summary Register a user
// @Description Registers a new user, generating a JWT token.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   body body authRequest true "username and password"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
func (rh *RequestHandler) RegisterRequest(c *gin.Context) {
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

	c.SetCookie("jwt", tokenString, 3600, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"register": true})
}

// LogoutRequest godoc
// @Summary Logout a user
// @Description Logs out a user, deleting the JWT token.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/logout [post]
func (rh *RequestHandler) LogoutRequest(c *gin.Context) {
	c.SetCookie("jwt", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"logout": true})
}

func GenerateJwtToken(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(env.GetEnvString("JWT_SECRET")))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	return tokenString, nil
}
