package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Zolet-hash/smart-rentals/internal/api/middleware"
	"github.com/Zolet-hash/smart-rentals/internal/database"
	"github.com/Zolet-hash/smart-rentals/internal/models"
	"github.com/Zolet-hash/smart-rentals/internal/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	db        *database.Database
	jwtSecret []byte
	// Add token expiration configuration
	tokenExpiration time.Duration
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *database.Database, jwtSecret []byte) *AuthHandler {
	return &AuthHandler{
		db:              db,
		jwtSecret:       jwtSecret,
		tokenExpiration: 24 * time.Hour, // Default 24 hour expiration
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var user models.UserRegister

	// Validate input JSON
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input format",
			"details": err.Error(),
		})
		return
	}

	// Additional validations
	if err := user.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := utils.ValidatePassword(user.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var exists bool
	err := h.db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
		user.Email).Scan(&exists)
	if err != nil {
		reqID, _ := c.Get("request_id")
		log.Printf("[%v] register: select exists error: %v", reqID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Database error",
			"trace_id": reqID,
		})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password processing failed"})
		return
	}

	// Insert user with transaction
	tx, err := h.db.DB.Begin() //group multiple database operations
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction start failed"})
		return
	}

	var id int
	err = tx.QueryRow(`
        INSERT INTO users (email, password_hash, full_name, phone, role) 
        VALUES ($1, $2, $3, $4, $5) 
        RETURNING id`,
		user.Email, hashedPassword, user.FullName, user.Phone, user.Role,
	).Scan(&id)

	if err != nil {
		tx.Rollback()
		reqID, _ := c.Get("request_id")
		log.Printf("[%v] register: insert user failed: %v", reqID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "User creation failed",
			"trace_id": reqID,
		})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": id,
	})
}

// Login handles user authentication and JWT generation
func (h *AuthHandler) Login(c *gin.Context) {
	var login models.UserLogin
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login data"})
		return
	}

	// Get user from database
	var user models.User
	err := h.db.DB.QueryRow(`
        SELECT id, email, password_hash, role 
        FROM users 
        WHERE email = $1`,
		login.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if err == sql.ErrNoRows {
		// Don't specify whether email or password was wrong
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		reqID, _ := c.Get("request_id")
		log.Printf("[%v] login: db error: %v", reqID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Internal server error",
			"trace_id": reqID,
		})
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(login.Password, user.PasswordHash) {
		// Use same message as above for security
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT with claims
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"iat":     now.Unix(),
		"exp":     now.Add(h.tokenExpiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	// Return token with expiration
	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"expires_in": h.tokenExpiration.Seconds(),
		"token_type": "Bearer",
		"role":       user.Role,
	})

}

// RefreshToken generates a new token for valid users
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Generate new token
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"iat":     now.Unix(),
		"exp":     now.Add(h.tokenExpiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token refresh failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      tokenString,
		"expires_in": h.tokenExpiration.Seconds(),
		"token_type": "Bearer",
	})
}

// Logout endpoint (optional - useful for client-side cleanup)
func (h *AuthHandler) Logout(c *gin.Context) {
	// Since JWT is stateless, server-side logout isn't needed
	// However, we can return instructions for the client
	c.JSON(http.StatusOK, gin.H{
		"message":      "Successfully logged out",
		"instructions": "Please remove the token from your client storage",
	})
}

// ListUsers returns all users in the system
func (h *AuthHandler) ListUsers(c *gin.Context) {
	query := `
		SELECT id, email, full_name, phone, role, created_at 
		FROM users 
		ORDER BY created_at DESC
	`
	rows, err := h.db.DB.Query(query)
	if err != nil {
		reqID, _ := c.Get("request_id")
		log.Printf("[%v] listUsers: db error: %v", reqID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	users := []gin.H{}
	for rows.Next() {
		var u struct {
			ID        int
			Email     string
			FullName  string
			Phone     string
			Role      string
			CreatedAt time.Time
		}
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.Phone, &u.Role, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, gin.H{
			"id":         u.ID,
			"email":      u.Email,
			"full_name":  u.FullName,
			"phone":      u.Phone,
			"role":       u.Role,
			"created_at": u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

// UpdateUser allows an admin to modify user details
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var input struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	query := `
		UPDATE users 
		SET email = $1, full_name = $2, phone = $3, role = $4, updated_at = NOW() 
		WHERE id = $5
	`
	_, err := h.db.DB.Exec(query, input.Email, input.FullName, input.Phone, input.Role, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser removes a user from the system
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, _ := strconv.Atoi(userIDStr)

	// Pre-check: Don't allow an admin to delete themselves via this endpoint (prevent lockout)
	currentUserID, err := middleware.GetUserID(c)
	if err == nil && userID == currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Security violation: You cannot delete your own administrative account."})
		return
	}

	_, err = h.db.DB.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ResetPassword allows an admin to reset a user's password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input struct {
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password is required"})
		return
	}

	// Validate password strength
	if err := utils.ValidatePassword(input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var exists bool
	err = h.db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		reqID, _ := c.Get("request_id")
		log.Printf("[%v] resetPassword: db error checking user: %v", reqID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Database error",
			"trace_id": reqID,
		})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		reqID, _ := c.Get("request_id")
		log.Printf("[%v] resetPassword: password hashing failed: %v", reqID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password processing failed"})
		return
	}

	// Update password in database
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	result, err := h.db.DB.Exec(query, hashedPassword, userID)
	if err != nil {
		reqID, _ := c.Get("request_id")
		log.Printf("[%v] resetPassword: db update failed: %v", reqID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Failed to reset password",
			"trace_id": reqID,
		})
		return
	}

	// Verify the update affected exactly one row
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password reset failed"})
		return
	}

	log.Printf("Admin reset password for user ID: %d", userID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
		"user_id": userID,
	})
}
