package user

import (
	"booking/internal/entity"
	"booking/internal/ports/repository"
	"booking/pkg/logger"
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type service struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) Service {
	return &service{userRepo: userRepo}
}

func (s *service) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	return s.userRepo.GetUserByID(ctx, id)
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return s.userRepo.GetUserByEmail(ctx, email)
}

func (s *service) GetUserByTelegramChatID(ctx context.Context, chatID int64) (*entity.User, error) {
	return s.userRepo.GetUserByTelegramChatID(ctx, chatID)
}

func (s *service) RegisterUser(ctx context.Context, email string, password string, telegramChatID int64) (*entity.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Generate random password if not provided
	var passwordHash string
	if password == "" {
		// In a real app, you would generate a strong random password
		// Here we use a simple default password for telegram registrations
		tempPassword := "TempPassword123!" // This should be randomly generated in production
		hash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Password hash error: %v", err)
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = string(hash)
	} else {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Password hash error: %v", err)
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = string(hash)
	}

	user := &entity.User{
		Email:        email,
		Username:     email[:strings.Index(email, "@")], // Simple username derivation
		PasswordHash: passwordHash,
	}

	// Set telegram chat ID if provided
	if telegramChatID != 0 {
		user.TelegramChatID.Valid = true
		user.TelegramChatID.Int64 = telegramChatID
	}

	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		logger.Error("User creation error: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = userID
	return user, nil
}

func (s *service) LinkTelegramToUserByEmail(ctx context.Context, email string, chatID int64) error {
	return s.userRepo.LinkTelegramToUserByEmail(ctx, email, chatID)
}
