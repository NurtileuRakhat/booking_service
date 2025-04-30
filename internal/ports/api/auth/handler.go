package auth

import (
	"booking/internal/entity"
	"booking/internal/ports/api/middleware"
	"booking/internal/ports/repository"
	"booking/pkg/logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

type AuthHandler struct {
	userRepo repository.UserRepository
}

type Response struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func NewAuthHandler(userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username" binding:"required"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Password hash error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := &entity.User{
		Email:        email,
		Username:     req.Username,
		Name:         req.Name,
		Surname:      req.Surname,
		PasswordHash: string(hash),
	}

	_, err = h.userRepo.CreateUser(c.Request.Context(), user)
	if err != nil {
		logger.Error("Registration error: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	logger.Info("User registered: %s", user.Email)
	c.JSON(http.StatusCreated, gin.H{"message": "registration successful"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	logger.Info("Login attempt: %s", req.Email)
	user, err := h.userRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		logger.Error("User not found for email: %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user data"})
		return
	}

	logger.Info("Login attempt: %s", req.Email)
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		logger.Error("Password mismatch for: %s", req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	accessToken, err := middleware.GenerateAccessToken(user)
	if err != nil {
		logger.Error("Could not generate access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate access token"})
		return
	}

	refreshToken, err := middleware.GenerateRefreshToken(user)
	if err != nil {
		logger.Error("Could not generate refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate refresh token"})
		return
	}

	c.JSON(http.StatusOK, Response{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
		return
	}

	refreshToken := parts[1]

	email, err := middleware.ValidateRefreshToken(refreshToken)
	if err != nil {
		logger.Error("Invalid refresh token: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	user, err := h.userRepo.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		logger.Error("User not found for email: %s", email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	newAccessToken, err := middleware.GenerateAccessToken(user)
	if err != nil {
		logger.Error("Could not generate access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate access token"})
		return
	}

	c.JSON(http.StatusOK, RefreshTokenResponse{
		AccessToken: newAccessToken,
	})
}
