package postgres

import (
	"booking/internal/entity"
	"booking/internal/ports/repository"
	"booking/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	var u entity.User
	err := r.db.GetContext(ctx, &u, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		logger.Error("No user found for id: %d, err: %v", id, err)
		return nil, err
	}
	logger.Info("User found for id: %d", id)
	return &u, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) (int64, error) {
	var id int64
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO users (email, username, name, surname, password_hash) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		user.Email, user.Username, user.Name, user.Surname, user.PasswordHash,
	).Scan(&id)
	if err != nil {
		logger.Error("Failed to create user (email: %s): %v", user.Email, err)
		return 0, err
	}
	logger.Info("User created successfully (id: %d, email: %s)", id, user.Email)
	return id, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	// Создаем структуру для хранения результата
	var user entity.User

	query := `SELECT * FROM users WHERE email = $1`
	logger.Info("Executing query: %s with email: '%s'", query, email)

	// Выполняем запрос
	err := r.db.GetContext(ctx, &user, query, email)

	if err != nil {
		// Если пользователь не найден
		if errors.Is(err, sql.ErrNoRows) {
			logger.Info("No user found for email: '%s'", email)
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		// Другие ошибки БД
		logger.Error("Database error for email '%s': %v", email, err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	logger.Info("User found for email: '%s' (id: %d)", email, user.ID)
	return &user, nil
}

func (r *UserRepository) GetUserByTelegramChatID(ctx context.Context, chatID int64) (*entity.User, error) {
	var user entity.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE telegram_chat_id = $1", chatID)
	if err != nil {
		logger.Error("No user found for telegram_chat_id: %d, err: %v", chatID, err)
		return nil, err
	}
	logger.Info("User found for telegram_chat_id: %d (id: %d)", chatID, user.ID)
	return &user, nil
}

func (r *UserRepository) LinkTelegramToUserByEmail(ctx context.Context, email string, chatID int64) error {
	query := `UPDATE users SET telegram_chat_id = $1 WHERE email = $2`
	_, err := r.db.ExecContext(ctx, query, chatID, email)
	return err
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET email = $1, 
			username = $2, 
			name = $3, 
			surname = $4, 
			telegram_chat_id = $5
		WHERE id = $6
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		user.Email,
		user.Username,
		user.Name,
		user.Surname,
		user.TelegramChatID,
		user.ID,
	)

	if err != nil {
		logger.Error("Failed to update user (id: %d): %v", user.ID, err)
		return err
	}

	logger.Info("User updated successfully (id: %d, email: %s)", user.ID, user.Email)
	return nil
}
