package telegram

import (
	"booking/internal/entity"
	"booking/internal/ports/repository"
	"booking/internal/usecase/booking"
	"booking/internal/usecase/user"
	"booking/internal/usecase/workspace"
	"booking/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
	"time"
)

type TelegramAdapter struct {
	bot *tgbotapi.BotAPI
}

func NewTelegramAdapter(token string) (*TelegramAdapter, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}
	logger.TelegramInfo("Authorized on account %s", bot.Self.UserName)
	return &TelegramAdapter{bot: bot}, nil
}

func (t *TelegramAdapter) SendBookingConfirmation(ctx context.Context, chatID int64, text string) error {
	if chatID == 0 {
		logger.TelegramError("Attempted to send booking confirmation with empty chat_id")
		return fmt.Errorf("Telegram chat_id is empty")
	}
	logger.TelegramInfo("Sending booking confirmation to chat %d: %s", chatID, text)
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.bot.Send(msg)
	if err != nil {
		logger.TelegramError("Failed to send booking confirmation to chat %d: %v", chatID, err)
		return fmt.Errorf("failed to send booking confirmation: %w", err)
	}
	logger.TelegramInfo("Booking confirmation sent to chat %d", chatID)
	return nil
}

func (t *TelegramAdapter) SendReminder(ctx context.Context, chatID int64, text string) error {
	if chatID == 0 {
		logger.TelegramError("Attempted to send reminder with empty chat_id")
		return fmt.Errorf("Telegram chat_id is empty")
	}
	logger.TelegramInfo("Sending reminder to chat %d: %s", chatID, text)
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.bot.Send(msg)
	if err != nil {
		logger.TelegramError("Failed to send reminder to chat %d: %v", chatID, err)
		return fmt.Errorf("failed to send reminder: %w", err)
	}
	logger.TelegramInfo("Reminder sent to chat %d", chatID)
	return nil
}

func (t *TelegramAdapter) sendMainMenuInline(chatID int64) {
	var rows [][]tgbotapi.InlineKeyboardButton
	myBookingsBtn := tgbotapi.NewInlineKeyboardButtonData("My Bookings", "menu:mybookings")
	createBookingBtn := tgbotapi.NewInlineKeyboardButtonData("Create Booking", "menu:createbooking")
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(myBookingsBtn))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(createBookingBtn))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, "Choose an action:")
	msg.ReplyMarkup = keyboard
	_, err := t.bot.Send(msg)
	if err != nil {
		logger.TelegramInfo("Failed to send inline main menu to chat %d: %v", chatID, err)
	}
}

func (t *TelegramAdapter) ListenForCallbacksAndCommands(
	userRepo repository.UserRepository,
	bookingService booking.Service,
	workspaceService workspace.Service,
	userService user.Service,
) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID

			if update.Message.IsCommand() && update.Message.Command() == "start" {
				args := strings.TrimSpace(update.Message.CommandArguments())
				if args != "" && strings.Contains(args, "@") {
					ctx := context.Background()
					existingUser, err := userRepo.GetUserByEmail(ctx, args)
					if err != nil && !strings.Contains(err.Error(), "not found") {
						logger.TelegramError("Error checking user with email %s: %v", args, err)
						t.SendMessage(chatID, "Ошибка при регистрации. Попробуйте позже.")
						continue
					}

					if existingUser != nil {
						existingUser.TelegramChatID = sql.NullInt64{
							Int64: chatID,
							Valid: true,
						}
						err = userRepo.UpdateUser(ctx, existingUser)
						if err != nil {
							logger.TelegramError("Error updating user %d with chatID %d: %v", existingUser.ID, chatID, err)
							t.SendMessage(chatID, "Ошибка при обновлении данных. Попробуйте позже.")
							continue
						}
						t.SendMessage(chatID, fmt.Sprintf("Привет! Вы успешно привязали аккаунт с email %s к Telegram", args))
					} else {
						newUser, err := userService.RegisterUser(ctx, args, "", chatID)
						if err != nil {
							logger.TelegramError("Error registering new user with email %s, chatID %d: %v", args, chatID, err)
							t.SendMessage(chatID, "Ошибка при регистрации. Попробуйте позже.")
							continue
						}
						logger.TelegramInfo("New user registered: email: %s, chatID: %d, userID: %d", args, chatID, newUser.ID)
						t.SendMessage(chatID, fmt.Sprintf("Привет! Регистрация с email %s прошла успешно", args))
					}
					t.sendMainMenuInline(chatID)
				} else {
					t.SendMessage(chatID, "Для регистрации используйте команду /start your@email.com")
				}
				continue
			}

			user, err := userRepo.GetUserByTelegramChatID(context.Background(), chatID)
			if err != nil || user == nil {
				logger.TelegramError("User with chatID %d not found. Error: %v", chatID, err)
				t.SendMessage(chatID, "Вы не зарегистрированы. Введите /start your@email.com для регистрации.")
				continue
			}

			if update.Message.IsCommand() {
				t.handleCommands(update, chatID, user, bookingService, workspaceService)
				continue
			}

			logger.TelegramInfo("Received unexpected text message from chat %d: %s", chatID, update.Message.Text)
			t.SendMessage(chatID, "Неизвестная команда или текст.")
			t.sendMainMenuInline(chatID)
			continue
		}

		if update.CallbackQuery != nil {
			t.handleCallbackQuery(update.CallbackQuery, userRepo, bookingService, workspaceService)
		}
	}
}

