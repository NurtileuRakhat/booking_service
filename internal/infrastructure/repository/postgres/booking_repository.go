package postgres

import (
	"booking/internal/entity"
	"booking/internal/ports/repository"
	"booking/pkg/logger"
	"context"
	"github.com/jmoiron/sqlx"
	"time"
)

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) repository.BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) CreateBooking(ctx context.Context, booking *entity.Booking) (int64, error) {
	query := `INSERT INTO bookings (user_id, workspace_id, start_time, end_time, price, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`
	var id int64
	logger.Info("Creating booking for user_id: %d, workspace_id: %d", booking.UserID, booking.WorkspaceID)
	err := r.db.QueryRowContext(ctx, query,
		booking.UserID,
		booking.WorkspaceID,
		booking.StartTime,
		booking.EndTime,
		booking.Price,
		booking.Status,
	).Scan(&id)
	if err != nil {
		logger.Error("Failed to create booking for user_id: %d, err: %v", booking.UserID, err)
		return 0, err
	}
	logger.Info("Booking created successfully (id: %d)", id)
	return id, nil
}

func (r *BookingRepository) GetConflictingBookings(ctx context.Context, workspaceID int64, start, end time.Time) ([]entity.Booking, error) {
	var bookings []entity.Booking
	logger.Info("Checking conflicts for workspace_id: %d, time: %s - %s", workspaceID, start.Format(time.RFC3339), end.Format(time.RFC3339))
	query := `SELECT * FROM bookings WHERE workspace_id = $1 AND status != 'cancelled' AND (
		(start_time, end_time) OVERLAPS ($2, $3)
	)`
	err := r.db.SelectContext(ctx, &bookings, query, workspaceID, start, end)
	if err != nil {
		logger.Error("Failed to check conflicts for workspace_id: %d, err: %v", workspaceID, err)
	}
	return bookings, err
}

func (r *BookingRepository) ListBookingsByUser(ctx context.Context, userID int64) ([]entity.Booking, error) {
	var bookings []entity.Booking
	logger.Info("Selecting bookings for user_id: %d", userID)
	err := r.db.SelectContext(ctx, &bookings, "SELECT * FROM bookings WHERE user_id = $1 ORDER BY start_time DESC", userID)
	if err != nil {
		logger.Error("Failed to select bookings for user_id: %d, err: %v", userID, err)
	}
	return bookings, err
}

func (r *BookingRepository) GetBookingByID(ctx context.Context, bookingID int64) (*entity.Booking, error) {
	var booking entity.Booking
	logger.Info("Getting booking by id: %d", bookingID)
	err := r.db.GetContext(ctx, &booking, "SELECT * FROM bookings WHERE id = $1", bookingID)
	if err != nil {
		logger.Error("Failed to get booking by id: %d, err: %v", bookingID, err)
		return nil, err
	}
	logger.Info("Booking found: id %d", bookingID)
	return &booking, nil
}

func (r *BookingRepository) CancelBooking(ctx context.Context, bookingID int64, penalty float64) error {
	logger.Info("Cancelling booking id: %d with penalty: %.2f", bookingID, penalty)
	_, err := r.db.ExecContext(ctx, `UPDATE bookings SET status = 'cancelled', price = price + $1, updated_at = NOW() WHERE id = $2`, penalty, bookingID)
	if err != nil {
		logger.Error("Failed to cancel booking id: %d, err: %v", bookingID, err)
		return err
	}
	logger.Info("Booking cancelled successfully (id: %d)", bookingID)
	return nil
}

func (r *BookingRepository) UpdateBooking(ctx context.Context, booking *entity.Booking) error {
	logger.Info("Updating booking id: %d", booking.ID)
	_, err := r.db.ExecContext(ctx, `
		UPDATE bookings SET
			user_id = $1,
			workspace_id = $2,
			start_time = $3,
			end_time = $4,
			price = $5,
			status = $6,
			updated_at = NOW(),
			google_calendar_event_id = $7
		WHERE id = $8`,
		booking.UserID,
		booking.WorkspaceID,
		booking.StartTime,
		booking.EndTime,
		booking.Price,
		booking.Status,
		booking.GoogleCalendarEventID,
		booking.ID,
	)
	if err != nil {
		logger.Error("Failed to update booking id: %d, err: %v", booking.ID, err)
		return err
	}
	logger.Info("Booking updated successfully (id: %d)", booking.ID)
	return nil
}

func (r *BookingRepository) ListAllBookings(ctx context.Context) ([]entity.Booking, error) {
	var bookings []entity.Booking
	logger.Info("Selecting all bookings")
	err := r.db.SelectContext(ctx, &bookings, "SELECT * FROM bookings")
	if err != nil {
		logger.Error("Failed to select all bookings, err: %v", err)
	}
	return bookings, err
}
