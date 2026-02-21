package config

import (
	"encoding/hex"
	"os"
	"testing"
)

// validHIPAAKey is a 32-byte key encoded as 64 hex characters, used by tests
// that need a valid production configuration.
var validHIPAAKey = hex.EncodeToString(make([]byte, 32))

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

func TestConfig_IsProduction(t *testing.T) {
	c := &Config{Env: "production"}
	if !c.IsProduction() {
		t.Error("expected IsProduction() to return true for production")
	}

	c.Env = "development"
	if c.IsProduction() {
		t.Error("expected IsProduction() to return false for development")
	}

	c.Env = "staging"
	if c.IsProduction() {
		t.Error("expected IsProduction() to return false for staging")
	}
}

func TestLoad_DefaultIsDevelopment(t *testing.T) {
	// Ensure ENV is not set so the default takes effect.
	os.Unsetenv("ENV")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	defer os.Unsetenv("DATABASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("expected default ENV to be 'development', got %q", cfg.Env)
	}

	if !cfg.IsDev() {
		t.Error("expected IsDev() to return true with default ENV")
	}
}

func TestValidate_ProductionRequiresAuthIssuer(t *testing.T) {
	// Production without AUTH_ISSUER should fail validation.
	c := &Config{
		Env:        "production",
		AuthIssuer: "",
	}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected Validate() to return error when ENV=production and AUTH_ISSUER is empty")
	}
}

func TestValidate_ProductionWithAuthIssuer(t *testing.T) {
	c := &Config{
		Env:                "production",
		AuthIssuer:         "https://auth.example.com",
		HIPAAEncryptionKey: validHIPAAKey,
	}
	err := c.Validate()
	if err != nil {
		t.Fatalf("unexpected Validate() error: %v", err)
	}
}

func TestValidate_StagingWithoutAuthIssuerUsesStandalone(t *testing.T) {
	// ENV=staging without AUTH_ISSUER resolves to standalone mode (valid).
	c := &Config{
		Env:        "staging",
		AuthIssuer: "",
	}
	err := c.Validate()
	if err != nil {
		t.Fatalf("unexpected Validate() error: standalone mode should be valid: %v", err)
	}
	if c.ResolvedAuthMode() != "standalone" {
		t.Fatalf("expected standalone auth mode, got %q", c.ResolvedAuthMode())
	}
}

func TestValidate_ExternalModeRequiresAuthIssuer(t *testing.T) {
	// Explicit AUTH_MODE=external without AUTH_ISSUER should fail.
	c := &Config{
		Env:      "staging",
		AuthMode: "external",
	}
	err := c.Validate()
	if err == nil {
		t.Fatal("expected Validate() to return error when AUTH_MODE=external and AUTH_ISSUER is empty")
	}
}

func TestValidate_DevelopmentDoesNotRequireAuthIssuer(t *testing.T) {
	c := &Config{
		Env:        "development",
		AuthIssuer: "",
	}
	err := c.Validate()
	if err != nil {
		t.Fatalf("unexpected Validate() error in development: %v", err)
	}
}
