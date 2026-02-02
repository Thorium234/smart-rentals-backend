package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Zolet-hash/smart-rentals/internal/api/middleware"
	"github.com/Zolet-hash/smart-rentals/internal/database"
	"github.com/gin-gonic/gin"
)

type CreateTenantInput struct {
	TenantName string  `json:"tenant_name" binding:"required"`
	PaymentNo1 string  `json:"payment_no1" binding:"required"`
	PaymentNo2 string  `json:"payment_no2"`
	Rent       float64 `json:"rent" binding:"required"`
}

type UpdateTenantInput struct {
	TenantName *string `json:"tenant_name"`
	PaymentNo1 *string `json:"payment_no1"`
	PaymentNo2 *string `json:"payment_no2"`
}

func CreateTenant(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		unitID := c.Param("unitId")
		var input CreateTenantInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Enforce Ownership: Unit -> Property -> Landlord
		var exists bool
		queryCheck := `
			SELECT EXISTS(
				SELECT 1 
				FROM units u
				JOIN properties p ON u.property_id = p.id
				WHERE u.id = $1 AND p.landlord_id = $2
			)`
		err = db.QueryRow(queryCheck, unitID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found or unauthorized"})
			return
		}

		// Transaction: Insert Tenant + Update Unit Vacancy
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Insert Tenant (Balance initialized to 0 or Rent? Spec said "Balance initialized = rent", logic usually assumes 0 and first month invoice adds it. Let's stick to 0 or explicit logic. The spec says "Balance initialized = rent". I will follow spec.)
		// Spec: "Balance initialized = rent"
		insertQuery := `
			INSERT INTO tenants (unit_id, landlord_id, tenant_name, payment_no1, payment_no2, rent, balance)
			VALUES ($1, $2, $3, $4, $5, $6, $6)
			RETURNING id, created_at
		`
		var tenantID int
		var createdAt string
		err = tx.QueryRow(insertQuery, unitID, landlordID, input.TenantName, input.PaymentNo1, input.PaymentNo2, input.Rent).Scan(&tenantID, &createdAt)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
			return
		}

		// Update Unit Vacancy to false
		_, err = tx.Exec("UPDATE units SET vacancy = false WHERE id = $1", unitID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update unit vacancy"})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Tenant onboarded successfully",
			"data": gin.H{
				"id":          tenantID,
				"unit_id":     unitID,
				"tenant_name": input.TenantName,
				"balance":     input.Rent, // Initial balance
				"created_at":  createdAt,
			},
		})
	}
}

func ListAllTenants(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Optional filter by unitId if route is /units/:unitId/tenants
		unitID := c.Param("unitId")

		query := `
			SELECT t.id, t.tenant_name, t.payment_no1, t.payment_no2, t.rent, t.balance, u.unit_name, p.title as property_title
			FROM tenants t
			JOIN units u ON t.unit_id = u.id
			JOIN properties p ON u.property_id = p.id
			WHERE t.landlord_id = $1
		`
		args := []interface{}{landlordID}

		if unitID != "" {
			query += " AND t.unit_id = $2"
			args = append(args, unitID)
		}

		query += " ORDER BY t.created_at DESC"

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenants"})
			return
		}
		defer rows.Close()

		tenants := []gin.H{}
		for rows.Next() {
			var t struct {
				ID            int
				TenantName    string
				PaymentNo1    string
				PaymentNo2    sql.NullString
				Rent          float64
				Balance       float64
				UnitName      string
				PropertyTitle string
			}
			if err := rows.Scan(&t.ID, &t.TenantName, &t.PaymentNo1, &t.PaymentNo2, &t.Rent, &t.Balance, &t.UnitName, &t.PropertyTitle); err != nil {
				continue
			}
			tenants = append(tenants, gin.H{
				"id":             t.ID,
				"tenant_name":    t.TenantName,
				"payment_no1":    t.PaymentNo1,
				"payment_no2":    t.PaymentNo2.String,
				"rent":           t.Rent,
				"balance":        t.Balance,
				"unit_name":      t.UnitName,
				"property_title": t.PropertyTitle,
			})
		}

		c.JSON(http.StatusOK, tenants) // Return array directly as per typical REST list
	}
}

func GetTenant(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tenantID := c.Param("tenantId")

		query := `
			SELECT t.id, t.tenant_name, t.payment_no1, t.payment_no2, t.rent, t.balance, t.unit_id, u.unit_name
			FROM tenants t
			JOIN units u ON t.unit_id = u.id
			WHERE t.id = $1 AND t.landlord_id = $2
		`

		var t struct {
			ID         int
			TenantName string
			PaymentNo1 string
			PaymentNo2 sql.NullString
			Rent       float64
			Balance    float64
			UnitID     int
			UnitName   string
		}

		err = db.QueryRow(query, tenantID, landlordID).Scan(&t.ID, &t.TenantName, &t.PaymentNo1, &t.PaymentNo2, &t.Rent, &t.Balance, &t.UnitID, &t.UnitName)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found or unauthorized"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenant"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":          t.ID,
			"tenant_name": t.TenantName,
			"payment_no1": t.PaymentNo1,
			"payment_no2": t.PaymentNo2.String,
			"rent":        t.Rent,
			"balance":     t.Balance,
			"unit_id":     t.UnitID,
			"unit_name":   t.UnitName,
		})
	}
}

func UpdateTenant(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tenantID := c.Param("tenantId")
		var input UpdateTenantInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify ownership
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1 AND landlord_id = $2)", tenantID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found or unauthorized"})
			return
		}

		// Dynamic update
		query := "UPDATE tenants SET updated_at = NOW()"
		args := []interface{}{}
		argID := 1

		if input.TenantName != nil {
			query += ", tenant_name = $" + strconv.Itoa(argID)
			args = append(args, *input.TenantName)
			argID++
		}
		if input.PaymentNo1 != nil {
			query += ", payment_no1 = $" + strconv.Itoa(argID)
			args = append(args, *input.PaymentNo1)
			argID++
		}
		if input.PaymentNo2 != nil {
			query += ", payment_no2 = $" + strconv.Itoa(argID)
			args = append(args, *input.PaymentNo2)
			argID++
		}

		query += " WHERE id = $" + strconv.Itoa(argID) + " AND landlord_id = $" + strconv.Itoa(argID+1)
		args = append(args, tenantID, landlordID)

		_, err = db.Exec(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tenant updated successfully"})
	}
}

func RemoveTenant(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tenantID := c.Param("tenantId")

		// Transaction: Delete Tenant + Set Unit Vacancy to true
		// (Assuming Hard Delete for now as per "RemoveTenant" but typical apps use soft delete.
		// Spec: "Soft delete recommended... Or hard delete with cascade". I'll do hard delete + vacancy update for simplicity unless soft delete columns exist).

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Get UnitID to free it up
		var unitID int
		err = tx.QueryRow("SELECT unit_id FROM tenants WHERE id = $1 AND landlord_id = $2", tenantID, landlordID).Scan(&unitID)
		if err == sql.ErrNoRows {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found or unauthorized"})
			return
		} else if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenant details"})
			return
		}

		// Delete Tenant
		_, err = tx.Exec("DELETE FROM tenants WHERE id = $1", tenantID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant"})
			return
		}

		// Update Unit Vacancy
		_, err = tx.Exec("UPDATE units SET vacancy = true WHERE id = $1", unitID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update unit vacancy"})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction commit failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tenant removed and unit vacated"})
	}
}
