package booking_test

import (
	"booking/internal/entity"
	"booking/internal/usecase/booking"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateBooking(t *testing.T) {
	// Общий шаблон для времени бронирования
	now := time.Now()
	locationGMT5 := time.FixedZone("GMT+5", 5*60*60)
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 10, 0, 0, 0, locationGMT5) // 10:00 завтра
	tomorrowEnd := tomorrow.Add(time.Hour)                                                 // 11:00 завтра

	testUser := &entity.User{
		ID:    1,
		Email: "test@example.com",
		TelegramChatID: sql.NullInt64{
			Int64: 123456789,
			Valid: true,
		},
	}

	testWorkspace := &entity.Workspace{
		ID:   1,
		Name: "Test Workspace",
	}

	t.Run("Successful Booking Creation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo := &MockBookingRepository{
			CreateBookingFn: func(ctx context.Context, booking *entity.Booking) (int64, error) {
				return 1, nil // Успешное создание с ID = 1
			},
			GetConflictingBookingsFn: func(ctx context.Context, workspaceID int64, start, end time.Time) ([]entity.Booking, error) {
				return []entity.Booking{}, nil // Нет конфликтов
			},
			UpdateBookingFn: func(ctx context.Context, booking *entity.Booking) error {
				return nil // Успешное обновление
			},
			// Добавляем заглушки для остальных методов
			GetBookingFn: func(ctx context.Context, id int64) (*entity.Booking, error) {
				return nil, nil
			},
			ListBookingsByUserFn: func(ctx context.Context, userID int64) ([]entity.Booking, error) {
				return nil, nil
			},
			MarkPastBookingsCompletedFn: func(ctx context.Context) (int, error) {
				return 0, nil
			},
			CancelBookingFn: func(ctx context.Context, bookingID int64, penalty float64) error {
				return nil
			},
			GetBookingByIDFn: func(ctx context.Context, bookingID int64) (*entity.Booking, error) {
				return nil, nil
			},
			ListAllBookingsFn: func(ctx context.Context) ([]entity.Booking, error) {
				return nil, nil
			},
		}

		mockUserRepo := &MockUserRepository{
			GetUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				return testUser, nil
			},
			// Добавляем заглушки для остальных методов
			CreateUserFn: func(ctx context.Context, user *entity.User) (int64, error) {
				return 0, nil
			},
			GetUserByEmailFn: func(ctx context.Context, email string) (*entity.User, error) {
				return nil, nil
			},
			GetUserByTelegramChatIDFn: func(ctx context.Context, chatID int64) (*entity.User, error) {
				return nil, nil
			},
			UpdateTelegramChatIDFn: func(ctx context.Context, userID, chatID int64) error {
				return nil
			},
		}

		mockWorkspaceRepo := &MockWorkspaceRepository{
			GetWorkspaceByIDFn: func(ctx context.Context, id int64) (*entity.Workspace, error) {
				return testWorkspace, nil
			},
			// Добавляем заглушки для остальных методов
			ListWorkspacesFn: func(ctx context.Context) ([]entity.Workspace, error) {
				return nil, nil
			},
			CreateWorkspaceFn: func(ctx context.Context, workspace *entity.Workspace) (int64, error) {
				return 0, nil
			},
		}

		mockPricingService := &MockPricingService{
			CalculatePriceFn: func(ctx context.Context, workspaceID int64, start, end time.Time, userID int64) (float64, error) {
				return 50.0, nil // 50 единиц валюты
			},
		}

		mockGCalAdapter := &MockGCalendarAdapter{
			ScheduleEventFn: func(ctx context.Context, summary, desc string, start, end time.Time) (string, error) {
				return "event-123", nil
			},
			DeleteEventFn: func(ctx context.Context, eventID string) error {
				return nil
			},
		}

		mockTelegramAdapter := &MockTelegramAdapter{
			SendBookingConfirmationFn: func(ctx context.Context, chatID int64, message string) error {
				return nil
			},
		}

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		result, err := service.CreateBooking(ctx, testUser.ID, testWorkspace.ID, tomorrow, tomorrowEnd)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, testUser.ID, result.UserID)
		assert.Equal(t, testWorkspace.ID, result.WorkspaceID)
		assert.Equal(t, tomorrow, result.StartTime)
		assert.Equal(t, tomorrowEnd, result.EndTime)
		assert.Equal(t, 50.0, result.Price)
		assert.Equal(t, entity.BookingStatusConfirmed, result.Status)
		assert.Equal(t, "event-123", result.GoogleCalendarEventID)
	})

	// Функция для создания заполненных моков с заглушками
	createCompleteMocks := func() (*MockBookingRepository, *MockUserRepository, *MockWorkspaceRepository, *MockPricingService, *MockGCalendarAdapter, *MockTelegramAdapter) {
		mockBookingRepo := &MockBookingRepository{
			CreateBookingFn: func(ctx context.Context, booking *entity.Booking) (int64, error) {
				return 0, nil
			},
			GetConflictingBookingsFn: func(ctx context.Context, workspaceID int64, start, end time.Time) ([]entity.Booking, error) {
				return nil, nil
			},
			UpdateBookingFn: func(ctx context.Context, booking *entity.Booking) error {
				return nil
			},
			GetBookingFn: func(ctx context.Context, id int64) (*entity.Booking, error) {
				return nil, nil
			},
			ListBookingsByUserFn: func(ctx context.Context, userID int64) ([]entity.Booking, error) {
				return nil, nil
			},
			MarkPastBookingsCompletedFn: func(ctx context.Context) (int, error) {
				return 0, nil
			},
			CancelBookingFn: func(ctx context.Context, bookingID int64, penalty float64) error {
				return nil
			},
			GetBookingByIDFn: func(ctx context.Context, bookingID int64) (*entity.Booking, error) {
				return nil, nil
			},
			ListAllBookingsFn: func(ctx context.Context) ([]entity.Booking, error) {
				return nil, nil
			},
		}

		mockUserRepo := &MockUserRepository{
			GetUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				return nil, nil
			},
			CreateUserFn: func(ctx context.Context, user *entity.User) (int64, error) {
				return 0, nil
			},
			GetUserByEmailFn: func(ctx context.Context, email string) (*entity.User, error) {
				return nil, nil
			},
			GetUserByTelegramChatIDFn: func(ctx context.Context, chatID int64) (*entity.User, error) {
				return nil, nil
			},
			UpdateTelegramChatIDFn: func(ctx context.Context, userID, chatID int64) error {
				return nil
			},
		}

		mockWorkspaceRepo := &MockWorkspaceRepository{
			GetWorkspaceByIDFn: func(ctx context.Context, id int64) (*entity.Workspace, error) {
				return nil, nil
			},
			ListWorkspacesFn: func(ctx context.Context) ([]entity.Workspace, error) {
				return nil, nil
			},
			CreateWorkspaceFn: func(ctx context.Context, workspace *entity.Workspace) (int64, error) {
				return 0, nil
			},
		}

		mockPricingService := &MockPricingService{
			CalculatePriceFn: func(ctx context.Context, workspaceID int64, start, end time.Time, userID int64) (float64, error) {
				return 0, nil
			},
		}

		mockGCalAdapter := &MockGCalendarAdapter{
			ScheduleEventFn: func(ctx context.Context, summary, desc string, start, end time.Time) (string, error) {
				return "", nil
			},
			DeleteEventFn: func(ctx context.Context, eventID string) error {
				return nil
			},
		}

		mockTelegramAdapter := &MockTelegramAdapter{
			SendBookingConfirmationFn: func(ctx context.Context, chatID int64, message string) error {
				return nil
			},
		}

		return mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter
	}

	t.Run("Invalid User ID", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		result, err := service.CreateBooking(ctx, 0, testWorkspace.ID, tomorrow, tomorrowEnd)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid user or workspace id")
	})

	t.Run("Invalid Workspace ID", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		result, err := service.CreateBooking(ctx, testUser.ID, 0, tomorrow, tomorrowEnd)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid user or workspace id")
	})

	t.Run("Start Time Not Before End Time", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		result, err := service.CreateBooking(ctx, testUser.ID, testWorkspace.ID, tomorrow, tomorrow) // Одинаковое время

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "start_time must be before end_time")
	})

	t.Run("Start Time In The Past", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		yesterday := time.Now().Add(-24 * time.Hour)
		yesterdayEnd := yesterday.Add(time.Hour)

		// Act
		result, err := service.CreateBooking(ctx, testUser.ID, testWorkspace.ID, yesterday, yesterdayEnd)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "start_time must be in the future")
	})

	t.Run("Start Time Outside Working Hours", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// 7:00 утра - слишком рано
		earlyTime := time.Date(now.Year(), now.Month(), now.Day()+1, 7, 0, 0, 0, locationGMT5)
		earlyTimeEnd := earlyTime.Add(time.Hour)

		// Act
		result, err := service.CreateBooking(ctx, testUser.ID, testWorkspace.ID, earlyTime, earlyTimeEnd)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "booking start time must be between 08:00 and 20:00")
	})

	t.Run("Invalid Booking Duration", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		endTime := tomorrow.Add(2 * time.Hour) // 2 часа вместо 1

		// Act
		result, err := service.CreateBooking(ctx, testUser.ID, testWorkspace.ID, tomorrow, endTime)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "booking duration must be exactly one hour")
	})

	t.Run("User Not Found", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		mockUserRepo.GetUserByIDFn = func(ctx context.Context, id int64) (*entity.User, error) {
			return nil, errors.New("user not found")
		}

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		result, err := service.CreateBooking(ctx, testUser.ID, testWorkspace.ID, tomorrow, tomorrowEnd)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("Workspace Not Found", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		mockUserRepo.GetUserByIDFn = func(ctx context.Context, id int64) (*entity.User, error) {
			return testUser, nil
		}

		mockWorkspaceRepo.GetWorkspaceByIDFn = func(ctx context.Context, id int64) (*entity.Workspace, error) {
			return nil, errors.New("workspace not found")
		}

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		result, err := service.CreateBooking(ctx, testUser.ID, testWorkspace.ID, tomorrow, tomorrowEnd)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "workspace not found")
	})

	t.Run("Booking Conflict", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		mockUserRepo.GetUserByIDFn = func(ctx context.Context, id int64) (*entity.User, error) {
			return testUser, nil
		}

		mockWorkspaceRepo.GetWorkspaceByIDFn = func(ctx context.Context, id int64) (*entity.Workspace, error) {
			return testWorkspace, nil
		}

		mockBookingRepo.GetConflictingBookingsFn = func(ctx context.Context, workspaceID int64, start, end time.Time) ([]entity.Booking, error) {
			return []entity.Booking{{ID: 999}}, nil // Есть конфликт
		}

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		result, err := service.CreateBooking(ctx, testUser.ID, testWorkspace.ID, tomorrow, tomorrowEnd)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "workspace is not available")
	})

	t.Run("Pricing Error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		mockUserRepo.GetUserByIDFn = func(ctx context.Context, id int64) (*entity.User, error) {
			return testUser, nil
		}

		mockWorkspaceRepo.GetWorkspaceByIDFn = func(ctx context.Context, id int64) (*entity.Workspace, error) {
			return testWorkspace, nil
		}

		mockBookingRepo.GetConflictingBookingsFn = func(ctx context.Context, workspaceID int64, start, end time.Time) ([]entity.Booking, error) {
			return []entity.Booking{}, nil // Нет конфликтов
		}

		mockPricingService.CalculatePriceFn = func(ctx context.Context, workspaceID int64, start, end time.Time, userID int64) (float64, error) {
			return 0, errors.New("pricing calculation failed")
		}

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		result, err := service.CreateBooking(ctx, testUser.ID, testWorkspace.ID, tomorrow, tomorrowEnd)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to calculate price")
	})
}

