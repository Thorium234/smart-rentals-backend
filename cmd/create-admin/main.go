package main

import (
	"fmt"
	"log"

	"github.com/Zolet-hash/smart-rentals/internal/config"
	"github.com/Zolet-hash/smart-rentals/internal/database"
	"github.com/Zolet-hash/smart-rentals/internal/utils"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := database.NewDatabase(cfg.GetDSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.DB.Close()

	// Admin credentials
	email := "landlord@test.com"
	password := "password123"
	fullName := "Test Landlord"
	phone := "0711223344"
	role := "landlord"

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Insert user
	query := `
		INSERT INTO users (email, password_hash, full_name, phone, role)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email) DO UPDATE 
		SET password_hash = EXCLUDED.password_hash,
		    full_name = EXCLUDED.full_name,
		    phone = EXCLUDED.phone,
		    role = EXCLUDED.role
		RETURNING id
	`

	var userID int
	err = db.DB.QueryRow(query, email, hashedPassword, fullName, phone, role).Scan(&userID)
	if err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	fmt.Printf("✅ Admin user created/updated successfully!\n")
	fmt.Printf("   ID: %d\n", userID)
	fmt.Printf("   Email: %s\n", email)
	fmt.Printf("   Password: %s\n", password)
	fmt.Printf("   Role: %s\n", role)
	fmt.Println("\nYou can now login with these credentials!")
}
