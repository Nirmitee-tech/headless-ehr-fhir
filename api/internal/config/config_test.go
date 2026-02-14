package config

import (
	"os"
	"testing"
)

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoad_WithDatabaseURL(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	defer os.Unsetenv("DATABASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DatabaseURL != "postgres://test:test@localhost:5432/test" {
		t.Errorf("expected DATABASE_URL to be set, got %s", cfg.DatabaseURL)
	}

	if cfg.Port != "8000" {
		t.Errorf("expected default port 8000, got %s", cfg.Port)
	}

	if cfg.DefaultTenant != "default" {
		t.Errorf("expected default tenant 'default', got %s", cfg.DefaultTenant)
	}

	if cfg.DBMaxConns != 20 {
		t.Errorf("expected default max conns 20, got %d", cfg.DBMaxConns)
	}
}

func TestConfig_IsDev(t *testing.T) {
	c := &Config{Env: "development"}
	if !c.IsDev() {
		t.Error("expected IsDev() to return true for development")
	}

	c.Env = "production"
	if c.IsDev() {
		t.Error("expected IsDev() to return false for production")
	}
}