func (t *TelegramAdapter) handleCommands(update tgbotapi.Update, chatID int64, user *entity.User,
	bookingService booking.Service, workspaceService workspace.Service) {

	switch update.Message.Command() {
	case "bookings":
		t.handleMyBookings(chatID, user.ID, bookingService)
	case "book":
		t.handleCreateBooking(chatID, workspaceService)
	default:
		t.SendMessage(chatID, "Unknown command.")
		t.sendMainMenuInline(chatID)
	}
}

func (t *TelegramAdapter) handleMyBookings(chatID int64, userID int64, bookingService booking.Service) {
	bookings, err := bookingService.ListBookingsByUser(context.Background(), userID)
	if err != nil {
		logger.TelegramInfo("Error listing bookings for user %d: %v", userID, err)
		t.SendMessage(chatID, "Error retrieving booking list.")
		t.sendMainMenuInline(chatID)
		return
	}

	var active []entity.Booking
	locationGMT5 := time.FixedZone("GMT+5", 5*60*60)
	now := time.Now().In(locationGMT5)
	for _, b := range bookings {
		if b.Status == entity.BookingStatusConfirmed && b.EndTime.In(locationGMT5).After(now) {
			active = append(active, b)
		}
	}

	if len(active) == 0 {
		t.SendMessage(chatID, "You have no active bookings.")
		t.sendMainMenuInline(chatID)
		return
	}

	t.SendMessage(chatID, "Your active bookings:")

	for _, booking := range active {
		text := fmt.Sprintf(
			"Booking #%d\nDate: %s\nTime: %s — %s\nStatus: %s",
			booking.ID,
			FormatDate(booking.StartTime.In(locationGMT5)),
			FormatTime(booking.StartTime.In(locationGMT5)),
			FormatTime(booking.EndTime.In(locationGMT5)),
			booking.Status,
		)
		button := tgbotapi.NewInlineKeyboardButtonData("Cancel", fmt.Sprintf("cancel:%d", booking.ID))
		keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{button})
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		_, err := t.bot.Send(msg)
		if err != nil {
			logger.TelegramInfo("Error sending booking information for booking #%d: %v", booking.ID, err)
		}
	}
	t.sendMainMenuInline(chatID)
}

func (t *TelegramAdapter) handleCreateBooking(chatID int64, workspaceService workspace.Service) {
	workspaces, err := workspaceService.ListWorkspaces(context.Background())
	if err != nil {
		logger.TelegramInfo("Error listing workspaces: %v", err)
		t.SendMessage(chatID, "Error retrieving workspace list.")
		t.sendMainMenuInline(chatID)
		return
	}

	if len(workspaces) == 0 {
		t.SendMessage(chatID, "No workspaces available for booking.")
		t.sendMainMenuInline(chatID)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, ws := range workspaces {
		btn := tgbotapi.NewInlineKeyboardButtonData(ws.Name, fmt.Sprintf("choosews:%d", ws.ID))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	backButton := tgbotapi.NewInlineKeyboardButtonData("« Back to menu", "back:mainmenu")
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(backButton))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, "Choose a workspace:")
	msg.ReplyMarkup = keyboard
	_, err = t.bot.Send(msg)
	if err != nil {
		logger.TelegramInfo("Error sending workspace selection message: %v", err)
	}
}

