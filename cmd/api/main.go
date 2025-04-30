package main

import (
	"booking/config"
	"booking/internal/infrastructure/repository/postgres"
	"booking/internal/ports/api"
	authhandler "booking/internal/ports/api/auth"
	bookinghandler "booking/internal/ports/api/booking"
	workspaceapi "booking/internal/ports/api/workspace"
	bookingusecase "booking/internal/usecase/booking"
	"booking/internal/usecase/pricing"
	userusecase "booking/internal/usecase/user"
	workspace2 "booking/internal/usecase/workspace"
	"booking/pkg/logger"
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"time"
)

func main() {
	if err := logger.Init("logs/project.log", "logs/telegram_bot.log"); err != nil {
		log.Fatalf("logger init failed: %v", err)
	}

	cfg := config.LoadConfig()
	pgURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.DBName, cfg.DB.SSLMode)

	logger.Info("Connecting to DB: %s", pgURL)
	db, err := sqlx.Connect("postgres", pgURL)
	if err != nil {
		logger.Error("Failed to connect to DB: %v", err)
		return
	}
	logger.Info("Successfully connected to DB")

	// Инициализация зависимостей
	userRepo := postgres.NewUserRepository(db)
	workspaceRepo := postgres.NewWorkspaceRepository(db)
	pricingService := pricing.NewPricingService(workspaceRepo)
	workspaceService := workspace2.NewWorkspaceService(workspaceRepo)
	bookingRepo := postgres.NewBookingRepository(db)

	userService := userusecase.NewUserService(userRepo)

	bookingService := bookingusecase.NewBookingService(
		bookingRepo,
		userRepo,
		workspaceRepo,
		pricingService,
		cfg.GCalendarAdapter,
		cfg.TelegramAdapter)

	bookingHandler := bookinghandler.NewBookingHandler(bookingService)
	workspaceHandler := workspaceapi.NewWorkspaceHandler(workspaceService)
	authHandler := authhandler.NewAuthHandler(userRepo)
	r := api.SetupRouter(bookingHandler, workspaceHandler, authHandler)

	go func() {
		logger.Info("Starting goroutine to mark past bookings as completed")
		for {
			err := bookingService.MarkPastBookingsCompleted(context.Background())
			if err != nil {
				logger.Error("Failed to mark past bookings as completed: %v", err)
			}
			time.Sleep(time.Minute)
		}
	}()

	go func() {
		logger.Info("Starting Telegram adapter listener")
		cfg.TelegramAdapter.ListenForCallbacksAndCommands(userRepo, bookingService, workspaceService, userService)
		logger.Info("Telegram adapter listener stopped")
	}()

	logger.Info("Starting API server on :8080")
	err = r.Run(":8080")
	if err != nil {
		logger.Error("API server stopped with error: %v", err)
	}
	logger.Info("API server stopped")
}
