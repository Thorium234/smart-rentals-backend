package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// 1. Load configuration
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// 2. Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to open database connection:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	fmt.Println("Successfully connected to the database")

	// 3. Create migrations table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal("Failed to create migrations table:", err)
	}

	// 4. Find all migration files
	migrationDir := "./migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		log.Fatal("Failed to read migrations directory:", err)
	}

	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}
	sort.Strings(migrationFiles)

	// 5. Apply migrations
	for _, filename := range migrationFiles {
		// Extract version from filename (assuming timestamp prefix)
		var version int64
		_, err := fmt.Sscanf(filename, "%d", &version)
		if err != nil {
			log.Printf("Skipping file %s: could not parse version from filename\n", filename)
			continue
		}

		// Check if migration already applied
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			log.Fatalf("Failed to check migration status for %s: %v\n", filename, err)
		}

		if exists {
			fmt.Printf("Migration %s already applied, skipping\n", filename)
			continue
		}

		fmt.Printf("Applying migration: %s\n", filename)

		content, err := os.ReadFile(filepath.Join(migrationDir, filename))
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v\n", filename, err)
		}

		// Use a transaction for each migration
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to start transaction for %s: %v\n", filename, err)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to execute migration %s: %v\n", filename, err)
		}

		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to record migration %s: %v\n", filename, err)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatalf("Failed to commit migration %s: %v\n", filename, err)
		}

		fmt.Printf("Successfully applied %s\n", filename)
	}

	fmt.Println("All migrations completed successfully!")
}
