package db

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type contextKey string

const (
	TenantIDKey contextKey = "tenant_id"
	DBConnKey   contextKey = "db_conn"
)

var tenantIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func TenantMiddleware(pool *pgxpool.Pool, defaultTenant string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID := extractTenantID(c, defaultTenant)

			if !tenantIDPattern.MatchString(tenantID) {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant identifier")
			}

			ctx := c.Request().Context()
			conn, err := pool.Acquire(ctx)
			if err != nil {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "database unavailable")
			}
			defer conn.Release()

			schema := fmt.Sprintf("tenant_%s", tenantID)
			_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s, shared, public", schema))
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "tenant resolution failed")
			}

			ctx = context.WithValue(ctx, TenantIDKey, tenantID)
			ctx = context.WithValue(ctx, DBConnKey, conn)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Set("tenant_id", tenantID)
			c.Set("db", conn)

			return next(c)
		}
	}
}

func extractTenantID(c echo.Context, defaultTenant string) string {
	// 1. Check JWT claim (set by auth middleware)
	if tid, ok := c.Get("jwt_tenant_id").(string); ok && tid != "" {
		return tid
	}

	// 2. Check X-Tenant-ID header
	if tid := c.Request().Header.Get("X-Tenant-ID"); tid != "" {
		return tid
	}

	// 3. Check query parameter
	if tid := c.QueryParam("tenant_id"); tid != "" {
		return tid
	}

	return defaultTenant
}

// ConnFromContext retrieves the tenant-scoped database connection from context.
func ConnFromContext(ctx context.Context) *pgxpool.Conn {
	conn, _ := ctx.Value(DBConnKey).(*pgxpool.Conn)
	return conn
}

// TenantFromContext retrieves the tenant ID from context.
func TenantFromContext(ctx context.Context) string {
	tid, _ := ctx.Value(TenantIDKey).(string)
	return tid
}

// CreateTenantSchema creates a new schema for a tenant and runs all migrations against it.
// The migrationsDir parameter specifies the directory containing migration SQL files.
// If migrationsDir is empty, migrations are skipped.
func CreateTenantSchema(ctx context.Context, pool *pgxpool.Pool, tenantID string, migrationsDir string) error {
	if !tenantIDPattern.MatchString(tenantID) {
		return fmt.Errorf("invalid tenant identifier: %s", tenantID)
	}

	schema := fmt.Sprintf("tenant_%s", tenantID)

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema))
	if err != nil {
		return fmt.Errorf("create schema %s: %w", schema, err)
	}

	if migrationsDir != "" {
		migrator := NewMigrator(pool, migrationsDir)
		if _, err := migrator.Up(ctx, schema); err != nil {
			return fmt.Errorf("run migrations for %s: %w", schema, err)
		}
	}

	return nil
}
