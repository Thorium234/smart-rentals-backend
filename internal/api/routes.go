package api

import (
	"github.com/Zolet-hash/smart-rentals/internal/api/handlers"
	"github.com/Zolet-hash/smart-rentals/internal/api/middleware"
	"github.com/Zolet-hash/smart-rentals/internal/config"
	"github.com/Zolet-hash/smart-rentals/internal/database"
	"github.com/Zolet-hash/smart-rentals/internal/services"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	db *database.Database,
	cfg *config.Config,
) {
	// Global middleware
	r.Use(middleware.CORS(cfg))
	r.Use(middleware.RequestID())

	authHandler := handlers.NewAuthHandler(db, []byte(cfg.JWT.Secret))
	paymentSvc := services.NewPaymentService(db, cfg)
	paymentHandler := handlers.NewPaymentHandler(paymentSvc)

	// API v1
	api := r.Group("/api/v1")

	// Public routes
	api.POST("/login",
		middleware.RateLimiter(), // <- limit requests
		authHandler.Login,
	)
	// api.GET("/mpesa/validation", handlers.MpesaValidation)
	// api.POST("/mpesa/confirmation", handlers.MpesaPaymentConfirmation(db))

	// M-Pesa Routes
	api.POST("/payments/c2b/validation", paymentHandler.C2BValidation)
	api.POST("/payments/c2b/confirmation", paymentHandler.C2BConfirmation)

	// Protected routes (require authentication)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware([]byte(cfg.JWT.Secret)))
	{
		protected.GET("/profile", getUserProfile)
		protected.POST("/refresh-token", authHandler.RefreshToken)
		protected.POST("/logout", authHandler.Logout)
	}

	// Admin routes
	admin := api.Group("/sudo")
	admin.Use(
		middleware.AuthMiddleware([]byte(cfg.JWT.Secret)), // sets user_id
		middleware.RequireRole(db, "admin"),               // enforces role
	)
	{
		admin.POST("/register", authHandler.Register)
		admin.GET("/users", authHandler.ListUsers)
		admin.PATCH("/users/:id", authHandler.UpdateUser)
		admin.DELETE("/users/:id", authHandler.DeleteUser)
		admin.PATCH("/users/:id/reset-password", authHandler.ResetPassword)
	}

	// Landlord routes
	landlord := api.Group("/")
	landlord.Use(
		middleware.AuthMiddleware([]byte(cfg.JWT.Secret)),
	)
	{
		// Properties
		landlord.POST("/properties", handlers.CreateProperty(db))
		landlord.GET("/properties", handlers.ListProperties(db))
		landlord.GET("/properties/:propertyId", handlers.GetProperty(db))
		landlord.PATCH("/properties/:propertyId", handlers.UpdateProperty(db))
		landlord.DELETE("/properties/:propertyId", handlers.DeleteProperty(db))

		// Units
		landlord.POST("/properties/:propertyId/units", handlers.CreateUnit(db))
		landlord.GET("/properties/:propertyId/units", handlers.GetUnitsByProperty(db))
		landlord.PATCH("/properties/:propertyId/units/:unitId", handlers.UpdateUnit(db))
		landlord.DELETE("/properties/:propertyId/units/:unitId", handlers.DeleteUnit(db))

		// Tenants
		landlord.GET("/tenants", handlers.ListAllTenants(db))
		landlord.POST("/units/:unitId/tenants", handlers.CreateTenant(db))
		landlord.GET("/units/:unitId/tenants", handlers.ListAllTenants(db))
		landlord.GET("/tenants/:tenantId", handlers.GetTenant(db))
		landlord.PUT("/tenants/:tenantId", handlers.UpdateTenant(db))
		landlord.DELETE("/tenants/:tenantId", handlers.RemoveTenant(db))

		// Payments
		landlord.GET("/payments", handlers.ListPayments(db))
		landlord.POST("/payments/cash", handlers.RecordCashPayment(db))
		landlord.PATCH("/payments/:id/assign", handlers.AssignPayment(db))
		landlord.GET("/tenants/:tenantId/history", handlers.GetTenantHistory(db))

		// Configuration
		landlord.POST("/config/mpesa", paymentHandler.UpdateConfig)
	}
}

func getUserProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")

	c.JSON(200, gin.H{
		"user_id": userID,
		"email":   email,
	})
}
