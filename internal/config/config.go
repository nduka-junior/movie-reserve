package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)
var Cfg *Config

type Config struct {
	Server struct {
		Port         string
		Host         string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}

	// Remove the split Database fields — we only need the full URL for Neon
	JWT struct {
		Secret        string
		TokenExpiry   time.Duration
		RefreshExpiry time.Duration
	}

	Environment string
	DatabaseURL string // ← new field
    Cloudinary struct {
        CloudName string
        APIKey    string
        APISecret string
    }
}

func Load() (*Config, error) {
    // Load .env from current directory (project root)
    if err := godotenv.Load(".env"); err != nil {
        fmt.Printf("Warning: could not load .env file: %v (falling back to system env vars)\n", err)
    }

    cfg := &Config{}
Cfg = cfg
    // Server config
    cfg.Server.Port = os.Getenv("SERVER_PORT")
    if cfg.Server.Port == "" {
        cfg.Server.Port = "8080"
    }
    cfg.Server.Host = os.Getenv("SERVER_HOST")
    if cfg.Server.Host == "" {
        cfg.Server.Host = "0.0.0.0"
    }
    cfg.Server.ReadTimeout = 15 * time.Second
    cfg.Server.WriteTimeout = 15 * time.Second

    // Database — load from env only (no hard-coding!)
    cfg.DatabaseURL = os.Getenv("DATABASE_URL")
    if cfg.DatabaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL is required in .env file or environment variables")
    }

    // Debug: show what was actually loaded
    fmt.Printf("Loaded DATABASE_URL: %s\n", cfg.DatabaseURL)

    // JWT config
    cfg.JWT.Secret = os.Getenv("JWT_SECRET")
    if cfg.JWT.Secret == "" {
        cfg.JWT.Secret = "your-secret-key"
    }
    if len(cfg.JWT.Secret) < 32 {
        fmt.Println("WARNING: JWT_SECRET is short — use at least 32 characters in production")
    }
    cfg.JWT.TokenExpiry   = 24 * time.Hour
    cfg.JWT.RefreshExpiry = 168 * time.Hour

    cfg.Environment = os.Getenv("ENV")
    if cfg.Environment == "" {
        cfg.Environment = "development"
    }
       

    return cfg, nil
}


func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {

		return value
	}
	return defaultValue
}



