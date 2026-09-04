package auth

import "server/internal/user"

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
} // @name AuthRequest

type LoginResponse struct {
	Token string                `json:"token"`
	User  user.UserDataResponse `json:"user"`
} // @name LoginResponse
