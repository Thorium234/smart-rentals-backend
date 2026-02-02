package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	ssl := os.Getenv("DB_SSLMODE")
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "zolet"
	}
	if ssl == "" {
		ssl = "disable"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, name, ssl)
	fmt.Println("DSN:", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println("open err:", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Println("ping err:", err)
		return
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", "debug@example.com").Scan(&exists)
	if err != nil {
		fmt.Println("select exists err:", err)
	} else {
		fmt.Println("select exists ok, exists=", exists)
	}

	// Try insert to see permission or constraint errors
	tx, err := db.Begin()
	if err != nil {
		fmt.Println("begin err:", err)
		return
	}
	_, err = tx.Exec("INSERT INTO users (email, password_hash, full_name, phone, role) VALUES ($1,$2,$3,$4,$5)", "dbg@example.com", "hash", "dbg", "123", "tenant")
	if err != nil {
		fmt.Println("insert err:", err)
		tx.Rollback()
		return
	}
	if err := tx.Commit(); err != nil {
		fmt.Println("commit err:", err)
		return
	}
	fmt.Println("insert succeeded")
}
