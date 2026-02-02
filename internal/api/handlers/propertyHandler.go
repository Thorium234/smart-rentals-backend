package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Zolet-hash/smart-rentals/internal/api/middleware"
	"github.com/Zolet-hash/smart-rentals/internal/database"
	"github.com/gin-gonic/gin"
)

type CreatePropertyInput struct {
	Title        string  `json:"title" binding:"required"`
	Description  string  `json:"description"`
	Location     string  `json:"location" binding:"required"`
	PropertyType string  `json:"property_type" binding:"required"`
	Vacancy      bool    `json:"vacancy"`
	TotalRent    float64 `json:"total_rent"`
}

type UpdatePropertyInput struct {
	Title        *string  `json:"title"`
	Description  *string  `json:"description"`
	Location     *string  `json:"location"`
	PropertyType *string  `json:"property_type"`
	Vacancy      *bool    `json:"vacancy"`
	TotalRent    *float64 `json:"total_rent"`
}

func CreateProperty(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get landlord from context
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// 2. Bind + validate input
		var input CreatePropertyInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 3. Execute query (scoped by landlord)
		query := `
			INSERT INTO properties (landlord_id, title, description, location, property_type, vacancy, total_rent)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at
		`

		var propertyID int
		var createdAt, updatedAt string
		err = db.QueryRow(query, landlordID, input.Title, input.Description, input.Location, input.PropertyType, input.Vacancy, input.TotalRent).
			Scan(&propertyID, &createdAt, &updatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create property"})
			return
		}

		// 4. Respond
		c.JSON(http.StatusCreated, gin.H{
			"message": "Property created successfully",
			"data": gin.H{
				"id":            propertyID,
				"landlord_id":   landlordID,
				"title":         input.Title,
				"description":   input.Description,
				"location":      input.Location,
				"property_type": input.PropertyType,
				"vacancy":       input.Vacancy,
				"total_rent":    input.TotalRent,
				"created_at":    createdAt,
				"updated_at":    updatedAt,
			},
		})
	}
}

func ListProperties(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		query := `
			SELECT id, title, description, location, property_type, vacancy, total_rent, created_at, updated_at
			FROM properties
			WHERE landlord_id = $1
			ORDER BY created_at DESC
		`

		rows, err := db.Query(query, landlordID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch properties"})
			return
		}
		defer rows.Close()

		properties := []gin.H{}
		for rows.Next() {
			var p struct {
				ID           int
				Title        string
				Description  string
				Location     string
				PropertyType string
				Vacancy      bool
				TotalRent    float64
				CreatedAt    string
				UpdatedAt    string
			}
			if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Location, &p.PropertyType, &p.Vacancy, &p.TotalRent, &p.CreatedAt, &p.UpdatedAt); err != nil {
				continue
			}
			properties = append(properties, gin.H{
				"id":            p.ID,
				"title":         p.Title,
				"description":   p.Description,
				"location":      p.Location,
				"property_type": p.PropertyType,
				"vacancy":       p.Vacancy,
				"total_rent":    p.TotalRent,
				"created_at":    p.CreatedAt,
				"updated_at":    p.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"data": properties})
	}
}

func GetProperty(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		propertyID := c.Param("propertyId")

		query := `
			SELECT id, title, description, location, property_type, vacancy, total_rent, created_at, updated_at
			FROM properties
			WHERE id = $1 AND landlord_id = $2
		`

		var p struct {
			ID           int
			Title        string
			Description  string
			Location     string
			PropertyType string
			Vacancy      bool
			TotalRent    float64
			CreatedAt    string
			UpdatedAt    string
		}

		err = db.QueryRow(query, propertyID, landlordID).Scan(&p.ID, &p.Title, &p.Description, &p.Location, &p.PropertyType, &p.Vacancy, &p.TotalRent, &p.CreatedAt, &p.UpdatedAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Property not found or unauthorized"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch property"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": gin.H{
			"id":            p.ID,
			"title":         p.Title,
			"description":   p.Description,
			"location":      p.Location,
			"property_type": p.PropertyType,
			"vacancy":       p.Vacancy,
			"total_rent":    p.TotalRent,
			"created_at":    p.CreatedAt,
			"updated_at":    p.UpdatedAt,
		}})
	}
}

func UpdateProperty(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		propertyID := c.Param("propertyId")

		var input UpdatePropertyInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify ownership first
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM properties WHERE id = $1 AND landlord_id = $2)", propertyID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Property not found or unauthorized"})
			return
		}

		// Build dynamic update query
		query := "UPDATE properties SET updated_at = NOW()"
		args := []interface{}{}
		argID := 1

		if input.Title != nil {
			query += ", title = $" + strconv.Itoa(argID)
			args = append(args, *input.Title)
			argID++
		}
		if input.Description != nil {
			query += ", description = $" + strconv.Itoa(argID)
			args = append(args, *input.Description)
			argID++
		}
		if input.Location != nil {
			query += ", location = $" + strconv.Itoa(argID)
			args = append(args, *input.Location)
			argID++
		}
		if input.PropertyType != nil {
			query += ", property_type = $" + strconv.Itoa(argID)
			args = append(args, *input.PropertyType)
			argID++
		}
		if input.Vacancy != nil {
			query += ", vacancy = $" + strconv.Itoa(argID)
			args = append(args, *input.Vacancy)
			argID++
		}
		if input.TotalRent != nil {
			query += ", total_rent = $" + strconv.Itoa(argID)
			args = append(args, *input.TotalRent)
			argID++
		}

		query += " WHERE id = $" + strconv.Itoa(argID) + " AND landlord_id = $" + strconv.Itoa(argID+1)
		args = append(args, propertyID, landlordID)

		_, err = db.Exec(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update property"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Property updated successfully"})
	}
}
func DeleteProperty(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		landlordID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		propertyID := c.Param("propertyId")

		// Verify ownership
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM properties WHERE id = $1 AND landlord_id = $2)", propertyID, landlordID).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Property not found or unauthorized"})
			return
		}

		// Hard delete (cascade will handle units and tenants if configured, but let's be careful.
		// For now, simple delete. If units/tenants exist, this will fail if FK constraints are strict).
		_, err = db.Exec("DELETE FROM properties WHERE id = $1 AND landlord_id = $2", propertyID, landlordID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete property. Ensure all units are vacated."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Property deleted successfully"})
	}
}
