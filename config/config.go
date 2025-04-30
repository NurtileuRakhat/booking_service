package config

import (
	"booking/internal/infrastructure/external/gcalendar"
	"booking/internal/infrastructure/external/telegram"
	"booking/pkg/logger"
	"github.com/joho/godotenv"
	"os"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type TelegramConfig struct {
	BotToken string
	ChatID   string
}

type Config struct {
	DB               DBConfig
	GCalendarAdapter *gcalendar.GCalendarAdapter
	Telegram         TelegramConfig
	TelegramAdapter  *telegram.TelegramAdapter
	JWTSecret        string
}

func LoadConfig() *Config {
	_ = godotenv.Load()
	cfg := &Config{
		DB: DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		},
		Telegram: TelegramConfig{
			BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
			ChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		},
		JWTSecret: os.Getenv("JWT_SECRET"),
	}

	calendarCredentials := os.Getenv("GOOGLE_ADMIN_CALENDAR_CREDENTIALS")
	calendarID := os.Getenv("GOOGLE_ADMIN_CALENDAR_ID")
	if calendarCredentials != "" && calendarID != "" {
		gcalAdapter, err := gcalendar.NewGCalendarAdapter(calendarCredentials, calendarID)
		if err != nil {
			logger.Error("failed to init Google Calendar adapter: %v", err)
		}
		cfg.GCalendarAdapter = gcalAdapter
	}

	if cfg.Telegram.BotToken != "" {
		tgAdapter, err := telegram.NewTelegramAdapter(cfg.Telegram.BotToken)
		if err != nil {
			logger.Error("failed to init Telegram adapter: %v", err)
		}
		cfg.TelegramAdapter = tgAdapter
	}
	if cfg.DB.Host == "" {
		logger.Error("DB_HOST is required")
	}

	return cfg
}
