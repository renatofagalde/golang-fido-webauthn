package http

import "time"

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
}

type UserResponse struct {
	Hash        string    `json:"hash"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
