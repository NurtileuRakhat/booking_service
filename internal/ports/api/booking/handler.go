package booking

import (
	"booking/internal/usecase/booking"
	"booking/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type BookingHandler struct {
	service booking.Service
}

func NewBookingHandler(service booking.Service) *BookingHandler {
	return &BookingHandler{service: service}
}

type CreateBookingRequest struct {
	WorkspaceID int64  `json:"workspace_id" binding:"required"`
	StartTime   string `json:"start_time" binding:"required"`
	EndTime     string `json:"end_time" binding:"required"`
}

type CreateBookingResponse struct {
	ID int64 `json:"id"`
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}
	userID, _ := strconv.ParseInt(c.GetString("user_id"), 10, 64) // JWT middleware sets user_id
	start, err1 := time.Parse(time.RFC3339, req.StartTime)
	end, err2 := time.Parse(time.RFC3339, req.EndTime)
	if err1 != nil || err2 != nil {
		logger.Error("Invalid datetime format: %v", err1)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid datetime format"})
		return
	}
	booking, err := h.service.CreateBooking(c.Request.Context(), userID, req.WorkspaceID, start, end)
	if err != nil {
		switch err.Error() {
		case "user not found", "workspace not found":
			logger.Error("User or workspace not found: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "workspace is not available for the selected time":
			logger.Error("Workspace is not available for the selected time: %v", err)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case "invalid user or workspace id", "start_time must be before end_time", "start_time must be in the future", "invalid input", "invalid datetime format":
			logger.Error("Invalid input: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			logger.Error("Internal server error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	logger.Info("Booking created successfully: %d", booking.ID)
	c.JSON(http.StatusCreated, CreateBookingResponse{ID: booking.ID})
}

func (h *BookingHandler) ListBookings(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("user_id"), 10, 64)
	bookings, err := h.service.ListBookingsByUser(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to list bookings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("user_id"), 10, 64)
	bookingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error("Invalid booking id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}
	penalty, err := h.service.CancelBooking(c.Request.Context(), userID, bookingID)
	if err != nil {
		if err.Error() == "booking not found" {
			logger.Error("Booking not found: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "forbidden" {
			logger.Error("Forbidden: %v", err)
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "already cancelled" {
			logger.Error("Already cancelled: %v", err)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		logger.Error("Internal server error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("Booking cancelled successfully: %d", bookingID)
	c.JSON(http.StatusOK, gin.H{"penalty": penalty})
}
