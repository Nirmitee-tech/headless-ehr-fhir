package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migration represents a single database migration loaded from a SQL file.
type Migration struct {
	Version   int
	Name      string
	SQL       string
	AppliedAt time.Time
}

// MigrationStatus represents the status of a migration (applied or pending).
type MigrationStatus struct {
	Version   int
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

// Migrator handles reading and applying SQL migration files against a PostgreSQL schema.
type Migrator struct {
	pool *pgxpool.Pool
	dir  string // path to migrations directory
}

// NewMigrator creates a new Migrator that reads migration files from migrationsDir
// and applies them using the provided connection pool.
func NewMigrator(pool *pgxpool.Pool, migrationsDir string) *Migrator {
	return &Migrator{
		pool: pool,
		dir:  migrationsDir,
	}
}

// EnsureMigrationsTable creates the _migrations tracking table in the given schema
// if it does not already exist.
func (m *Migrator) EnsureMigrationsTable(ctx context.Context, schema string) error {
	query := fmt.Sprintf(`SET search_path TO %s;
CREATE TABLE IF NOT EXISTS _migrations (
    version INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMPTZ DEFAULT NOW()
)`, schema)

	_, err := m.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("create _migrations table in %s: %w", schema, err)
	}
	return nil
}

// LoadMigrations reads all .sql files from the migrations directory, parses the
// version number from the filename prefix (e.g., "001_core.sql" -> version 1),
// and returns them sorted by version. Files that do not start with a numeric
// prefix are silently skipped.
func (m *Migrator) LoadMigrations() ([]Migration, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory %s: %w", m.dir, err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// Parse version number from filename prefix (e.g., "001_core.sql" -> 1)
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			// Skip files without a numeric prefix
			continue
		}

		content, err := os.ReadFile(filepath.Join(m.dir, name))
		if err != nil {
			return nil, fmt.Errorf("read migration file %s: %w", name, err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// AppliedVersions queries the _migrations table in the given schema and returns
// a map of version numbers that have already been applied.
func (m *Migrator) AppliedVersions(ctx context.Context, schema string) (map[int]bool, error) {
	query := fmt.Sprintf(`SELECT version FROM %s._migrations`, schema)
	rows, err := m.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query applied versions in %s: %w", schema, err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan migration version: %w", err)
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied versions: %w", err)
	}

	return applied, nil
}

// Up applies all pending migrations in version order against the given schema.
// Each migration runs in its own transaction. Returns the count of applied migrations.
func (m *Migrator) Up(ctx context.Context, schema string) (int, error) {
	return m.UpTo(ctx, schema, 0)
}

// UpTo applies pending migrations up to (and including) targetVersion against the
// given schema. If targetVersion is 0, all pending migrations are applied.
// Each migration runs in its own transaction. Returns the count of applied migrations.
func (m *Migrator) UpTo(ctx context.Context, schema string, targetVersion int) (int, error) {
	if err := m.EnsureMigrationsTable(ctx, schema); err != nil {
		return 0, err
	}

	migrations, err := m.LoadMigrations()
	if err != nil {
		return 0, err
	}

	applied, err := m.AppliedVersions(ctx, schema)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, mig := range migrations {
		if targetVersion > 0 && mig.Version > targetVersion {
			break
		}
		if applied[mig.Version] {
			continue
		}

		if err := m.applyMigration(ctx, schema, mig); err != nil {
			return count, fmt.Errorf("apply migration %d (%s): %w", mig.Version, mig.Name, err)
		}
		count++
	}

	return count, nil
}

// applyMigration runs a single migration in a transaction and records it in
// the _migrations table.
func (m *Migrator) applyMigration(ctx context.Context, schema string, mig Migration) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set search_path for the transaction
	if _, err := tx.Exec(ctx, fmt.Sprintf("SET search_path TO %s, shared, public", schema)); err != nil {
		return fmt.Errorf("set search_path: %w", err)
	}

	// Execute migration SQL
	if _, err := tx.Exec(ctx, mig.SQL); err != nil {
		return fmt.Errorf("execute SQL: %w", err)
	}

	// Record the migration
	if _, err := tx.Exec(ctx,
		"INSERT INTO _migrations (version, name) VALUES ($1, $2)",
		mig.Version, mig.Name,
	); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit(ctx)
}

// Status returns the status of all known migrations (both applied and pending)
// for the given schema.
func (m *Migrator) Status(ctx context.Context, schema string) ([]MigrationStatus, error) {
	if err := m.EnsureMigrationsTable(ctx, schema); err != nil {
		return nil, err
	}

	migrations, err := m.LoadMigrations()
	if err != nil {
		return nil, err
	}

	// Query all applied migrations with their timestamps
	query := fmt.Sprintf(`SELECT version, applied_at FROM %s._migrations`, schema)
	rows, err := m.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query migration status in %s: %w", schema, err)
	}
	defer rows.Close()

	appliedMap := make(map[int]time.Time)
	for rows.Next() {
		var v int
		var at time.Time
		if err := rows.Scan(&v, &at); err != nil {
			return nil, fmt.Errorf("scan migration status: %w", err)
		}
		appliedMap[v] = at
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate migration status: %w", err)
	}

	var statuses []MigrationStatus
	for _, mig := range migrations {
		status := MigrationStatus{
			Version: mig.Version,
			Name:    mig.Name,
		}
		if at, ok := appliedMap[mig.Version]; ok {
			status.Applied = true
			appliedAt := at
			status.AppliedAt = &appliedAt
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}