func (t *TelegramAdapter) handleCallbackQuery(callback *tgbotapi.CallbackQuery,
	userRepo repository.UserRepository, bookingService booking.Service,
	workspaceService workspace.Service) {

	chatID := callback.Message.Chat.ID
	data := callback.Data
	userID := int64(callback.From.ID)

	t.bot.Send(tgbotapi.NewCallback(callback.ID, ""))

	user, err := userRepo.GetUserByTelegramChatID(context.Background(), userID)
	if err != nil || user == nil {
		logger.TelegramInfo("User with Telegram ID %d not found for callback %s. Error: %v", userID, data, err)
		t.SendMessage(chatID, "You are not registered. Please enter /start your@email.com to register.")
		t.deleteMessage(chatID, callback.Message.MessageID)
		return
	}

	switch data {
	case "menu:mybookings":
		t.handleMyBookings(chatID, user.ID, bookingService)
		t.deleteMessage(chatID, callback.Message.MessageID)
		return
	case "menu:createbooking":
		t.handleCreateBooking(chatID, workspaceService)
		t.deleteMessage(chatID, callback.Message.MessageID)
		return
	case "back:workspaces":
		t.handleCreateBooking(chatID, workspaceService)
		t.deleteMessage(chatID, callback.Message.MessageID)
		return
	case "back:mainmenu":
		t.sendMainMenuInline(chatID)
		t.deleteMessage(chatID, callback.Message.MessageID)
		return
	}

	if strings.HasPrefix(data, "choosews:") {
		t.handleChooseWorkspace(chatID, data, bookingService)
		t.deleteMessage(chatID, callback.Message.MessageID)
	} else if strings.HasPrefix(data, "chooseday:") {
		t.handleChooseDay(chatID, data, bookingService)
		t.deleteMessage(chatID, callback.Message.MessageID)
	} else if strings.HasPrefix(data, "choosehour:") {
		t.handleChooseHour(chatID, data, user.ID, bookingService)
		t.deleteMessage(chatID, callback.Message.MessageID)
	} else if strings.HasPrefix(data, "cancel:") {
		t.handleCancelBooking(chatID, data, user.ID, bookingService)
		t.deleteMessage(chatID, callback.Message.MessageID)
	} else {
		logger.TelegramInfo("Received unknown callback data: %s from user %d", data, userID)
		t.SendMessage(chatID, "Unknown action.")
		t.deleteMessage(chatID, callback.Message.MessageID)
		t.sendMainMenuInline(chatID)
	}
}

func (t *TelegramAdapter) handleChooseWorkspace(chatID int64, data string, bookingService booking.Service) {
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		logger.TelegramInfo("Invalid data format for choosews: %s", data)
		t.SendMessage(chatID, "Error selecting workspace.")
		t.sendMainMenuInline(chatID)
		return
	}

	wsID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		logger.TelegramInfo("Error parsing workspace ID from data '%s': %v", data, err)
		t.SendMessage(chatID, "Invalid workspace ID.")
		t.sendMainMenuInline(chatID)
		return
	}

	locationGMT5 := time.FixedZone("GMT+5", 5*60*60)
	now := time.Now().In(locationGMT5)
	var rows [][]tgbotapi.InlineKeyboardButton

	for i := 0; i < 7; i++ {
		day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, locationGMT5).AddDate(0, 0, i)

		displayDate := day.Format("02 Jan (Mon)")
		dataDate := day.Format("2006-01-02")

		isDayAvailable := !day.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, locationGMT5))

		if isDayAvailable {
			btn := tgbotapi.NewInlineKeyboardButtonData(displayDate, fmt.Sprintf("chooseday:%d:%s", wsID, dataDate))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
		}
	}

	if len(rows) == 0 {
		t.SendMessage(chatID, "Unfortunately, there are no available days for this workspace in the next week.")
		backButton := tgbotapi.NewInlineKeyboardButtonData("« Back to workspaces", "back:workspaces")

		keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(backButton))
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("No available days for workspace %d this week.", wsID))
		msg.ReplyMarkup = keyboard
		_, err = t.bot.Send(msg)
		if err != nil {
			logger.TelegramInfo("Error sending 'no days' message for ws %d: %v", wsID, err)
		}
		return
	}
	backButton := tgbotapi.NewInlineKeyboardButtonData("« Back to workspaces", "back:workspaces")
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(backButton))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Choose a day for workspace %d:", wsID))
	msg.ReplyMarkup = keyboard
	_, err = t.bot.Send(msg)
	if err != nil {
		logger.TelegramInfo("Error sending day selection message for ws %d: %v", wsID, err)
	}
}

