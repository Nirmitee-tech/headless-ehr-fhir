package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port                string   `mapstructure:"PORT"`
	Env                 string   `mapstructure:"ENV"`
	DatabaseURL         string   `mapstructure:"DATABASE_URL"`
	DBMaxConns          int32    `mapstructure:"DB_MAX_CONNS"`
	DBMinConns          int32    `mapstructure:"DB_MIN_CONNS"`
	RedisURL            string   `mapstructure:"REDIS_URL"`
	AuthIssuer          string   `mapstructure:"AUTH_ISSUER"`
	AuthJWKSURL         string   `mapstructure:"AUTH_JWKS_URL"`
	AuthAudience        string   `mapstructure:"AUTH_AUDIENCE"`
	DefaultTenant       string   `mapstructure:"DEFAULT_TENANT"`
	CORSOrigins         []string `mapstructure:"CORS_ORIGINS"`
	HIPAAEncryptionKey  string   `mapstructure:"HIPAA_ENCRYPTION_KEY"`
	RateLimitRPS        float64  `mapstructure:"RATE_LIMIT_RPS"`
	RateLimitBurst      int      `mapstructure:"RATE_LIMIT_BURST"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("PORT", "8000")
	v.SetDefault("ENV", "development")
	v.SetDefault("DB_MAX_CONNS", 20)
	v.SetDefault("DB_MIN_CONNS", 5)
	v.SetDefault("DEFAULT_TENANT", "default")
	v.SetDefault("CORS_ORIGINS", "http://localhost:3000")
	v.SetDefault("RATE_LIMIT_RPS", 100)
	v.SetDefault("RATE_LIMIT_BURST", 200)

	// Bind env vars explicitly so Unmarshal picks them up
	v.BindEnv("PORT")
	v.BindEnv("ENV")
	v.BindEnv("DATABASE_URL")
	v.BindEnv("DB_MAX_CONNS")
	v.BindEnv("DB_MIN_CONNS")
	v.BindEnv("REDIS_URL")
	v.BindEnv("AUTH_ISSUER")
	v.BindEnv("AUTH_JWKS_URL")
	v.BindEnv("AUTH_AUDIENCE")
	v.BindEnv("DEFAULT_TENANT")
	v.BindEnv("CORS_ORIGINS")
	v.BindEnv("HIPAA_ENCRYPTION_KEY")
	v.BindEnv("RATE_LIMIT_RPS")
	v.BindEnv("RATE_LIMIT_BURST")

	// Try reading .env file, but don't fail if missing
	_ = v.ReadInConfig()

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.CORSOrigins == nil {
		origins := v.GetString("CORS_ORIGINS")
		if origins != "" {
			cfg.CORSOrigins = strings.Split(origins, ",")
		}
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func (c *Config) IsDev() bool {
	return c.Env == "development"
}
