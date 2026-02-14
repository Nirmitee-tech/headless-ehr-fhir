package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ehr/ehr/internal/domain/clinical"
	"github.com/ehr/ehr/internal/domain/encounter"
	"github.com/ehr/ehr/internal/domain/identity"
	"github.com/ehr/ehr/internal/domain/medication"
	"github.com/ehr/ehr/internal/platform/db"
)

// testDB holds the shared database infrastructure for integration tests.
type testDB struct {
	Pool          *pgxpool.Pool
	ConnStr       string
	MigrationsDir string
}

// globalDB is the package-level test database, initialized once in TestMain.
var globalDB *testDB

func TestMain(m *testing.M) {
	ctx := context.Background()

	tdb, cleanup, err := setupPostgresContainer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup postgres container: %v\n", err)
		os.Exit(1)
	}

	globalDB = tdb
	code := m.Run()
	cleanup()
	os.Exit(code)
}

// setupPostgresContainer starts a Postgres 16 container using raw Docker commands
// via the pgxpool connection. Since testcontainers-go requires network access to
// download, we use a simpler approach: connect to an existing Postgres or start one.
// For CI/testcontainers, we use the testcontainers library.
func setupPostgresContainer(ctx context.Context) (*testDB, func(), error) {
	migrationsDir := findMigrationsDir()

	// Try testcontainers first
	connStr, cleanup, err := startTestcontainersPostgres(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("start postgres container: %w", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		cleanup()
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}

	return &testDB{
		Pool:          pool,
		ConnStr:       connStr,
		MigrationsDir: migrationsDir,
	}, func() {
		pool.Close()
		cleanup()
	}, nil
}

func startTestcontainersPostgres(ctx context.Context) (string, func(), error) {
	// Use testcontainers-go to spin up postgres:16-alpine
	return startWithTestcontainers(ctx)
}

// findMigrationsDir locates the migrations directory relative to this test file.
func findMigrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	// test/integration -> api root
	apiRoot := filepath.Join(dir, "..", "..")
	return filepath.Join(apiRoot, "migrations")
}

// createTenantSchema creates a new tenant schema and runs all migrations.
func createTenantSchema(t *testing.T, ctx context.Context, tenantID string) {
	t.Helper()
	err := db.CreateTenantSchema(ctx, globalDB.Pool, tenantID, globalDB.MigrationsDir)
	if err != nil {
		t.Fatalf("create tenant schema %s: %v", tenantID, err)
	}
}

// dropTenantSchema drops a tenant schema for cleanup.
func dropTenantSchema(t *testing.T, ctx context.Context, tenantID string) {
	t.Helper()
	schema := fmt.Sprintf("tenant_%s", tenantID)
	_, err := globalDB.Pool.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))
	if err != nil {
		t.Logf("warning: failed to drop schema %s: %v", schema, err)
	}
}

// setSearchPath sets the search path for a connection to the tenant schema.
func setSearchPath(ctx context.Context, pool *pgxpool.Pool, tenantID string) error {
	schema := fmt.Sprintf("tenant_%s", tenantID)
	_, err := pool.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", schema))
	return err
}

// execWithSchema executes SQL within a specific tenant schema.
func execWithSchema(ctx context.Context, pool *pgxpool.Pool, tenantID string, sql string, args ...interface{}) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	schema := fmt.Sprintf("tenant_%s", tenantID)
	_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", schema))
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, sql, args...)
	return err
}

// withTenantConn acquires a connection, sets the search path to the tenant schema,
// and passes it to the callback. The connection is released after the callback.
func withTenantConn(ctx context.Context, pool *pgxpool.Pool, tenantID string, fn func(ctx context.Context) error) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	schema := fmt.Sprintf("tenant_%s", tenantID)
	_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", schema))
	if err != nil {
		return fmt.Errorf("set search_path: %w", err)
	}

	// Put the connection into context so repos can find it
	ctx = context.WithValue(ctx, db.DBConnKey, conn)
	return fn(ctx)
}

// connFromCtx retrieves the pgxpool.Conn from the context for direct SQL queries.
func connFromCtx(ctx context.Context) *pgxpool.Conn {
	return db.ConnFromContext(ctx)
}

// uniqueTenantID generates a unique tenant ID for test isolation.
func uniqueTenantID(prefix string) string {
	short := strings.ReplaceAll(uuid.New().String()[:8], "-", "")
	return fmt.Sprintf("%s_%s", prefix, short)
}