func TestCancelBooking(t *testing.T) {
	// Подготовка общих данных для тестов
	ctx := context.Background()
	now := time.Now()

	// Существующее бронирование
	existingBooking := &entity.Booking{
		ID:                    1,
		UserID:                1,
		WorkspaceID:           2,
		StartTime:             now.Add(48 * time.Hour), // через 2 дня
		EndTime:               now.Add(49 * time.Hour),
		Price:                 100.0,
		Status:                entity.BookingStatusConfirmed,
		GoogleCalendarEventID: "event-123",
	}

	// Бронирование скоро начнется (менее 24 часов)
	soonBooking := &entity.Booking{
		ID:                    2,
		UserID:                1,
		WorkspaceID:           2,
		StartTime:             now.Add(12 * time.Hour), // через 12 часов
		EndTime:               now.Add(13 * time.Hour),
		Price:                 100.0,
		Status:                entity.BookingStatusConfirmed,
		GoogleCalendarEventID: "event-456",
	}

	// Чужое бронирование
	otherUserBooking := &entity.Booking{
		ID:          3,
		UserID:      2, // другой пользователь
		WorkspaceID: 2,
		StartTime:   now.Add(48 * time.Hour),
		EndTime:     now.Add(49 * time.Hour),
		Price:       100.0,
		Status:      entity.BookingStatusConfirmed,
	}

	// Уже отмененное бронирование
	cancelledBooking := &entity.Booking{
		ID:          4,
		UserID:      1,
		WorkspaceID: 2,
		StartTime:   now.Add(48 * time.Hour),
		EndTime:     now.Add(49 * time.Hour),
		Price:       100.0,
		Status:      entity.BookingStatusCancelled,
	}

	// Функция для создания заполненных моков с заглушками (копия из TestCreateBooking)
	createCompleteMocks := func() (*MockBookingRepository, *MockUserRepository, *MockWorkspaceRepository, *MockPricingService, *MockGCalendarAdapter, *MockTelegramAdapter) {
		mockBookingRepo := &MockBookingRepository{
			CreateBookingFn: func(ctx context.Context, booking *entity.Booking) (int64, error) {
				return 0, nil
			},
			GetConflictingBookingsFn: func(ctx context.Context, workspaceID int64, start, end time.Time) ([]entity.Booking, error) {
				return nil, nil
			},
			UpdateBookingFn: func(ctx context.Context, booking *entity.Booking) error {
				return nil
			},
			GetBookingFn: func(ctx context.Context, id int64) (*entity.Booking, error) {
				return nil, nil
			},
			ListBookingsByUserFn: func(ctx context.Context, userID int64) ([]entity.Booking, error) {
				return nil, nil
			},
			MarkPastBookingsCompletedFn: func(ctx context.Context) (int, error) {
				return 0, nil
			},
			CancelBookingFn: func(ctx context.Context, bookingID int64, penalty float64) error {
				return nil
			},
			GetBookingByIDFn: func(ctx context.Context, bookingID int64) (*entity.Booking, error) {
				return nil, nil
			},
			ListAllBookingsFn: func(ctx context.Context) ([]entity.Booking, error) {
				return nil, nil
			},
		}

		mockUserRepo := &MockUserRepository{
			GetUserByIDFn: func(ctx context.Context, id int64) (*entity.User, error) {
				return nil, nil
			},
			CreateUserFn: func(ctx context.Context, user *entity.User) (int64, error) {
				return 0, nil
			},
			GetUserByEmailFn: func(ctx context.Context, email string) (*entity.User, error) {
				return nil, nil
			},
			GetUserByTelegramChatIDFn: func(ctx context.Context, chatID int64) (*entity.User, error) {
				return nil, nil
			},
			UpdateTelegramChatIDFn: func(ctx context.Context, userID, chatID int64) error {
				return nil
			},
		}

		mockWorkspaceRepo := &MockWorkspaceRepository{
			GetWorkspaceByIDFn: func(ctx context.Context, id int64) (*entity.Workspace, error) {
				return nil, nil
			},
			ListWorkspacesFn: func(ctx context.Context) ([]entity.Workspace, error) {
				return nil, nil
			},
			CreateWorkspaceFn: func(ctx context.Context, workspace *entity.Workspace) (int64, error) {
				return 0, nil
			},
		}

		mockPricingService := &MockPricingService{
			CalculatePriceFn: func(ctx context.Context, workspaceID int64, start, end time.Time, userID int64) (float64, error) {
				return 0, nil
			},
		}

		mockGCalAdapter := &MockGCalendarAdapter{
			ScheduleEventFn: func(ctx context.Context, summary, desc string, start, end time.Time) (string, error) {
				return "", nil
			},
			DeleteEventFn: func(ctx context.Context, eventID string) error {
				return nil
			},
		}

		mockTelegramAdapter := &MockTelegramAdapter{
			SendBookingConfirmationFn: func(ctx context.Context, chatID int64, message string) error {
				return nil
			},
		}

		return mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter
	}

	t.Run("Successful Cancellation No Penalty", func(t *testing.T) {
		// Arrange
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		mockBookingRepo.GetBookingByIDFn = func(ctx context.Context, bookingID int64) (*entity.Booking, error) {
			return existingBooking, nil
		}

		mockBookingRepo.CancelBookingFn = func(ctx context.Context, bookingID int64, penalty float64) error {
			return nil
		}

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		penalty, err := service.CancelBooking(ctx, 1, existingBooking.ID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0.0, penalty, "Should be no penalty for cancellation more than 24h in advance")
	})

	t.Run("Successful Cancellation With Penalty", func(t *testing.T) {
		// Arrange
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		mockBookingRepo.GetBookingByIDFn = func(ctx context.Context, bookingID int64) (*entity.Booking, error) {
			return soonBooking, nil
		}

		mockBookingRepo.CancelBookingFn = func(ctx context.Context, bookingID int64, penalty float64) error {
			return nil
		}

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		penalty, err := service.CancelBooking(ctx, 1, soonBooking.ID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 50.0, penalty, "Should be 50% penalty for cancellation less than 24h in advance")
	})

	t.Run("Booking Not Found", func(t *testing.T) {
		// Arrange
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		mockBookingRepo.GetBookingByIDFn = func(ctx context.Context, bookingID int64) (*entity.Booking, error) {
			return nil, errors.New("booking not found")
		}

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		// Act
		_, err := service.CancelBooking(ctx, 1, 999)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "booking not found")
	})

	t.Run("Not User's Booking", func(t *testing.T) {
		// Arrange
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		mockBookingRepo.GetBookingByIDFn = func(ctx context.Context, bookingID int64) (*entity.Booking, error) {
			return otherUserBooking, nil
		}

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		_, err := service.CancelBooking(ctx, 1, otherUserBooking.ID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "you can only cancel your own bookings")
	})

	t.Run("Already Cancelled", func(t *testing.T) {
		mockBookingRepo, mockUserRepo, mockWorkspaceRepo, mockPricingService, mockGCalAdapter, mockTelegramAdapter := createCompleteMocks()

		mockBookingRepo.GetBookingByIDFn = func(ctx context.Context, bookingID int64) (*entity.Booking, error) {
			return cancelledBooking, nil
		}

		service := booking.NewBookingService(
			mockBookingRepo,
			mockUserRepo,
			mockWorkspaceRepo,
			mockPricingService,
			mockGCalAdapter,
			mockTelegramAdapter,
		)

		_, err := service.CancelBooking(ctx, 1, cancelledBooking.ID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "booking is already cancelled")
	})
}
