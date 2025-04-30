package booking_test

import (
	"booking/internal/entity"
	"context"
	"time"
)

// Mock репозитория бронирований
type MockBookingRepository struct {
	CreateBookingFn             func(ctx context.Context, booking *entity.Booking) (int64, error)
	GetBookingFn                func(ctx context.Context, id int64) (*entity.Booking, error)
	UpdateBookingFn             func(ctx context.Context, booking *entity.Booking) error
	ListBookingsByUserFn        func(ctx context.Context, userID int64) ([]entity.Booking, error)
	GetConflictingBookingsFn    func(ctx context.Context, workspaceID int64, start, end time.Time) ([]entity.Booking, error)
	MarkPastBookingsCompletedFn func(ctx context.Context) (int, error)
	CancelBookingFn             func(ctx context.Context, bookingID int64, penalty float64) error
	GetBookingByIDFn            func(ctx context.Context, bookingID int64) (*entity.Booking, error)
	ListAllBookingsFn           func(ctx context.Context) ([]entity.Booking, error)
}

func (m *MockBookingRepository) CreateBooking(ctx context.Context, booking *entity.Booking) (int64, error) {
	return m.CreateBookingFn(ctx, booking)
}

func (m *MockBookingRepository) GetBooking(ctx context.Context, id int64) (*entity.Booking, error) {
	return m.GetBookingFn(ctx, id)
}

func (m *MockBookingRepository) UpdateBooking(ctx context.Context, booking *entity.Booking) error {
	return m.UpdateBookingFn(ctx, booking)
}

func (m *MockBookingRepository) ListBookingsByUser(ctx context.Context, userID int64) ([]entity.Booking, error) {
	return m.ListBookingsByUserFn(ctx, userID)
}

func (m *MockBookingRepository) GetConflictingBookings(ctx context.Context, workspaceID int64, start, end time.Time) ([]entity.Booking, error) {
	return m.GetConflictingBookingsFn(ctx, workspaceID, start, end)
}

func (m *MockBookingRepository) MarkPastBookingsCompleted(ctx context.Context) (int, error) {
	return m.MarkPastBookingsCompletedFn(ctx)
}

func (m *MockBookingRepository) CancelBooking(ctx context.Context, bookingID int64, penalty float64) error {
	return m.CancelBookingFn(ctx, bookingID, penalty)
}

func (m *MockBookingRepository) GetBookingByID(ctx context.Context, bookingID int64) (*entity.Booking, error) {
	return m.GetBookingByIDFn(ctx, bookingID)
}

func (m *MockBookingRepository) ListAllBookings(ctx context.Context) ([]entity.Booking, error) {
	return m.ListAllBookingsFn(ctx)
}

// Mock репозитория пользователей
type MockUserRepository struct {
	GetUserByIDFn             func(ctx context.Context, id int64) (*entity.User, error)
	CreateUserFn              func(ctx context.Context, user *entity.User) (int64, error)
	GetUserByEmailFn          func(ctx context.Context, email string) (*entity.User, error)
	GetUserByTelegramChatIDFn func(ctx context.Context, chatID int64) (*entity.User, error)
	UpdateTelegramChatIDFn    func(ctx context.Context, userID, chatID int64) error
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	return m.GetUserByIDFn(ctx, id)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *entity.User) (int64, error) {
	return m.CreateUserFn(ctx, user)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return m.GetUserByEmailFn(ctx, email)
}

func (m *MockUserRepository) GetUserByTelegramChatID(ctx context.Context, chatID int64) (*entity.User, error) {
	return m.GetUserByTelegramChatIDFn(ctx, chatID)
}

func (m *MockUserRepository) UpdateTelegramChatID(ctx context.Context, userID, chatID int64) error {
	return m.UpdateTelegramChatIDFn(ctx, userID, chatID)
}

// Mock репозитория рабочих пространств
type MockWorkspaceRepository struct {
	GetWorkspaceByIDFn func(ctx context.Context, id int64) (*entity.Workspace, error)
	ListWorkspacesFn   func(ctx context.Context) ([]entity.Workspace, error)
	CreateWorkspaceFn  func(ctx context.Context, workspace *entity.Workspace) (int64, error)
}

func (m *MockWorkspaceRepository) GetWorkspaceByID(ctx context.Context, id int64) (*entity.Workspace, error) {
	return m.GetWorkspaceByIDFn(ctx, id)
}

func (m *MockWorkspaceRepository) ListWorkspaces(ctx context.Context) ([]entity.Workspace, error) {
	return m.ListWorkspacesFn(ctx)
}

func (m *MockWorkspaceRepository) CreateWorkspace(ctx context.Context, workspace *entity.Workspace) (int64, error) {
	return m.CreateWorkspaceFn(ctx, workspace)
}

// Mock сервиса цен
type MockPricingService struct {
	CalculatePriceFn func(ctx context.Context, workspaceID int64, start, end time.Time, userID int64) (float64, error)
}

func (m *MockPricingService) CalculatePrice(ctx context.Context, workspaceID int64, start, end time.Time, userID int64) (float64, error) {
	return m.CalculatePriceFn(ctx, workspaceID, start, end, userID)
}

// Mock Google Calendar
type MockGCalendarAdapter struct {
	ScheduleEventFn func(ctx context.Context, summary, desc string, start, end time.Time) (string, error)
	DeleteEventFn   func(ctx context.Context, eventID string) error
}

func (m *MockGCalendarAdapter) ScheduleEvent(ctx context.Context, summary, desc string, start, end time.Time) (string, error) {
	return m.ScheduleEventFn(ctx, summary, desc, start, end)
}

func (m *MockGCalendarAdapter) DeleteEvent(ctx context.Context, eventID string) error {
	return m.DeleteEventFn(ctx, eventID)
}

// Mock Telegram
type MockTelegramAdapter struct {
	SendBookingConfirmationFn func(ctx context.Context, chatID int64, message string) error
}

func (m *MockTelegramAdapter) SendBookingConfirmation(ctx context.Context, chatID int64, message string) error {
	return m.SendBookingConfirmationFn(ctx, chatID, message)
}
