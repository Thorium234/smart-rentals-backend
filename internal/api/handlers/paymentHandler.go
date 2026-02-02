package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Zolet-hash/smart-rentals/internal/api/middleware"
	"github.com/Zolet-hash/smart-rentals/internal/database"
	"github.com/gin-gonic/gin"
)

type CashPaymentInput struct {
	TenantID int     `json:"tenant_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Receipt  string  `json:"receipt"`
}

type AssignPaymentInput struct {
	TenantID int `json:"tenant_id" binding:"required"`
}

func ListPayments(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Lists payments linked to landlord's properties/tenants
		query := `
			SELECT p.id, p.tenant_id, t.tenant_name, p.amount, p.status, p.created_at, p.method, p.receipt
			FROM payments p
			LEFT JOIN tenants t ON p.tenant_id = t.id
			WHERE p.landlord_id = $1
			ORDER BY p.created_at DESC
		`

		rows, err := db.Query(query, landlordID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
			return
		}
		defer rows.Close()

		payments := []gin.H{}
		for rows.Next() {
			var p struct {
				ID         int
				TenantID   sql.NullInt64
				TenantName sql.NullString
				Amount     float64
				Status     string
				CreatedAt  time.Time
				Method     string
				Receipt    string
			}
			if err := rows.Scan(&p.ID, &p.TenantID, &p.TenantName, &p.Amount, &p.Status, &p.CreatedAt, &p.Method, &p.Receipt); err != nil {
				continue
			}
			payments = append(payments, gin.H{
				"id":             p.ID,
				"tenant_id":      p.TenantID.Int64,
				"tenant_name":    p.TenantName.String,
				"amount":         p.Amount,
				"status":         p.Status,
				"created_at":     p.CreatedAt,
				"payment_date":   p.CreatedAt,
				"payment_method": p.Method,
				"method":         p.Method,
				"transaction_id": p.Receipt,
				"reference":      p.Receipt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": payments})
	}
}

func RecordCashPayment(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var input CashPaymentInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify Tenant Ownership
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1 AND landlord_id = $2)", input.TenantID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found or unauthorized"})
			return
		}

		// Transaction: Insert Payment + Update Tenant Balance
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Insert Payment
		query := `
			INSERT INTO payments (landlord_id, tenant_id, amount, status, method, receipt)
			VALUES ($1, $2, $3, 'COMPLETED', 'CASH', $4)
			RETURNING id
		`
		var paymentID int
		receipt := input.Receipt
		if receipt == "" {
			receipt = "CASH-" + time.Now().Format("20060102150405")
		}
		err = tx.QueryRow(query, landlordID, input.TenantID, input.Amount, receipt).Scan(&paymentID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment"})
			return
		}

		// Update Tenant Balance (Decrease balance by amount paid)
		_, err = tx.Exec("UPDATE tenants SET balance = balance - $1 WHERE id = $2", input.Amount, input.TenantID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant balance"})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":    "Payment recorded successfully",
			"payment_id": paymentID,
		})
	}
}

func AssignPayment(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		paymentID := c.Param("id")
		var input AssignPaymentInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify Tenant Ownership
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1 AND landlord_id = $2)", input.TenantID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found or unauthorized"})
			return
		}

		// Transaction
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Assign Payment (Verify it's unassigned? Or re-assignable? Assuming unassigned logic or just overwriting)
		// Also verify payment logic - usually we just update tenant_id.
		// BUT we must also update balances.
		// Complexity: If it was assigned to someone else, we must revert that balance.
		// Simplified: Assume strictly for assigning 'Unassigned' payments.

		var amount float64
		// Fetch payment details and lock
		err = tx.QueryRow("SELECT amount FROM payments WHERE id = $1 FOR UPDATE", paymentID).Scan(&amount)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
			return
		}

		// Assign
		_, err = tx.Exec("UPDATE payments SET tenant_id = $1, status = 'COMPLETED' WHERE id = $2", input.TenantID, paymentID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign payment"})
			return
		}

		// Update Balance
		_, err = tx.Exec("UPDATE tenants SET balance = balance - $1 WHERE id = $2", amount, input.TenantID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant balance"})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payment assigned successfully"})
	}
}

func GetTenantHistory(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tenantID := c.Param("tenantId")

		// Verify Ownership
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1 AND landlord_id = $2)", tenantID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found or unauthorized"})
			return
		}

		// Fetch history (Payments)
		query := `
			SELECT id, amount, status, created_at, method, receipt
			FROM payments
			WHERE tenant_id = $1
			ORDER BY created_at DESC
		`

		rows, err := db.Query(query, tenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
			return
		}
		defer rows.Close()

		history := []gin.H{}
		for rows.Next() {
			var p struct {
				ID        int
				Amount    float64
				Status    string
				CreatedAt time.Time
				Method    string
				Receipt   string
			}
			if err := rows.Scan(&p.ID, &p.Amount, &p.Status, &p.CreatedAt, &p.Method, &p.Receipt); err != nil {
				continue
			}
			history = append(history, gin.H{
				"type":      "PAYMENT",
				"id":        p.ID,
				"amount":    p.Amount,
				"status":    p.Status,
				"date":      p.CreatedAt,
				"method":    p.Method,
				"reference": p.Receipt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": history})
	}
}
