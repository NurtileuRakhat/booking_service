// --- booking/internal/usecase/booking/service.go ---
package booking

import (
	"booking/internal/entity"
	"booking/internal/ports/repository"
	"booking/pkg/logger"
	"context"
	"errors"
	"fmt"
	"time"
)

// BookingService
type BookingService struct {
	repo            repository.BookingRepository
	userRepo        repository.UserRepository
	workspaceRepo   repository.WorkspaceRepository
	pricingService  PricingService
	gcalAdapter     GCalendarAdapter
	telegramAdapter TelegramAdapter
}

// PricingService
type PricingService interface {
	CalculatePrice(ctx context.Context, workspaceID int64, start, end time.Time, userID int64) (float64, error)
}

// TelegramAdapter
type TelegramAdapter interface {
	SendBookingConfirmation(ctx context.Context, chatID int64, message string) error
}

// GCalendarAdapter - Google Calendar.
type GCalendarAdapter interface {
	ScheduleEvent(ctx context.Context, summary, desc string, start, end time.Time) (string, error)
	DeleteEvent(ctx context.Context, eventID string) error
}

// NewBookingService
func NewBookingService(
	repo repository.BookingRepository,
	userRepo repository.UserRepository,
	workspaceRepo repository.WorkspaceRepository,
	pricingService PricingService,
	gcalAdapter GCalendarAdapter,
	telegramAdapter TelegramAdapter,
) *BookingService {
	return &BookingService{
		repo:            repo,
		userRepo:        userRepo,
		workspaceRepo:   workspaceRepo,
		pricingService:  pricingService,
		gcalAdapter:     gcalAdapter,
		telegramAdapter: telegramAdapter,
	}
}

