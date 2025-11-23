package middleware

import (
	"net/http"
	"shopping-cart/config"
	"shopping-cart/models"
	"strings"
	"time" // This line is already present, so we don't need to remove it.

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the user token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token by looking up in sessions table
		var session models.Session
		if err := config.DB.Where("token = ?", token).First(&session).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Optional: check expiry
		if session.ExpiresAt != nil {
			// if expired
			if session.ExpiresAt.Before(time.Now()) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
				c.Abort()
				return
			}
		}

		var user models.User
		if err := config.DB.First(&user, session.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session user"})
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("user", user)
		c.Next()
	}
}
