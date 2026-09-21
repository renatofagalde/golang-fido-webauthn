package http

import "time"

type CreateUserRequest struct {
	Username    string `json:"user_name" binding: "requerid"`
	Displayname string `json:"display_name" binding: "requerid"`
}

type UserResponse struct {
	Hash        string    `json:"hash"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
