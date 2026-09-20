package application

import (
	"context"

	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/domain"
)

type UserUsecase interface {
	// o handler http, depende desta interface, nao da implementacao
	Create(ctx context.Context, input CreateUserInput) (*domain.User, error)
}

type userUsecase struct {
	repository domain.UserRepository
}

func NewUserService(repository domain.UserRepository) UserUsecase {
	return &userUsecase{repository: repository}
}
