package main

import (
	"log"
	"net/http"

	"github.com/Zolet-hash/smart-rentals/internal/api"
	"github.com/Zolet-hash/smart-rentals/internal/config"
	"github.com/Zolet-hash/smart-rentals/internal/database"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("Starting Smart Rentals API...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("CRITICAL: Failed to load config: %v", err)
	}

	log.Printf("Configuration loaded (ENV: %s)", cfg.Environment)

	// Validate configuration - fail fast if required vars are missing
	if err := cfg.Validate(); err != nil {
		log.Fatalf("CRITICAL: Configuration validation failed: %v", err)
	}

	log.Println("Configuration validated, connecting to database...")

	// Initialize database
	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		log.Fatalf("CRITICAL: Failed to connect to database: %v", err)
	}
	defer db.DB.Close()

	log.Println("Database connected successfully")
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router with middleware
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Health check route for production verification
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "smart-rentals-backend",
			"env":     cfg.Environment,
		})
	})

	// CORS is handled by middleware.CORS(cfg) in routes.go
	// No hardcoded CORS here - all CORS configuration is in config

	// Initialize routes using the dedicated routes package
	api.SetupRoutes(r, db, cfg)

	// Start server with configured host and port
	// In production, binding to ":PORT" is most reliable (equivalent to 0.0.0.0:PORT)
	serverAddr := ":" + cfg.Server.Port
	log.Printf("Server preparing to listen on %s (environment: %s)", serverAddr, cfg.Environment)

	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	log.Printf("🚀 Starting HTTP server on %s...", serverAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("CRITICAL: Server failed to start on %s: %v", serverAddr, err)
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
