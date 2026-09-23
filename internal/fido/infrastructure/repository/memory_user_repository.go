// package repository teremos os acessos aos dados
package repository

import (
	"context"
	"sync"
	"time"

	"github.com/renatofagalde/golang-fido-webauthn/internal/fido/domain"
	"github.com/renatofagalde/golang-fido-webauthn/pkg/ptr"
)

type inMemoryRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
	seq   int64
}

func NewInMemoryUserRepository() domain.UserRepository {
	return &inMemoryRepository{users: make(map[string]*domain.User)}
}

func (r *inMemoryRepository) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return domain.ErrInvalidInput
	}

	r.mu.Lock()

	defer r.mu.Unlock()

	r.seq++
	user.ID = r.seq
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	out := *user
	r.users[user.Hash] = &out
	return nil
}

func (r *inMemoryRepository) GetByHash(ctx context.Context, hash string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[hash]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return ptr.Of(*u), nil
}

func (r *inMemoryRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}
