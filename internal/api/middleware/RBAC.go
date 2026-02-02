package middleware

import (
	"database/sql"
	"net/http"

	"github.com/Zolet-hash/smart-rentals/internal/database"
	"github.com/gin-gonic/gin"
)

func RequireRole(db *database.Database, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get user ID from context (set by JWT auth middleware)
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized: user id not found in context",
			})
			return
		}

		var userID uint
		switch v := userIDValue.(type) {
		case uint:
			userID = v
		case float64:
			userID = uint(v)
		case int:
			userID = uint(v)
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user id type",
			})
			return
		}

		// Fetch user role only (fast path)
		var role string
		err := db.DB.QueryRowContext(
			c.Request.Context(),
			`SELECT role FROM users WHERE id = $1`,
			userID,
		).Scan(&role)

		if err == sql.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "user not found",
			})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
			})
			return
		}

		// Check role
		if role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden: insufficient permissions",
			})
			return
		}

		// Store minimal user info in context
		c.Set("userID", userID)
		c.Set("role", role)

		c.Next()
	}
}
