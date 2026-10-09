package auth

import "server/internal/user"

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
} // @name AuthRequest

type LoginResponse struct {
	Token string            `json:"token"`
	User  user.UserResponce `json:"user"`
} // @name LoginResponse
