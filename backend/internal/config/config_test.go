package config

import (
	"os"
	"testing"
)

func TestLoad_MissingDBPassword_ReturnsError(t *testing.T) {
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("SESSION_SECRET")
	os.Setenv("APP_ENV", "development")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DB_PASSWORD is missing, got nil")
	}
}

func TestLoad_ValidConfig_ReturnsConfig(t *testing.T) {
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("APP_ENV", "development")
	defer os.Unsetenv("DB_PASSWORD")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.DB.Name != "zeiterfassung" {
		t.Errorf("expected db name zeiterfassung, got %s", cfg.DB.Name)
	}
	if cfg.DB.Password != "testpass" {
		t.Errorf("expected db password testpass, got %s", cfg.DB.Password)
	}
}

func TestLoad_ProductionRequiresSessionSecret(t *testing.T) {
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("APP_ENV", "production")
	os.Unsetenv("SESSION_SECRET")
	defer os.Unsetenv("DB_PASSWORD")
	defer os.Unsetenv("APP_ENV")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when SESSION_SECRET is missing in production")
	}
}

func TestDSN_Format(t *testing.T) {
	db := DBConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "testdb",
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "disable",
	}

	expected := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
	if dsn := db.DSN(); dsn != expected {
		t.Errorf("expected DSN %q, got %q", expected, dsn)
	}
}

func TestGetEnvInt_Fallback(t *testing.T) {
	os.Unsetenv("TEST_INT_VAR")
	if val := getEnvInt("TEST_INT_VAR", 42); val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
}

func TestGetEnvInt_InvalidValue_ReturnsFallback(t *testing.T) {
	os.Setenv("TEST_INT_VAR", "not-a-number")
	defer os.Unsetenv("TEST_INT_VAR")

	if val := getEnvInt("TEST_INT_VAR", 42); val != 42 {
		t.Errorf("expected 42 for invalid int, got %d", val)
	}
}
