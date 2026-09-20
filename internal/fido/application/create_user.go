// Package application contém os casos de uso (use cases) da aplicação.
package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/domain"
)

func (uc *userUsecase) Create(ctx context.Context, input CreateUserInput) (*domain.User, error) {
	if input.Username == "" || input.Displayname == "" {
		return nil, domain.ErrInvalidInput
	}

	taken, err := uc.repository.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, domain.ErrUsernametaken
	}
	hash, _ := uuid.NewV7()
	user := &domain.User{
		Hash:        hash.String(),
		Username:    input.Username,
		DisplayName: input.Displayname,
		IsActive:    true,
	}

	if err := uc.repository.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
