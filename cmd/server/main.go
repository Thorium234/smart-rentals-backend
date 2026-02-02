package main

import (
	"log"

	"github.com/Zolet-hash/smart-rentals/internal/api"
	"github.com/Zolet-hash/smart-rentals/internal/config"
	"github.com/Zolet-hash/smart-rentals/internal/database"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.DB.Close()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Register routes
	api.SetupRoutes(r, db, cfg)

	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
