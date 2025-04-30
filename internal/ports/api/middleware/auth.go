package middleware

import (
	"booking/internal/entity"
	"booking/pkg/logger"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"os"
	"strings"
	"time"
)

var secretKey []byte = []byte(os.Getenv("JWT_SECRET"))

func GenerateAccessToken(user *entity.User) (string, error) {
	claims := &jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 240).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	logger.Info("Generated access token for user %d", user.ID)
	return token.SignedString(secretKey)
}

func GenerateRefreshToken(user *entity.User) (string, error) {
	claims := &jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 240).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	logger.Info("Generated refresh token for user %d", user.ID)
	return token.SignedString(secretKey)
}

func ParseRefreshToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error("Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		logger.Info("Parsed refresh token")
		return secretKey, nil
	})
}

func ParseAccessToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error("Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		logger.Info("Parsed access token")
		return secretKey, nil
	})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			logger.Error("Authorization header is missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		token, err := ParseAccessToken(tokenString)
		if err != nil || !token.Valid {
			logger.Error("Invalid token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Error("Invalid token claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if ok {
			logger.Info("User ID found in token: %d", userID)
			c.Set("user_id", fmt.Sprintf("%.0f", userID))
		} else if email, ok := claims["email"].(string); ok {
			logger.Info("Email found in token: %s", email)
			c.Set("email", email)
		}
		c.Next()
	}
}

func ValidateRefreshToken(tokenString string) (string, error) {
	token, err := ParseRefreshToken(tokenString)
	if err != nil {
		logger.Error("Invalid refresh token: %v", err)
		return "", err
	}

	if !token.Valid {
		logger.Error("Invalid refresh token")
		return "", fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error("Invalid token claims")
		return "", fmt.Errorf("invalid token claims")
	}

	email, ok := claims["email"].(string)
	if !ok {
		logger.Error("Email claim not found")
		return "", fmt.Errorf("email claim not found")
	}

	return email, nil
}
