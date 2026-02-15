package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrationSmartLaunchContexts is the SQL DDL for the smart_launch_contexts
// table. It is safe to execute multiple times (uses IF NOT EXISTS). Callers
// can run this at application startup as an auto-migration step.
const MigrationSmartLaunchContexts = `
CREATE TABLE IF NOT EXISTS smart_launch_contexts (
    id          TEXT PRIMARY KEY,
    context_json JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_smart_launch_contexts_expires_at
    ON smart_launch_contexts (expires_at);
`

// ---------------------------------------------------------------------------
// pgRow / pgConn abstractions (allow unit testing without a real DB)
// ---------------------------------------------------------------------------

// pgRow represents a single row returned by QueryRow.
type pgRow interface {
	Scan(dest ...any) error
}

// pgConn is the minimal database interface required by PGLaunchContextStore.
// Both *pgxpool.Pool (via a thin adapter) and test mocks implement this.
type pgConn interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgRow
	Exec(ctx context.Context, sql string, args ...any) error
}

// ---------------------------------------------------------------------------
// PGLaunchContextStore
// ---------------------------------------------------------------------------

// PGLaunchContextStore is a PostgreSQL-backed implementation of
// LaunchContextStorer. Launch contexts are stored in the
// smart_launch_contexts table as JSONB with an explicit expires_at column
// that the database uses for filtering.
type PGLaunchContextStore struct {
	db  pgConn
	ttl time.Duration
}

// NewPGLaunchContextStore creates a PG-backed store. The db parameter must
// satisfy the pgConn interface -- use NewPGLaunchContextStoreFromPool to wrap
// a *pgxpool.Pool, or pass a mock in tests.
func NewPGLaunchContextStore(db pgConn, ttl time.Duration) *PGLaunchContextStore {
	return &PGLaunchContextStore{db: db, ttl: ttl}
}

// Save implements LaunchContextStorer. It inserts or replaces (upsert) the
// launch context in the database.
func (s *PGLaunchContextStore) Save(ctx context.Context, id string, lc *LaunchContext) error {
	data, err := json.Marshal(launchContextToJSON(lc))
	if err != nil {
		return fmt.Errorf("marshal launch context: %w", err)
	}

	expiresAt := lc.CreatedAt.Add(s.ttl)

	const query = `INSERT INTO smart_launch_contexts (id, context_json, created_at, expires_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET context_json = EXCLUDED.context_json,
                                created_at  = EXCLUDED.created_at,
                                expires_at  = EXCLUDED.expires_at`

	if err := s.db.Exec(ctx, query, id, data, lc.CreatedAt, expiresAt); err != nil {
		return fmt.Errorf("save launch context: %w", err)
	}
	return nil
}

// Get implements LaunchContextStorer. It selects the row only if it has not
// expired.
func (s *PGLaunchContextStore) Get(ctx context.Context, id string) (*LaunchContext, error) {
	const query = `SELECT context_json FROM smart_launch_contexts
WHERE id = $1 AND expires_at > now()`

	var data []byte
	if err := s.db.QueryRow(ctx, query, id).Scan(&data); err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get launch context: %w", err)
	}

	var j launchContextJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, fmt.Errorf("unmarshal launch context: %w", err)
	}
	return launchContextFromJSON(j), nil
}

// Consume implements LaunchContextStorer. It atomically deletes and returns
// the row using DELETE ... RETURNING.
func (s *PGLaunchContextStore) Consume(ctx context.Context, id string) (*LaunchContext, error) {
	const query = `DELETE FROM smart_launch_contexts
WHERE id = $1 AND expires_at > now()
RETURNING context_json`

	var data []byte
	if err := s.db.QueryRow(ctx, query, id).Scan(&data); err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("consume launch context: %w", err)
	}

	var j launchContextJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, fmt.Errorf("unmarshal launch context: %w", err)
	}
	return launchContextFromJSON(j), nil
}

// Cleanup deletes all expired rows from the table.
func (s *PGLaunchContextStore) Cleanup(ctx context.Context) error {
	const query = `DELETE FROM smart_launch_contexts WHERE expires_at <= now()`
	if err := s.db.Exec(ctx, query); err != nil {
		return fmt.Errorf("cleanup launch contexts: %w", err)
	}
	return nil
}

// isNoRows returns true when the error represents a "no rows" condition.
// It works with both pgx (pgx.ErrNoRows) and the mock used in tests.
func isNoRows(err error) bool {
	if err == pgx.ErrNoRows {
		return true
	}
	return err != nil && strings.Contains(err.Error(), "no rows")
}

// ---------------------------------------------------------------------------
// pgxPoolWrapper adapts *pgxpool.Pool to the pgConn interface
// ---------------------------------------------------------------------------

// pgxPoolWrapper wraps a *pgxpool.Pool so it satisfies the pgConn interface.
// The adapter is necessary because pgxpool.Pool.Exec returns
// (pgconn.CommandTag, error) whereas pgConn.Exec returns only error.
type pgxPoolWrapper struct {
	pool *pgxpool.Pool
}

func (w *pgxPoolWrapper) QueryRow(ctx context.Context, sql string, args ...any) pgRow {
	return w.pool.QueryRow(ctx, sql, args...)
}

func (w *pgxPoolWrapper) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := w.pool.Exec(ctx, sql, args...)
	return err
}

// NewPGLaunchContextStoreFromPool creates a PG-backed store directly from a
// *pgxpool.Pool. This is the recommended constructor for production use.
func NewPGLaunchContextStoreFromPool(pool *pgxpool.Pool, ttl time.Duration) *PGLaunchContextStore {
	return &PGLaunchContextStore{
		db:  &pgxPoolWrapper{pool: pool},
		ttl: ttl,
	}
}
