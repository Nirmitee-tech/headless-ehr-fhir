package db

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
)

type contextKey string

const (
	TenantIDKey contextKey = "tenant_id"
	DBConnKey   contextKey = "db_conn"
	DBTxKey     contextKey = "db_tx"
)

var tenantIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func TenantMiddleware(pool *pgxpool.Pool, defaultTenant string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip tenant resolution for public infrastructure paths that
			// do not need a database connection (health, metrics, etc.).
			if auth.IsPublicPath(c.Path()) {
				return next(c)
			}

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

			// Set session variables for RLS policies (migration 036).
			// These allow PostgreSQL RLS functions to verify that the
			// search_path and tenant context are consistent, and to
			// attribute writes to the authenticated user.
			if _, err = conn.Exec(ctx, "SET app.current_tenant_id = "+quoteLiteral(tenantID)); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "tenant context failed")
			}

			userID := auth.UserIDFromContext(ctx)
			if userID != "" {
				if _, err = conn.Exec(ctx, "SET app.current_user_id = "+quoteLiteral(userID)); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "user context failed")
				}
			}

			roles := auth.RolesFromContext(ctx)
			if len(roles) > 0 {
				if _, err = conn.Exec(ctx, "SET app.current_user_roles = "+quoteLiteral(strings.Join(roles, ","))); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "role context failed")
				}
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

// WithTx starts a transaction using the connection from context and returns a new context
// containing the transaction. The caller must commit or rollback the returned pgx.Tx.
func WithTx(ctx context.Context) (context.Context, pgx.Tx, error) {
	conn := ConnFromContext(ctx)
	if conn == nil {
		return ctx, nil, fmt.Errorf("no database connection in context")
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return ctx, nil, fmt.Errorf("begin transaction: %w", err)
	}
	txCtx := context.WithValue(ctx, DBTxKey, tx)
	return txCtx, tx, nil
}

// TxFromContext retrieves the active transaction from context, if any.
func TxFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(DBTxKey).(pgx.Tx)
	return tx
}

// quoteLiteral returns a PostgreSQL-safe quoted string literal.
// It wraps the value in single quotes and escapes any embedded single quotes,
// preventing SQL injection in SET commands where parameterized queries
// are not supported for session variable values.
func quoteLiteral(s string) string {
	escaped := strings.ReplaceAll(s, "'", "''")
	return "'" + escaped + "'"
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