func (t *TelegramAdapter) handleChooseDay(chatID int64, data string, bookingService booking.Service) {
	parts := strings.SplitN(data, ":", 3)
	if len(parts) != 3 {
		logger.TelegramInfo("Invalid data format for chooseday: %s", data)
		t.SendMessage(chatID, "Error selecting day.")
		t.sendMainMenuInline(chatID)
		return
	}

	wsID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		logger.TelegramInfo("Error parsing workspace ID from day data '%s': %v", data, err)
		t.SendMessage(chatID, "Invalid workspace ID.")
		t.sendMainMenuInline(chatID)
		return
	}

	dateStr := parts[2]
	locationGMT5 := time.FixedZone("GMT+5", 5*60*60)
	date, err := time.ParseInLocation("2006-01-02", dateStr, locationGMT5)
	if err != nil {
		logger.TelegramInfo("Error parsing date '%s' in GMT+5: %v", dateStr, err)
		t.SendMessage(chatID, "Invalid date.")
		t.sendMainMenuInline(chatID)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	nowInGMT5 := time.Now().In(locationGMT5)
	for h := 8; h < 21; h++ {
		start := time.Date(date.Year(), date.Month(), date.Day(), h, 0, 0, 0, locationGMT5)
		end := start.Add(time.Hour)

		if end.Before(nowInGMT5) {
			continue
		}

		conflicts, err := bookingService.GetConflictingBookings(context.Background(), wsID, start, end)
		if err != nil {
			logger.TelegramInfo("Error checking conflicts for ws %d, slot %s-%s (GMT+5): %v", wsID, start.Format("2006-01-02 15:04"), end.Format("2006-01-02 15:04"), err)
			continue
		}

		isAvailable := conflicts

		if !isAvailable {
			timeStr := start.Format("15:04")
			btn := tgbotapi.NewInlineKeyboardButtonData(timeStr, fmt.Sprintf("choosehour:%d:%s", wsID, start.Format("2006-01-02 15:04")))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
		}
	}

	if len(rows) == 0 {
		backButton := tgbotapi.NewInlineKeyboardButtonData("« Back to days", fmt.Sprintf("choosews:%d", wsID))
		keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(backButton))
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("No available hours on %s for workspace %d.", date.Format("02 Jan 2006"), wsID))
		msg.ReplyMarkup = keyboard
		_, err = t.bot.Send(msg)
		if err != nil {
			logger.TelegramInfo("Error sending 'no slots' message for ws %d, day %s: %v", wsID, dateStr, err)
		}
		return
	}

	backButton := tgbotapi.NewInlineKeyboardButtonData("« Back to days", fmt.Sprintf("choosews:%d", wsID))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(backButton))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Choose a time on %s for workspace %d:", date.Format("02 Jan 2006"), wsID))
	msg.ReplyMarkup = keyboard
	_, err = t.bot.Send(msg)
	if err != nil {
		logger.TelegramInfo("Error sending hour selection message for ws %d, day %s: %v", wsID, dateStr, err)
	}
}

