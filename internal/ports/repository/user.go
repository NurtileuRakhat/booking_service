package repository

import (
	"booking/internal/entity"
	"context"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
	CreateUser(ctx context.Context, user *entity.User) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByTelegramChatID(ctx context.Context, chatID int64) (*entity.User, error)
	LinkTelegramToUserByEmail(ctx context.Context, email string, chatID int64) error
	UpdateUser(ctx context.Context, user *entity.User) error
}
