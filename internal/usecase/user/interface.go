package user

import (
	"booking/internal/entity"
	"context"
)

type Service interface {
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
}
