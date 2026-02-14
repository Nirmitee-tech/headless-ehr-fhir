package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestMultiTenantIsolation(t *testing.T) {
	ctx := context.Background()
	tenantA := uniqueTenantID("tenantA")
	tenantB := uniqueTenantID("tenantB")

	createTenantSchema(t, ctx, tenantA)
	defer dropTenantSchema(t, ctx, tenantA)
	createTenantSchema(t, ctx, tenantB)
	defer dropTenantSchema(t, ctx, tenantB)

	t.Run("Patient_Isolation", func(t *testing.T) {
		// Create patients in tenant A
		pA1 := createTestPatient(t, ctx, globalDB.Pool, tenantA, "AliceTenantA", "Smith", "MRN-A-001")
		pA2 := createTestPatient(t, ctx, globalDB.Pool, tenantA, "BobTenantA", "Jones", "MRN-A-002")

		// Create patients in tenant B
		pB1 := createTestPatient(t, ctx, globalDB.Pool, tenantB, "CharlieTenantB", "Brown", "MRN-B-001")

		// Verify tenant A sees only its patients
		var totalA int
		err := withTenantConn(ctx, globalDB.Pool, tenantA, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			return conn.QueryRow(ctx, "SELECT COUNT(*) FROM patient").Scan(&totalA)
		})
		if err != nil {
			t.Fatalf("count patients in tenant A: %v", err)
		}
		if totalA != 2 {
			t.Errorf("expected 2 patients in tenant A, got %d", totalA)
		}

		// Verify tenant B sees only its patients
		var totalB int
		err = withTenantConn(ctx, globalDB.Pool, tenantB, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			return conn.QueryRow(ctx, "SELECT COUNT(*) FROM patient").Scan(&totalB)
		})
		if err != nil {
			t.Fatalf("count patients in tenant B: %v", err)
		}
		if totalB != 1 {
			t.Errorf("expected 1 patient in tenant B, got %d", totalB)
		}

		// Verify IDs don't cross tenants: tenant B cannot see tenant A patients
		err = withTenantConn(ctx, globalDB.Pool, tenantB, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			var count int
			err := conn.QueryRow(ctx,
				"SELECT COUNT(*) FROM patient WHERE id = $1", pA1.ID).Scan(&count)
			if err != nil {
				return err
			}
			if count != 0 {
				return fmt.Errorf("tenant B should not see tenant A patient (pA1), found %d", count)
			}
			err = conn.QueryRow(ctx,
				"SELECT COUNT(*) FROM patient WHERE id = $1", pA2.ID).Scan(&count)
			if err != nil {
				return err
			}
			if count != 0 {
				return fmt.Errorf("tenant B should not see tenant A patient (pA2), found %d", count)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("cross-tenant visibility check: %v", err)
		}

		// Verify tenant A cannot see tenant B patients
		err = withTenantConn(ctx, globalDB.Pool, tenantA, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			var count int
			err := conn.QueryRow(ctx,
				"SELECT COUNT(*) FROM patient WHERE id = $1", pB1.ID).Scan(&count)
			if err != nil {
				return err
			}
			if count != 0 {
				return fmt.Errorf("tenant A should not see tenant B patient (pB1), found %d", count)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("cross-tenant visibility check (reverse): %v", err)
		}
	})

	t.Run("Same_MRN_Different_Tenants", func(t *testing.T) {
		// Both tenants should allow the same MRN since they're in different schemas
		createTestPatient(t, ctx, globalDB.Pool, tenantA, "SharedMRN_A", "Last_A", "MRN-SHARED-001")
		createTestPatient(t, ctx, globalDB.Pool, tenantB, "SharedMRN_B", "Last_B", "MRN-SHARED-001")

		// Verify each tenant sees its own patient with MRN-SHARED-001
		err := withTenantConn(ctx, globalDB.Pool, tenantA, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			var firstName string
			err := conn.QueryRow(ctx,
				"SELECT first_name FROM patient WHERE mrn = $1", "MRN-SHARED-001").Scan(&firstName)
			if err != nil {
				return err
			}
			if firstName != "SharedMRN_A" {
				return fmt.Errorf("expected SharedMRN_A in tenant A, got %s", firstName)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("tenant A MRN lookup: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantB, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			var firstName string
			err := conn.QueryRow(ctx,
				"SELECT first_name FROM patient WHERE mrn = $1", "MRN-SHARED-001").Scan(&firstName)
			if err != nil {
				return err
			}
			if firstName != "SharedMRN_B" {
				return fmt.Errorf("expected SharedMRN_B in tenant B, got %s", firstName)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("tenant B MRN lookup: %v", err)
		}
	})

	t.Run("Practitioner_Isolation", func(t *testing.T) {
		createTestPractitioner(t, ctx, globalDB.Pool, tenantA, "DocTenantA", "MD")
		createTestPractitioner(t, ctx, globalDB.Pool, tenantB, "DocTenantB1", "DO")
		createTestPractitioner(t, ctx, globalDB.Pool, tenantB, "DocTenantB2", "MD")

		var totalA, totalB int
		err := withTenantConn(ctx, globalDB.Pool, tenantA, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			return conn.QueryRow(ctx, "SELECT COUNT(*) FROM practitioner").Scan(&totalA)
		})
		if err != nil {
			t.Fatalf("count practitioners in tenant A: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantB, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			return conn.QueryRow(ctx, "SELECT COUNT(*) FROM practitioner").Scan(&totalB)
		})
		if err != nil {
			t.Fatalf("count practitioners in tenant B: %v", err)
		}

		if totalA != 1 {
			t.Errorf("expected 1 practitioner in tenant A, got %d", totalA)
		}
		if totalB != 2 {
			t.Errorf("expected 2 practitioners in tenant B, got %d", totalB)
		}
	})

	t.Run("Encounter_Isolation", func(t *testing.T) {
		patA := createTestPatient(t, ctx, globalDB.Pool, tenantA, "EncPatA", "Test", "MRN-ENC-ISO-A")
		patB := createTestPatient(t, ctx, globalDB.Pool, tenantB, "EncPatB", "Test", "MRN-ENC-ISO-B")

		// Create encounters in each tenant
		createTestEncounter(t, ctx, globalDB.Pool, tenantA, patA.ID, nil)
		createTestEncounter(t, ctx, globalDB.Pool, tenantA, patA.ID, nil)
		createTestEncounter(t, ctx, globalDB.Pool, tenantB, patB.ID, nil)

		var totalA, totalB int
		err := withTenantConn(ctx, globalDB.Pool, tenantA, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			return conn.QueryRow(ctx, "SELECT COUNT(*) FROM encounter").Scan(&totalA)
		})
		if err != nil {
			t.Fatalf("count encounters in tenant A: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantB, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			return conn.QueryRow(ctx, "SELECT COUNT(*) FROM encounter").Scan(&totalB)
		})
		if err != nil {
			t.Fatalf("count encounters in tenant B: %v", err)
		}

		if totalA != 2 {
			t.Errorf("expected 2 encounters in tenant A, got %d", totalA)
		}
		if totalB != 1 {
			t.Errorf("expected 1 encounter in tenant B, got %d", totalB)
		}
	})

	t.Run("Schema_Existence", func(t *testing.T) {
		// Verify both schemas actually exist in the database
		// Note: PostgreSQL lowercases unquoted identifiers, so schema names are lowercase
		for _, tid := range []string{tenantA, tenantB} {
			schema := strings.ToLower(fmt.Sprintf("tenant_%s", tid))
			var exists bool
			err := globalDB.Pool.QueryRow(ctx,
				"SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = $1)",
				schema).Scan(&exists)
			if err != nil {
				t.Fatalf("check schema existence for %s: %v", schema, err)
			}
			if !exists {
				t.Errorf("schema %s should exist", schema)
			}
		}
	})

	t.Run("Tables_Exist_In_Each_Schema", func(t *testing.T) {
		expectedTables := []string{
			"patient", "practitioner", "encounter", "organization",
			"condition", "observation", "allergy_intolerance",
			"medication", "medication_request",
		}

		for _, tid := range []string{tenantA, tenantB} {
			schema := strings.ToLower(fmt.Sprintf("tenant_%s", tid))
			for _, table := range expectedTables {
				var exists bool
				err := globalDB.Pool.QueryRow(ctx,
					`SELECT EXISTS(
						SELECT 1 FROM information_schema.tables
						WHERE table_schema = $1 AND table_name = $2
					)`, schema, table).Scan(&exists)
				if err != nil {
					t.Fatalf("check table %s.%s: %v", schema, table, err)
				}
				if !exists {
					t.Errorf("table %s.%s should exist", schema, table)
				}
			}
		}
	})

	t.Run("Cross_Tenant_FK_Cannot_Reference", func(t *testing.T) {
		// Create a patient in tenant A
		patA := createTestPatient(t, ctx, globalDB.Pool, tenantA, "FKCrossA", "Test", "MRN-FKCROSS-A")

		// Try to create an encounter in tenant B referencing tenant A's patient
		// This should fail because the patient doesn't exist in tenant B's schema
		err := withTenantConn(ctx, globalDB.Pool, tenantB, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			_, err := conn.Exec(ctx,
				`INSERT INTO encounter (id, fhir_id, status, class_code, patient_id, period_start)
				 VALUES (gen_random_uuid(), gen_random_uuid()::text, 'planned', 'AMB', $1, NOW())`,
				patA.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected FK violation when referencing cross-tenant patient")
		}
	})
}

func TestMultiTenantDirectSQL(t *testing.T) {
	// This test uses direct SQL (no repos) to verify multi-tenant isolation
	// at the database level, ensuring search_path controls visibility.
	ctx := context.Background()
	tenantC := uniqueTenantID("tenantC")
	tenantD := uniqueTenantID("tenantD")

	createTenantSchema(t, ctx, tenantC)
	defer dropTenantSchema(t, ctx, tenantC)
	createTenantSchema(t, ctx, tenantD)
	defer dropTenantSchema(t, ctx, tenantD)

	t.Run("DirectSQL_Insert_And_Query", func(t *testing.T) {
		// Insert into tenant C
		err := withTenantConn(ctx, globalDB.Pool, tenantC, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			_, err := conn.Exec(ctx,
				`INSERT INTO organization (id, fhir_id, name, active) VALUES (gen_random_uuid(), 'org-c-1', 'Org C', true)`)
			return err
		})
		if err != nil {
			t.Fatalf("insert org in tenant C: %v", err)
		}

		// Insert into tenant D (2 orgs)
		err = withTenantConn(ctx, globalDB.Pool, tenantD, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			_, err := conn.Exec(ctx,
				`INSERT INTO organization (id, fhir_id, name, active) VALUES (gen_random_uuid(), 'org-d-1', 'Org D1', true)`)
			if err != nil {
				return err
			}
			_, err = conn.Exec(ctx,
				`INSERT INTO organization (id, fhir_id, name, active) VALUES (gen_random_uuid(), 'org-d-2', 'Org D2', true)`)
			return err
		})
		if err != nil {
			t.Fatalf("insert orgs in tenant D: %v", err)
		}

		// Query tenant C - should see 1 org
		var countC int
		err = withTenantConn(ctx, globalDB.Pool, tenantC, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			return conn.QueryRow(ctx, "SELECT COUNT(*) FROM organization").Scan(&countC)
		})
		if err != nil {
			t.Fatalf("count orgs in C: %v", err)
		}
		if countC != 1 {
			t.Errorf("expected 1 org in tenant C, got %d", countC)
		}

		// Query tenant D - should see 2 orgs
		var countD int
		err = withTenantConn(ctx, globalDB.Pool, tenantD, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			return conn.QueryRow(ctx, "SELECT COUNT(*) FROM organization").Scan(&countD)
		})
		if err != nil {
			t.Fatalf("count orgs in D: %v", err)
		}
		if countD != 2 {
			t.Errorf("expected 2 orgs in tenant D, got %d", countD)
		}

		// Verify tenant C cannot see tenant D's org by fhir_id
		err = withTenantConn(ctx, globalDB.Pool, tenantC, func(ctx context.Context) error {
			conn := connFromCtx(ctx)
			var name string
			err := conn.QueryRow(ctx, "SELECT name FROM organization WHERE fhir_id = 'org-d-1'").Scan(&name)
			if err == pgx.ErrNoRows {
				return nil // expected: tenant C can't see tenant D data
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("tenant C should NOT see tenant D's org, but found: %s", name)
		})
		if err != nil {
			t.Fatalf("cross-tenant org visibility: %v", err)
		}
	})
}
