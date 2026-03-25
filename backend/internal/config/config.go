package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port            int
	Env             string
	DB              DBConfig
	CORSAllowedURLs []string
	SessionSecret   string
	SessionMaxAge   time.Duration
	AzureAD         AzureADConfig
	FrontendURL     string
}

// AzureADConfig holds Azure AD / Entra ID OIDC settings.
type AzureADConfig struct {
	TenantID     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// DBConfig holds database connection parameters.
type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// DSN returns the PostgreSQL connection string.
func (db DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode,
	)
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")

	cfg := &Config{
		Port: getEnvInt("APP_PORT", 8080),
		Env:  getEnv("APP_ENV", "development"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			Port:     getEnvInt("DB_PORT", 5432),
			Name:     getEnv("DB_NAME", "zeiterfassung"),
			User:     getEnv("DB_USER", "zeit_app"),
			Password: os.Getenv("DB_PASSWORD"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		CORSAllowedURLs: getEnvList("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		SessionSecret:   os.Getenv("SESSION_SECRET"),
		SessionMaxAge:   time.Duration(getEnvInt("SESSION_MAX_AGE", 86400)) * time.Second,
		AzureAD: AzureADConfig{
			TenantID:     os.Getenv("AZURE_TENANT_ID"),
			ClientID:     os.Getenv("AZURE_CLIENT_ID"),
			ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
			RedirectURL:  getEnv("AZURE_REDIRECT_URL", frontendURL+"/api/auth/callback"),
		},
		FrontendURL: frontendURL,
	}

	if cfg.DB.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
	}
	if cfg.SessionSecret == "" && cfg.Env == "production" {
		return nil, fmt.Errorf("SESSION_SECRET environment variable is required in production")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return n
}
