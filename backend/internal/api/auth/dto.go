package auth

import "server/internal/user"

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string                `json:"token"`
	User  user.UserDataResponse `json:"user"`
}
