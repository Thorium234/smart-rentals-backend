package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server struct {
		Port         string
		Host         string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}
	Database struct {
		URL string // PRIMARY: Full DATABASE_URL for production
	}
	JWT struct {
		Secret        string
		TokenExpiry   time.Duration
		RefreshExpiry time.Duration
	}
	CORS struct {
		AllowedOrigins []string
		AllowedMethods []string
		AllowedHeaders []string
	}
	Environment          string
	MpesaEnvironment     string
	MpesaCallbackBaseURL string
	LogLevel             string
}

func Load() (*Config, error) {
	// Load .env if exists (for local development only)
	// In production (Render), environment variables are set directly
	godotenv.Load()

	cfg := &Config{}

	// Server config
	cfg.Server.Port = getEnv("APP_PORT", "8080")
	cfg.Server.Host = getEnv("SERVER_HOST", "0.0.0.0")
	cfg.Server.ReadTimeout = 15 * time.Second
	cfg.Server.WriteTimeout = 15 * time.Second

	// Database config - DATABASE_URL is PRIMARY
	cfg.Database.URL = os.Getenv("DATABASE_URL")

	// JWT config - NO DEFAULTS for security-critical values
	cfg.JWT.Secret = os.Getenv("JWT_SECRET")

	// Parse JWT expiry or use default
	expiryStr := getEnv("JWT_EXPIRES_IN", "24h")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiry = 24 * time.Hour
	}
	cfg.JWT.TokenExpiry = expiry
	cfg.JWT.RefreshExpiry = time.Hour * 168 // 7 days

	// CORS config - REQUIRED for production
	originsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	if originsStr != "" {
		cfg.CORS.AllowedOrigins = strings.Split(originsStr, ",")
		// Trim whitespace from each origin
		for i := range cfg.CORS.AllowedOrigins {
			cfg.CORS.AllowedOrigins[i] = strings.TrimSpace(cfg.CORS.AllowedOrigins[i])
		}
	}

	methodsStr := getEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	cfg.CORS.AllowedMethods = strings.Split(methodsStr, ",")

	headersStr := getEnv("CORS_ALLOWED_HEADERS", "Authorization,Content-Type")
	cfg.CORS.AllowedHeaders = strings.Split(headersStr, ",")

	// Environment
	cfg.Environment = getEnv("APP_ENV", "development")
	cfg.MpesaEnvironment = getEnv("MPESA_ENV", "sandbox")
	cfg.MpesaCallbackBaseURL = os.Getenv("MPESA_CALLBACK_BASE_URL")

	// Logging
	cfg.LogLevel = getEnv("LOG_LEVEL", "info")

	return cfg, nil
}

// Validate checks that all required configuration is present
// This should be called immediately after Load() to fail fast
func (c *Config) Validate() error {
	// Database validation
	if c.Database.URL == "" {
		return errors.New("DATABASE_URL is required - must be set in environment variables")
	}

	// JWT validation
	if c.JWT.Secret == "" {
		return errors.New("JWT_SECRET is required - generate with: openssl rand -base64 64")
	}
	if len(c.JWT.Secret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters for security")
	}

	// CORS validation - required in production
	if c.Environment == "production" {
		if len(c.CORS.AllowedOrigins) == 0 {
			return errors.New("CORS_ALLOWED_ORIGINS is required in production - no wildcards allowed")
		}
		// Check for wildcard in production
		for _, origin := range c.CORS.AllowedOrigins {
			if origin == "*" {
				return errors.New("CORS wildcard (*) is not allowed in production - specify explicit origins")
			}
		}
	}

	// M-Pesa callback URL validation for production
	if c.Environment == "production" && c.MpesaCallbackBaseURL == "" {
		return errors.New("MPESA_CALLBACK_BASE_URL is required in production")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDSN returns the database connection string
// In production, this is simply the DATABASE_URL
func (c *Config) GetDSN() string {
	return c.Database.URL
}
