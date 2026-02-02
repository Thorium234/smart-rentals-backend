package middleware

import (
	"crypto/rand"
	"fmt"

	"github.com/gin-gonic/gin"
)

// RequestID middleware generates a unique ID for each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate simple ID
		b := make([]byte, 6)
		if _, err := rand.Read(b); err != nil {
			// Fallback to timestamp if random fails
			c.Set("request_id", fmt.Sprintf("req-%d", c.Request.Context().Value("timestamp")))
		} else {
			id := fmt.Sprintf("req_%x", b)
			c.Writer.Header().Set("X-Request-Id", id)
			c.Set("request_id", id)
		}
		c.Next()
	}
}
