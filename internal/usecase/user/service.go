package user

import (
	"booking/internal/entity"
	"booking/internal/ports/repository"
	"context"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	return s.repo.GetUserByID(ctx, id)
}
