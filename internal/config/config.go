package config

import (
	"fmt"
	"os"
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
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	JWT struct {
		Secret        string
		TokenExpiry   time.Duration
		RefreshExpiry time.Duration
	}
	Environment      string
	MpesaEnvironment string
}

func Load() (*Config, error) {
	godotenv.Load() //Load .env if exists

	cfg := &Config{}

	//Server config
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")
	cfg.Server.Host = getEnv("SERVER_HOST", "0.0.0.0")

	//Database config
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.User = getEnv("DB_USER", "zolet")
	cfg.Database.Password = getEnv("DB_PASSWORD", "test")
	cfg.Database.DBName = getEnv("DB_NAME", "auth_service")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")

	//JWT config
	cfg.JWT.Secret = getEnv("JWT_SECRET", "")
	cfg.JWT.TokenExpiry = time.Hour * 24    //24 hours
	cfg.JWT.RefreshExpiry = time.Hour * 168 //7 days

	cfg.Environment = getEnv("ENV", "development")
	cfg.MpesaEnvironment = getEnv("MPESA_ENVIRONMENT", "sandbox")
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}
