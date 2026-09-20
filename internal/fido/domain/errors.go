package domain

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUsernametaken = errors.New("username already taken")
)
