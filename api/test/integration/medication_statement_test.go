package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/medication"
	"github.com/google/uuid"
)

func TestMedicationStatementCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("medstmt")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "StmtPatient", "Test", "MRN-STMT-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "StmtDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *medication.MedicationStatement
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			ms := &medication.MedicationStatement{
				Status:              "active",
				CategoryCode:        ptrStr("outpatient"),
				CategoryDisplay:     ptrStr("Outpatient"),
				MedicationCode:      ptrStr("197696"),
				MedicationDisplay:   ptrStr("Lisinopril 10 MG Oral Tablet"),
				PatientID:           patient.ID,
				InformationSourceID: &practitioner.ID,
				EffectiveDatetime:   &now,
				DateAsserted:        &now,
				ReasonCode:          ptrStr("38341003"),
				ReasonDisplay:       ptrStr("Hypertension"),
				DosageText:          ptrStr("Take 1 tablet daily"),
				DosageRouteCode:     ptrStr("PO"),
				DosageRouteDisplay:  ptrStr("Oral"),
				DoseQuantity:        ptrFloat(10),
				DoseUnit:            ptrStr("mg"),
				DosageTimingCode:    ptrStr("QD"),
				DosageTimingDisplay: ptrStr("Once daily"),
				Note:                ptrStr("Patient reports taking medication as prescribed"),
			}
			if err := repo.Create(ctx, ms); err != nil {
				return err
			}
			created = ms
			return nil
		})
		if err != nil {
			t.Fatalf("Create medication statement: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			ms := &medication.MedicationStatement{
				Status:            "active",
				MedicationCode:    ptrStr("test"),
				MedicationDisplay: ptrStr("Test Med"),
				PatientID:         uuid.New(), // non-existent
			}
			return repo.Create(ctx, ms)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("Create_WithMedicationRef", func(t *testing.T) {
		med := createTestMedication(t, ctx, globalDB.Pool, tenantID, "197696-stmt", "Lisinopril 10mg")
		now := time.Now()

		var created *medication.MedicationStatement
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			ms := &medication.MedicationStatement{
				Status:       "active",
				MedicationID: &med.ID,
				PatientID:    patient.ID,
				DateAsserted: &now,
				DosageText:   ptrStr("10mg daily"),
			}
			if err := repo.Create(ctx, ms); err != nil {
				return err
			}
			created = ms
			return nil
		})
		if err != nil {
			t.Fatalf("Create with medication ref: %v", err)
		}
		if created.MedicationID == nil || *created.MedicationID != med.ID {
			t.Errorf("expected MedicationID=%s, got %v", med.ID, created.MedicationID)
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var stmtID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			ms := &medication.MedicationStatement{
				Status:            "active",
				MedicationCode:    ptrStr("314076"),
				MedicationDisplay: ptrStr("Metformin 500mg"),
				PatientID:         patient.ID,
				DateAsserted:      &now,
				DosageText:        ptrStr("500mg twice daily"),
				DoseQuantity:      ptrFloat(500),
				DoseUnit:          ptrStr("mg"),
			}
			if err := repo.Create(ctx, ms); err != nil {
				return err
			}
			stmtID = ms.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *medication.MedicationStatement
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, stmtID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected Status=active, got %s", fetched.Status)
		}
		if fetched.MedicationCode == nil || *fetched.MedicationCode != "314076" {
			t.Errorf("expected MedicationCode=314076, got %v", fetched.MedicationCode)
		}
		if fetched.MedicationDisplay == nil || *fetched.MedicationDisplay != "Metformin 500mg" {
			t.Errorf("expected MedicationDisplay=Metformin 500mg, got %v", fetched.MedicationDisplay)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected PatientID=%s, got %s", patient.ID, fetched.PatientID)
		}
		if fetched.DosageText == nil || *fetched.DosageText != "500mg twice daily" {
			t.Errorf("expected DosageText='500mg twice daily', got %v", fetched.DosageText)
		}
		if fetched.DoseQuantity == nil || *fetched.DoseQuantity != 500 {
			t.Errorf("expected DoseQuantity=500, got %v", fetched.DoseQuantity)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		now := time.Now()
		var stmt *medication.MedicationStatement
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			ms := &medication.MedicationStatement{
				Status:            "active",
				MedicationCode:    ptrStr("fhirid-test"),
				MedicationDisplay: ptrStr("FHIR ID Test Med"),
				PatientID:         patient.ID,
				DateAsserted:      &now,
			}
			if err := repo.Create(ctx, ms); err != nil {
				return err
			}
			stmt = ms
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *medication.MedicationStatement
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, stmt.FHIRID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.ID != stmt.ID {
			t.Errorf("expected ID=%s, got %s", stmt.ID, fetched.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var stmt *medication.MedicationStatement
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			ms := &medication.MedicationStatement{
				Status:            "active",
				MedicationCode:    ptrStr("update-code"),
				MedicationDisplay: ptrStr("Update Test Med"),
				PatientID:         patient.ID,
				DateAsserted:      &now,
				DosageText:        ptrStr("1 tablet daily"),
				DoseQuantity:      ptrFloat(10),
				DoseUnit:          ptrStr("mg"),
			}
			if err := repo.Create(ctx, ms); err != nil {
				return err
			}
			stmt = ms
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		// Update status and dosage
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			stmt.Status = "completed"
			stmt.StatusReasonCode = ptrStr("completed-course")
			stmt.StatusReasonDisplay = ptrStr("Completed course of treatment")
			stmt.MedicationDisplay = ptrStr("Updated Med Name")
			endTime := time.Now()
			stmt.EffectiveEnd = &endTime
			stmt.DosageText = ptrStr("2 tablets daily")
			stmt.DoseQuantity = ptrFloat(20)
			stmt.DoseUnit = ptrStr("mg")
			stmt.Note = ptrStr("Course completed successfully")
			return repo.Update(ctx, stmt)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *medication.MedicationStatement
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, stmt.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected Status=completed, got %s", fetched.Status)
		}
		if fetched.StatusReasonCode == nil || *fetched.StatusReasonCode != "completed-course" {
			t.Errorf("expected StatusReasonCode=completed-course, got %v", fetched.StatusReasonCode)
		}
		if fetched.MedicationDisplay == nil || *fetched.MedicationDisplay != "Updated Med Name" {
			t.Errorf("expected MedicationDisplay=Updated Med Name, got %v", fetched.MedicationDisplay)
		}
		if fetched.EffectiveEnd == nil {
			t.Error("expected non-nil EffectiveEnd after update")
		}
		if fetched.DosageText == nil || *fetched.DosageText != "2 tablets daily" {
			t.Errorf("expected DosageText='2 tablets daily', got %v", fetched.DosageText)
		}
		if fetched.DoseQuantity == nil || *fetched.DoseQuantity != 20 {
			t.Errorf("expected DoseQuantity=20, got %v", fetched.DoseQuantity)
		}
		if fetched.Note == nil || *fetched.Note != "Course completed successfully" {
			t.Errorf("expected Note='Course completed successfully', got %v", fetched.Note)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		// Create a few more statements for the same patient
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			ms1 := &medication.MedicationStatement{
				Status:            "active",
				MedicationCode:    ptrStr("list-code-1"),
				MedicationDisplay: ptrStr("List Med 1"),
				PatientID:         patient.ID,
				DateAsserted:      &now,
			}
			if err := repo.Create(ctx, ms1); err != nil {
				return err
			}
			ms2 := &medication.MedicationStatement{
				Status:            "active",
				MedicationCode:    ptrStr("list-code-2"),
				MedicationDisplay: ptrStr("List Med 2"),
				PatientID:         patient.ID,
				DateAsserted:      &now,
			}
			return repo.Create(ctx, ms2)
		})
		if err != nil {
			t.Fatalf("Create extra statements: %v", err)
		}

		var results []*medication.MedicationStatement
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total < 2 {
			t.Errorf("expected at least 2 medication statements for patient, got %d", total)
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*medication.MedicationStatement
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 active medication statement")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Search_ByPatient", func(t *testing.T) {
		var results []*medication.MedicationStatement
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by patient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 medication statement for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("ListByPatient_DifferentPatient", func(t *testing.T) {
		otherPatient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "OtherPatient", "Stmt", "MRN-STMT-OTHER")

		var results []*medication.MedicationStatement
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, otherPatient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient other patient: %v", err)
		}
		if total != 0 {
			t.Errorf("expected 0 medication statements for other patient, got %d", total)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("Create_WithEffectivePeriod", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		now := time.Now()
		var created *medication.MedicationStatement
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			ms := &medication.MedicationStatement{
				Status:            "completed",
				MedicationCode:    ptrStr("period-test"),
				MedicationDisplay: ptrStr("Period Test Med"),
				PatientID:         patient.ID,
				EffectiveStart:    &start,
				EffectiveEnd:      &end,
				DateAsserted:      &now,
			}
			if err := repo.Create(ctx, ms); err != nil {
				return err
			}
			created = ms
			return nil
		})
		if err != nil {
			t.Fatalf("Create with effective period: %v", err)
		}

		var fetched *medication.MedicationStatement
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, created.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.EffectiveStart == nil {
			t.Error("expected non-nil EffectiveStart")
		}
		if fetched.EffectiveEnd == nil {
			t.Error("expected non-nil EffectiveEnd")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var stmtID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			ms := &medication.MedicationStatement{
				Status:            "active",
				MedicationCode:    ptrStr("delete-test"),
				MedicationDisplay: ptrStr("Delete Test Med"),
				PatientID:         patient.ID,
				DateAsserted:      &now,
			}
			if err := repo.Create(ctx, ms); err != nil {
				return err
			}
			stmtID = ms.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			return repo.Delete(ctx, stmtID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationStatementRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, stmtID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted medication statement")
		}
	})
}
