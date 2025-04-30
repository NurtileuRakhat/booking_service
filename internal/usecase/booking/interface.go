package booking

import (
	"booking/internal/entity"
	"context"
	"time"
)

type Service interface {
	CreateBooking(ctx context.Context, userID int64, workspaceID int64, start, end time.Time) (*entity.Booking, error)
	ListBookingsByUser(ctx context.Context, userID int64) ([]entity.Booking, error)
	CancelBooking(ctx context.Context, userID, bookingID int64) (float64, error)
	GetConflictingBookings(ctx context.Context, workspaceID int64, start, end time.Time) (bool, error)
}
