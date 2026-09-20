package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByHash(ctx context.Context, hash string) (*User, error)
	ExistsByUsername(ctx context.Context) (bool, error)
}