// CreateBooking создает новое бронирование.
func (s *BookingService) CreateBooking(ctx context.Context, userID int64, workspaceID int64, start, end time.Time) (*entity.Booking, error) {
	logger.Info("CreateBooking called with userID=%d, workspaceID=%d, start=%s, end=%s", userID, workspaceID, start.Format(time.RFC3339), end.Format(time.RFC3339))
	if userID <= 0 || workspaceID <= 0 {
		logger.Error("Invalid user or workspace id: userID=%d, workspaceID=%d", userID, workspaceID)
		return nil, errors.New("invalid user or workspace id")
	}
	if !start.Before(end) {
		logger.Error("Invalid booking interval: start_time >= end_time (start=%s, end=%s)", start.Format(time.RFC3339), end.Format(time.RFC3339))
		return nil, errors.New("start_time must be before end_time")
	}

	// Проверка на будущее время в текущем 	часовом поясе сервера/приложения
	if start.Before(time.Now()) {
		logger.Error("Invalid booking time: start_time must be in the future (start=%s, now=%s)", start.Format(time.RFC3339), time.Now().Format(time.RFC3339))
		return nil, errors.New("start_time must be in the future")
	}

	// Проверка на рабочее время (8 до 21 включительно)

	if start.Hour() < 8 || start.Hour() >= 21 {
		logger.Error("Invalid booking time: start_time must be between 08:00 and 20:00 (start=%s)", start.Format(time.RFC3339))
		return nil, errors.New("booking start time must be between 08:00 and 20:00")
	}

	// Also check if the duration is exactly 1 hour based on the time selection in Telegram
	if end.Sub(start) != time.Hour {
		logger.Error("Invalid booking duration: %s. Expected 1 hour.", end.Sub(start))
		return nil, errors.New("booking duration must be exactly one hour")
	}

	// Проверка пользователя
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Error("User %d not found: %v", userID, err)
		return nil, errors.New("user not found")
	}

	// Проверка workspace
	workspace, err := s.workspaceRepo.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		logger.Error("Workspace %d not found: %v", workspaceID, err)
		return nil, errors.New("workspace not found")
	}

	// Check availability using the repository method
	flag, err := s.repo.GetConflictingBookings(ctx, workspaceID, start, end)
	if err != nil {
		logger.Error("Error checking availability for ws %d, %s-%s: %v", workspaceID, start.String(), end.String(), err)
		return nil, errors.New("failed to check availability")
	}
	if flag {
		return nil, errors.New("workspace is not available")
	}

	// Price
	price, err := s.pricingService.CalculatePrice(ctx, workspaceID, start, end, userID)
	if err != nil {
		logger.Error("Error calculating price for ws %d, %s-%s, user %d: %v", workspaceID, start.String(), end.String(), userID, err)
		return nil, errors.New("failed to calculate price")
	}

	// Создание бронирования
	booking := &entity.Booking{
		UserID:      userID,
		WorkspaceID: workspaceID,
		StartTime:   start,
		EndTime:     end,
		Price:       price,
		Status:      entity.BookingStatusConfirmed,
	}
	id, err := s.repo.CreateBooking(ctx, booking)
	if err != nil {
		logger.Error("Error creating booking in repository: %v", err)
		return nil, errors.New("failed to create booking")
	}
	booking.ID = id

	// Google Calendar Integration
	if s.gcalAdapter != nil {
		summary := fmt.Sprintf("Booking #%d - Workspace %s", booking.ID, workspace.Name)
		description := fmt.Sprintf("User: %s\nTime: %s - %s (GMT+5)", user.Email, FormatTime(booking.StartTime), FormatTime(booking.EndTime))

		eventID, err := s.gcalAdapter.ScheduleEvent(ctx, summary, description, booking.StartTime, booking.EndTime)
		if err != nil {
			logger.Error("Failed to create event in Google Calendar for booking %d: %v", booking.ID, err)
		} else {
			booking.GoogleCalendarEventID = eventID
			err = s.repo.UpdateBooking(ctx, booking)
			if err != nil {
				logger.Error("Failed to update booking %d with Google Calendar event ID: %v", booking.ID, err)
			} else {
				logger.Info("Booking %d updated with Google Calendar event ID: %s", booking.ID, eventID)
			}
		}
	}

	// Telegram notification
	if s.telegramAdapter != nil {
		if user.TelegramChatID.Valid {
			chatID := user.TelegramChatID.Int64
			locationGMT5 := time.FixedZone("GMT+5", 5*60*60)
			msg := fmt.Sprintf("✅ Booking created successfully!\n\nBooking #%d\nDate: %s\nTime: %s — %s\nStatus: %s",
				booking.ID,
				FormatDate(booking.StartTime.In(locationGMT5)),
				FormatTime(booking.StartTime.In(locationGMT5)),
				FormatTime(booking.EndTime.In(locationGMT5)),
				workspace.Name,
			)
			err := s.telegramAdapter.SendBookingConfirmation(ctx, chatID, msg)
			if err != nil {
				logger.TelegramError("Failed to send Telegram booking confirmation to chat %d: %v", chatID, err)
			} else {
				logger.TelegramInfo("Telegram booking confirmation sent to chat %d for booking %d", chatID, booking.ID)
			}
		} else {
			logger.TelegramError("Telegram chat_id is not set for user %d for confirmation.", userID)
		}
	}

	logger.Info("Booking %d created for user %d in workspace %d from %s to %s",
		booking.ID, userID, workspaceID, booking.StartTime.String(), booking.EndTime.String())

	return booking, nil
}

func (s *BookingService) ListBookingsByUser(ctx context.Context, userID int64) ([]entity.Booking, error) {
	logger.Info("Listing bookings for user %d", userID)
	return s.repo.ListBookingsByUser(ctx, userID)
}

