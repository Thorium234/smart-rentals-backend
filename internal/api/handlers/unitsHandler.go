package handlers

import (
	"net/http"

	"github.com/Zolet-hash/smart-rentals/internal/api/middleware"
	"github.com/Zolet-hash/smart-rentals/internal/database"
	"github.com/gin-gonic/gin"
)

type CreateUnitInput struct {
	UnitName  string  `json:"unit_name" binding:"required"`
	UnitType  string  `json:"unit_type" binding:"required"`
	UnitPrice float64 `json:"unit_price" binding:"required"`
}

func CreateUnit(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		propertyID := c.Param("propertyId")
		var input CreateUnitInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Enforce Ownership: Property must belong to landlord
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM properties WHERE id = $1 AND landlord_id = $2)", propertyID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Property not found or unauthorized"})
			return
		}

		// Insert Unit
		query := `
			INSERT INTO units (property_id, unit_name, unit_type, unit_price, vacancy)
			VALUES ($1, $2, $3, $4, true)
			RETURNING id, created_at
		`
		var unitID int
		var createdAt string
		err = db.QueryRow(query, propertyID, input.UnitName, input.UnitType, input.UnitPrice).Scan(&unitID, &createdAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create unit"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Unit created successfully",
			"data": gin.H{
				"id":          unitID,
				"property_id": propertyID,
				"unit_name":   input.UnitName,
				"unit_type":   input.UnitType,
				"unit_price":  input.UnitPrice,
				"vacancy":     true,
				"created_at":  createdAt,
			},
		})
	}
}

func GetUnitsByProperty(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		propertyID := c.Param("propertyId")

		// Verify ownership of property first
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM properties WHERE id = $1 AND landlord_id = $2)", propertyID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Property not found or unauthorized"})
			return
		}

		query := `
			SELECT id, unit_name, unit_type, unit_price, vacancy
			FROM units
			WHERE property_id = $1
			ORDER BY unit_name ASC
		`

		rows, err := db.Query(query, propertyID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
			return
		}
		defer rows.Close()

		units := []gin.H{}
		for rows.Next() {
			var u struct {
				ID        int
				UnitName  string
				UnitType  string
				UnitPrice float64
				Vacancy   bool
			}
			if err := rows.Scan(&u.ID, &u.UnitName, &u.UnitType, &u.UnitPrice, &u.Vacancy); err != nil {
				continue
			}
			units = append(units, gin.H{
				"id":         u.ID,
				"unit_name":  u.UnitName,
				"unit_type":  u.UnitType,
				"unit_price": u.UnitPrice,
				"vacancy":    u.Vacancy,
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": units})
	}
}

type UpdateUnitInput struct {
	UnitName  *string  `json:"unit_name"`
	UnitType  *string  `json:"unit_type"`
	UnitPrice *float64 `json:"unit_price"`
}

func UpdateUnit(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		propertyID := c.Param("propertyId")
		unitID := c.Param("unitId")

		var input UpdateUnitInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify ownership: Unit -> Property -> Landlord
		var exists bool
		query := `
			SELECT EXISTS(
				SELECT 1 FROM units u
				JOIN properties p ON u.property_id = p.id
				WHERE u.id = $1 AND p.id = $2 AND p.landlord_id = $3
			)`
		err = db.QueryRow(query, unitID, propertyID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found or unauthorized"})
			return
		}

		// Update logic... (dynamic query like others)
		// For simplicity, let's just do a direct update since it's a small object
		_, err = db.Exec(`
			UPDATE units SET 
				unit_name = COALESCE($1, unit_name), 
				unit_type = COALESCE($2, unit_type), 
				unit_price = COALESCE($3, unit_price),
				updated_at = NOW()
			WHERE id = $4`,
			input.UnitName, input.UnitType, input.UnitPrice, unitID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update unit"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Unit updated successfully"})
	}
}

func DeleteUnit(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		propertyID := c.Param("propertyId")
		unitID := c.Param("unitId")

		// Verify ownership
		var exists bool
		query := `
			SELECT EXISTS(
				SELECT 1 FROM units u
				JOIN properties p ON u.property_id = p.id
				WHERE u.id = $1 AND p.id = $2 AND p.landlord_id = $3
			)`
		err = db.QueryRow(query, unitID, propertyID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found or unauthorized"})
			return
		}

		_, err = db.Exec("DELETE FROM units WHERE id = $1", unitID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete unit. Ensure it has no active tenants."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Unit deleted successfully"})
	}
}
