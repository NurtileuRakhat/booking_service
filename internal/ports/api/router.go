package api

import (
	"booking/internal/ports/api/auth"
	"booking/internal/ports/api/booking"
	"booking/internal/ports/api/middleware"
	"booking/internal/ports/api/workspace"
	"github.com/gin-gonic/gin"
)

func SetupRouter(bookingHandler *booking.BookingHandler, workspaceHandler *workspace.WorkspaceHandler, authHandler *auth.AuthHandler) *gin.Engine {
	r := gin.Default()

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.Refresh)

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("/bookings", bookingHandler.CreateBooking)
		api.GET("/bookings", bookingHandler.ListBookings)
		api.DELETE("/bookings/:id", bookingHandler.CancelBooking)
		api.GET("/workspaces", workspaceHandler.ListWorkspaces)
		api.POST("/workspaces", workspaceHandler.CreateWorkspace)
	}
	return r
}