func (t *TelegramAdapter) handleChooseHour(chatID int64, data string, userID int64,
	bookingService booking.Service) {

	parts := strings.SplitN(data, ":", 3)
	if len(parts) != 3 {
		logger.TelegramInfo("Invalid data format for choosehour: %s", data)
		t.SendMessage(chatID, "Error selecting time.")
		t.sendMainMenuInline(chatID)
		return
	}

	wsID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		logger.TelegramInfo("Error parsing workspace ID from hour data '%s': %v", data, err)
		t.SendMessage(chatID, "Invalid workspace ID.")
		t.sendMainMenuInline(chatID)
		return
	}

	startTimeStr := parts[2]
	locationGMT5 := time.FixedZone("GMT+5", 5*60*60)
	start, err := time.ParseInLocation("2006-01-02 15:04", startTimeStr, locationGMT5)
	if err != nil {
		logger.TelegramInfo("Error parsing start time '%s' in GMT+5: %v", startTimeStr, err)
		t.SendMessage(chatID, "Invalid time.")
		t.sendMainMenuInline(chatID)
		return
	}

	end := start.Add(time.Hour)

	_, err = bookingService.CreateBooking(context.Background(), userID, wsID, start, end)
	if err != nil {
		logger.TelegramInfo("Error during booking creation for user %d, ws %d, start %s (GMT+5): %v", userID, wsID, start.Format("2006-01-02 15:04"), err)
		errMsg := fmt.Sprintf("Error creating booking: %v", err)
		if strings.Contains(err.Error(), "not available") {
			errMsg = "Unfortunately, the selected time is no longer available. Please choose another time or day."
		} else if strings.Contains(err.Error(), "in the future") {
			errMsg = "The selected time has already passed. Please choose another time."
		} else if strings.Contains(err.Error(), "booking start time must be between") {
			errMsg = "Bookings are only possible between 8:00 and 21:00."
		} else if strings.Contains(err.Error(), "duration must be exactly one hour") {
			errMsg = "Bookings can only be made for a full hour."
		}

		t.SendMessage(chatID, errMsg)
		t.sendMainMenuInline(chatID)
		return
	}

	t.sendMainMenuInline(chatID)
}

func (t *TelegramAdapter) handleCancelBooking(chatID int64, data string, userID int64,
	bookingService booking.Service) {

	bookingIDStr := strings.TrimPrefix(data, "cancel:")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		logger.TelegramInfo("Error parsing booking ID from cancel data '%s': %v", data, err)
		t.SendMessage(chatID, "Error canceling booking: invalid booking ID.")
		t.sendMainMenuInline(chatID)
		return
	}

	penalty, err := bookingService.CancelBooking(context.Background(), userID, bookingID)
	if err != nil {
		logger.TelegramInfo("Error during booking cancellation for user %d, booking %d: %v", userID, bookingID, err)
		errMsg := fmt.Sprintf("Error canceling booking: %v", err)
		if strings.Contains(err.Error(), "not found") {
			errMsg = "Booking not found or already deleted."
		} else if strings.Contains(err.Error(), "forbidden") {
			errMsg = "You can only cancel your own bookings."
		} else if strings.Contains(err.Error(), "already cancelled") {
			errMsg = "This booking has already been canceled."
		} else {
			errMsg = "Failed to cancel booking."
		}

		t.SendMessage(chatID, errMsg)
		t.sendMainMenuInline(chatID)
		return
	}

	cancelConfirmText := "✅ Booking canceled successfully!"
	if penalty > 0 {
		cancelConfirmText = fmt.Sprintf("✅ Booking canceled successfully!\nPenalty applied: %.2f", penalty)
	}

	t.SendMessage(chatID, cancelConfirmText)
	t.sendMainMenuInline(chatID)
}

func (t *TelegramAdapter) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := t.bot.Send(msg)
	if err != nil {
		logger.TelegramInfo("Error sending message to chat %d: %v", chatID, err)
	}
}

func FormatDate(t time.Time) string {
	return t.Format("02 Jan 2006")
}

func FormatTime(t time.Time) string {
	return t.Format("15:04")
}

func (t *TelegramAdapter) deleteMessage(chatID int64, messageID int) {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := t.bot.Request(deleteMsg)
	if err != nil {
		logger.TelegramInfo("Error deleting message in chat %d: %v", chatID, err)
	}
}
