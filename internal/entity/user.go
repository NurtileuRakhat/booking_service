package entity

import "database/sql"

type User struct {
	ID           int64         `db:"id" json:"id"`
	Email        string        `db:"email" json:"email"`
	Username     string        `db:"username" json:"username"`
	Name         string        `db:"name" json:"name"`
	Surname      string        `db:"surname" json:"surname"`
	PasswordHash string        `db:"password_hash" json:"-"`
	TelegramChatID sql.NullInt64 `db:"telegram_chat_id"`
}
