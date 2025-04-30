package user

import (
	"booking/internal/entity"
	"context"
)

type Service interface {
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByTelegramChatID(ctx context.Context, chatID int64) (*entity.User, error)
	RegisterUser(ctx context.Context, email string, password string, telegramChatID int64) (*entity.User, error)
	LinkTelegramToUserByEmail(ctx context.Context, email string, chatID int64) error
}