// CancelBooking отменяет бронирование.
func (s *BookingService) CancelBooking(ctx context.Context, userID, bookingID int64) (float64, error) {
	logger.Info("Attempting to cancel booking %d by user %d", bookingID, userID)
	booking, err := s.repo.GetBookingByID(ctx, bookingID)
	if err != nil {
		logger.Error("Booking %d not found for cancellation by user %d: %v", bookingID, userID, err)
		return 0, errors.New("booking not found")
	}
	if booking.UserID != userID {
		logger.Error("User %d attempted to cancel booking %d belonging to user %d", userID, bookingID, booking.UserID)
		return 0, errors.New("forbidden: you can only cancel your own bookings")
	}
	if booking.Status == entity.BookingStatusCancelled {
		logger.Info("Booking %d is already cancelled", bookingID)
		return 0, errors.New("booking is already cancelled")
	}

	// if cancellation less than 24h before start — 50% of price
	penalty := 0.0
	if time.Until(booking.StartTime) < 24*time.Hour {
		penalty = booking.Price * 0.5
		logger.Info("Applying 50%% penalty for booking %d cancellation (less than 24h notice). Penalty: %.2f", bookingID, penalty)
	} else {
		logger.Info("No penalty applied for booking %d cancellation (more than 24h notice).", bookingID)
	}

	// Cancel in DB
	err = s.repo.CancelBooking(ctx, bookingID, penalty)
	if err != nil {
		logger.Error("Failed to cancel booking %d in repository: %v", bookingID, err)
		return 0, errors.New("failed to cancel booking")
	}

	// Remove from Google Calendar if event ID exists
	if s.gcalAdapter != nil && booking.GoogleCalendarEventID != "" {
		err := s.gcalAdapter.DeleteEvent(ctx, booking.GoogleCalendarEventID)
		if err != nil {
			logger.Error("Failed to delete Google Calendar event '%s' for booking %d: %v", booking.GoogleCalendarEventID, bookingID, err)
		} else {
			logger.Info("Google Calendar event '%s' deleted for booking %d", booking.GoogleCalendarEventID, bookingID)
		}
	}

	logger.Info("Booking %d successfully cancelled by user %d", bookingID, userID)
	return penalty, nil
}

// GetConflictingBookings возвращает список бронирований, конфликтующих с заданным интервалом.
func (s *BookingService) GetConflictingBookings(ctx context.Context, workspaceID int64, start, end time.Time) (bool, error) {
	logger.Info("Checking conflicts for workspace %d from %s to %s", workspaceID, start.String(), end.String()) // Too verbose
	return s.repo.GetConflictingBookings(ctx, workspaceID, start, end)
}

// MarkPastBookingsCompleted помечает прошедшие бронирования как завершенные.
func (s *BookingService) MarkPastBookingsCompleted(ctx context.Context) error {
	logger.Info("Starting process to mark past bookings as completed.")
	bookings, err := s.repo.ListAllBookings(ctx)
	if err != nil {
		logger.Error("Error listing all bookings to mark completed: %v", err)
		return fmt.Errorf("failed to list all bookings: %w", err)
	}

	locationGMT5 := time.FixedZone("GMT+5", 5*60*60)
	now := time.Now().In(locationGMT5)
	countCompleted := 0

	for _, b := range bookings {
		bookingEndTimeInZone := b.EndTime.In(locationGMT5)
		if b.Status == entity.BookingStatusConfirmed && bookingEndTimeInZone.Before(now) {
			logger.Info("Marking booking %d as completed (EndTime: %s < Now: %s)", b.ID, bookingEndTimeInZone.Format("2006-01-02 15:04"), now.Format("2006-01-02 15:04"))
			b.Status = entity.BookingStatusCompleted
			err := s.repo.UpdateBooking(ctx, &b)
			if err != nil {
				logger.Error("Failed to update booking %d to completed: %v", b.ID, err)
			} else {
				countCompleted++
			}
		}
	}
	logger.Info("Finished marking past bookings as completed. Marked %d bookings.", countCompleted)
	return nil
}

func FormatDate(t time.Time) string {
	return t.Format("02 Jan 2006")
}

func FormatTime(t time.Time) string {
	return t.Format("15:04")
}
