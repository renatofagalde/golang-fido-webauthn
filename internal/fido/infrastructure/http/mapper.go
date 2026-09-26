package http

import (
	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/application"
	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/domain"
)

func toCreateUserInput(request CreateUserRequest) application.CreateUserInput {
	return application.CreateUserInput{Username: request.Username, Displayname: request.Displayname}
}

func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		Hash:        u.Hash,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
	}
}
