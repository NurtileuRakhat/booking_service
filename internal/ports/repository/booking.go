package repository

import (
	"booking/internal/entity"
	"context"
	"time"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *entity.Booking) (int64, error)
	GetConflictingBookings(ctx context.Context, workspaceID int64, start, end time.Time) ([]entity.Booking, error)
	ListBookingsByUser(ctx context.Context, userID int64) ([]entity.Booking, error)
	CancelBooking(ctx context.Context, bookingID int64, penalty float64) error
	GetBookingByID(ctx context.Context, bookingID int64) (*entity.Booking, error)
	UpdateBooking(ctx context.Context, booking *entity.Booking) error
	ListAllBookings(ctx context.Context) ([]entity.Booking, error)
}
