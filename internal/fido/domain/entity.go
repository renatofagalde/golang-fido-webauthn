package domain

import "time"

type User struct {
	ID   int64
	Hash string
	Username,
	DisplayName string

	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	CreatedBy *int64
	UpdatedBy *int64
}
