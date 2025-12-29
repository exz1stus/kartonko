package api

import (
	"encoding/json"
	"net/http"
	"server/internal/env"

	"fmt"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  fmt.Sprintf("%s/auth/google/callback", env.GetEnvString("API_ORIGIN")),
	ClientID:     env.GetEnvString("GOOGLE_CLIENT_ID"),
	ClientSecret: env.GetEnvString("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

func (rh *api) GoogleLoginRequest(c *gin.Context) {
	redirect := c.Query("redirect")
	state := redirect
	url := googleOauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

type UserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (rh *api) GoogleCallbackRequest(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	redirectURL := state

	token, err := googleOauthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to exchange codes"})
		return
	}

	client := googleOauthConfig.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch user info"})
		return
	}

	defer resp.Body.Close()
	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to decode user info"})
		return
	}

	user, err := rh.models.Users.GetOrRegisterGoogle(userInfo.Name, userInfo.Email, userInfo.ID, userInfo.Picture)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	tokenString, err := GenerateJwtToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Path:     "/",
		Domain:   env.GetEnvString("API_ORIGIN"),
		MaxAge:   int(JWT_COOKIE_MAX_AGE.Seconds()),
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteNoneMode,
	})

	if redirectURL != "" {
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	c.JSON(http.StatusOK, gin.H{"register": true})
}
