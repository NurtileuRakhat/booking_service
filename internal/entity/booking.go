package entity

import "time"

type BookingStatus string

const (
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusCompleted BookingStatus = "completed"
)

type Booking struct {
	ID          int64         `db:"id" json:"id"`
	UserID      int64         `db:"user_id" json:"user_id"`
	WorkspaceID int64         `db:"workspace_id" json:"workspace_id"`
	StartTime   time.Time     `db:"start_time" json:"start_time"`
	EndTime     time.Time     `db:"end_time" json:"end_time"`
	Price       float64       `db:"price" json:"price"`
	Status      BookingStatus `db:"status" json:"status"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
	GoogleCalendarEventID string `db:"google_calendar_event_id" json:"google_calendar_event_id"`
}
