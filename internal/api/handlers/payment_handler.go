package handlers

import (
	"net/http"

	"github.com/Zolet-hash/smart-rentals/internal/services"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	Service *services.PaymentService
}

func NewPaymentHandler(service *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{Service: service}
}

// C2BValidation - Safaricom sends a request here to validate the transaction
func (h *PaymentHandler) C2BValidation(c *gin.Context) {
	// We should always accept payment in this case, unless we want to blacklist specific MSISDNs.
	// We just return a success response immediately.
	c.JSON(http.StatusOK, gin.H{
		"ResultCode": 0,
		"ResultDesc": "Accepted",
	})
}

// C2BConfirmation - Safaricom sends the actual payment details here
func (h *PaymentHandler) C2BConfirmation(c *gin.Context) {
	var payload services.C2BConfirmationPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Process in background or foreground? Foreground is safer for acknowledgment.
	// Safaricom expects a response within a few seconds.
	// We process it synchronously for now.
	err := h.Service.ProcessCallback(payload)
	if err != nil {
		// Log error but still return success to Safaricom so they stop retrying?
		// If we return failure, they might retry. If it's a logic error (e.g. DB down), we might want retry.
		// If it's a duplicate, we already handled it in service to return nil.
		// If it's unknown shortcode, we might want to flag it.
		// For now, let's log and return success to avoid queue backup at Safaricom side,
		// unless it's a transient error.
		// But for this implementation, we return success.
		c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Received"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ResultCode": 0,
		"ResultDesc": "Success",
	})
}

type UpdateConfigRequest struct {
	ShortCode         string `json:"short_code" binding:"required"`
	ShortCodeType     string `json:"short_code_type" binding:"required"` // "paybill" or "till"
	ConsumerKey       string `json:"consumer_key" binding:"required"`
	ConsumerSecret    string `json:"consumer_secret" binding:"required"`
	Environment       string `json:"environment" binding:"required,oneof=sandbox production"`
	ValidationEnabled bool   `json:"validation_enabled"`
}

// UpdateConfig - Landlord saves their M-Pesa keys
func (h *PaymentHandler) UpdateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get Landlord ID from context (set by auth middleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var landlordID uint
	switch v := userIDVal.(type) {
	case float64:
		landlordID = uint(v)
	case int:
		landlordID = uint(v)
	case uint:
		landlordID = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// Base URL for callbacks
	baseURL := "https://" + c.Request.Host

	err := h.Service.SaveLandlordConfig(
		landlordID,
		req.ShortCode,
		req.ShortCodeType,
		req.ConsumerKey,
		req.ConsumerSecret,
		req.Environment,
		req.ValidationEnabled,
		baseURL,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration saved and URLs registered successfully"})
}