// Helper to create a test organization
func createTestOrganization(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	fhirID := id.String()
	err := execWithSchema(ctx, pool, tenantID,
		`INSERT INTO organization (id, fhir_id, name, type_code, active)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, fhirID, "Test Hospital", "prov", true)
	if err != nil {
		t.Fatalf("create test organization: %v", err)
	}
	return id
}

// Helper to create a test practitioner using the repo
func createTestPractitioner(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID, firstName, lastName string) *identity.Practitioner {
	t.Helper()
	var result *identity.Practitioner
	err := withTenantConn(ctx, pool, tenantID, func(ctx context.Context) error {
		repo := identity.NewPractitionerRepo(pool)
		p := &identity.Practitioner{
			Active:    true,
			FirstName: firstName,
			LastName:  lastName,
		}
		if err := repo.Create(ctx, p); err != nil {
			return err
		}
		result = p
		return nil
	})
	if err != nil {
		t.Fatalf("create test practitioner: %v", err)
	}
	return result
}

// Helper to create a test patient using the repo
func createTestPatient(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID, firstName, lastName, mrn string) *identity.Patient {
	t.Helper()
	var result *identity.Patient
	gender := "male"
	err := withTenantConn(ctx, pool, tenantID, func(ctx context.Context) error {
		repo := identity.NewPatientRepo(pool)
		p := &identity.Patient{
			Active:    true,
			MRN:       mrn,
			FirstName: firstName,
			LastName:  lastName,
			Gender:    &gender,
		}
		if err := repo.Create(ctx, p); err != nil {
			return err
		}
		result = p
		return nil
	})
	if err != nil {
		t.Fatalf("create test patient: %v", err)
	}
	return result
}

// Helper to create a test encounter
func createTestEncounter(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID string, patientID uuid.UUID, practitionerID *uuid.UUID) *encounter.Encounter {
	t.Helper()
	var result *encounter.Encounter
	err := withTenantConn(ctx, pool, tenantID, func(ctx context.Context) error {
		repo := encounter.NewRepo(pool)
		enc := &encounter.Encounter{
			Status:                "in-progress",
			ClassCode:             "AMB",
			PatientID:             patientID,
			PrimaryPractitionerID: practitionerID,
			PeriodStart:           time.Now(),
		}
		if err := repo.Create(ctx, enc); err != nil {
			return err
		}
		result = enc
		return nil
	})
	if err != nil {
		t.Fatalf("create test encounter: %v", err)
	}
	return result
}

// Helper to create a test medication
func createTestMedication(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID, codeValue, codeDisplay string) *medication.Medication {
	t.Helper()
	var result *medication.Medication
	err := withTenantConn(ctx, pool, tenantID, func(ctx context.Context) error {
		repo := medication.NewMedicationRepoPG(pool)
		med := &medication.Medication{
			CodeValue:   codeValue,
			CodeDisplay: codeDisplay,
			Status:      "active",
		}
		if err := repo.Create(ctx, med); err != nil {
			return err
		}
		result = med
		return nil
	})
	if err != nil {
		t.Fatalf("create test medication: %v", err)
	}
	return result
}

// Helper to create a test condition
func createTestCondition(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tenantID string, patientID uuid.UUID) *clinical.Condition {
	t.Helper()
	var result *clinical.Condition
	err := withTenantConn(ctx, pool, tenantID, func(ctx context.Context) error {
		repo := clinical.NewConditionRepoPG(pool)
		cond := &clinical.Condition{
			PatientID:      patientID,
			ClinicalStatus: "active",
			CodeValue:      "38341003",
			CodeDisplay:    "Hypertension",
		}
		if err := repo.Create(ctx, cond); err != nil {
			return err
		}
		result = cond
		return nil
	})
	if err != nil {
		t.Fatalf("create test condition: %v", err)
	}
	return result
}

// runMigrationsManually runs migration SQL files against a schema directly,
// used as a fallback if the Migrator approach has issues.
func runMigrationsManually(ctx context.Context, pool *pgxpool.Pool, schema, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type migFile struct {
		version int
		name    string
	}
	var files []migFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		parts := strings.SplitN(e.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}
		v, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		files = append(files, migFile{version: v, name: e.Name()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].version < files[j].version
	})

	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(migrationsDir, f.name))
		if err != nil {
			return fmt.Errorf("read %s: %w", f.name, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", f.name, err)
		}

		_, err = tx.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", schema))
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("set search_path for %s: %w", f.name, err)
		}

		_, err = tx.Exec(ctx, string(content))
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("exec %s: %w", f.name, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", f.name, err)
		}
	}

	return nil
}

// ptrStr returns a pointer to the given string.
func ptrStr(s string) *string { return &s }

// ptrFloat returns a pointer to the given float64.
func ptrFloat(f float64) *float64 { return &f }

// ptrInt returns a pointer to the given int.
func ptrInt(i int) *int { return &i }

// ptrBool returns a pointer to the given bool.
func ptrBool(b bool) *bool { return &b }

// ptrTime returns a pointer to the given time.
func ptrTime(t time.Time) *time.Time { return &t }

// ptrUUID returns a pointer to the given UUID.
func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }
