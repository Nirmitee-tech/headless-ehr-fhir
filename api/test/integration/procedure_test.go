package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/clinical"
	"github.com/google/uuid"
)

func TestProcedureRecordCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("proc")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ProcPatient", "Test", "MRN-PROC-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ProcDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *clinical.ProcedureRecord
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			proc := &clinical.ProcedureRecord{
				Status:            "completed",
				PatientID:         patient.ID,
				RecorderID:        &practitioner.ID,
				CodeSystem:        ptrStr("http://snomed.info/sct"),
				CodeValue:         "80146002",
				CodeDisplay:       "Appendectomy",
				CategoryCode:      ptrStr("24642003"),
				CategoryDisplay:   ptrStr("Surgical procedure"),
				PerformedDatetime: &now,
				BodySiteCode:      ptrStr("66754008"),
				BodySiteDisplay:   ptrStr("Appendix"),
				OutcomeCode:       ptrStr("385669000"),
				OutcomeDisplay:    ptrStr("Successful"),
				AnesthesiaType:    ptrStr("general"),
				CPTCode:           ptrStr("44970"),
				Note:              ptrStr("Laparoscopic appendectomy without complications"),
			}
			if err := repo.Create(ctx, proc); err != nil {
				return err
			}
			created = proc
			return nil
		})
		if err != nil {
			t.Fatalf("Create procedure: %v", err)
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
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			proc := &clinical.ProcedureRecord{
				Status:    "completed",
				PatientID: uuid.New(), // non-existent
				CodeValue: "80146002",
				CodeDisplay: "Appendectomy",
			}
			return repo.Create(ctx, proc)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var procID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			proc := &clinical.ProcedureRecord{
				Status:            "completed",
				PatientID:         patient.ID,
				CodeValue:         "27687003",
				CodeDisplay:       "Colonoscopy",
				PerformedDatetime: &now,
			}
			if err := repo.Create(ctx, proc); err != nil {
				return err
			}
			procID = proc.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.ProcedureRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, procID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CodeValue != "27687003" {
			t.Errorf("expected CodeValue=27687003, got %s", fetched.CodeValue)
		}
		if fetched.CodeDisplay != "Colonoscopy" {
			t.Errorf("expected CodeDisplay=Colonoscopy, got %s", fetched.CodeDisplay)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected Status=completed, got %s", fetched.Status)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected PatientID=%s, got %s", patient.ID, fetched.PatientID)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var proc *clinical.ProcedureRecord
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			p := &clinical.ProcedureRecord{
				Status:      "completed",
				PatientID:   patient.ID,
				CodeValue:   "11101003",
				CodeDisplay: "Biopsy",
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			proc = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.ProcedureRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, proc.FHIRID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.ID != proc.ID {
			t.Errorf("expected ID=%s, got %s", proc.ID, fetched.ID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var proc *clinical.ProcedureRecord
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			p := &clinical.ProcedureRecord{
				Status:      "in-progress",
				PatientID:   patient.ID,
				CodeValue:   "387713003",
				CodeDisplay: "Surgical procedure on knee",
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			proc = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			proc.Status = "completed"
			proc.OutcomeCode = ptrStr("385669000")
			proc.OutcomeDisplay = ptrStr("Successful")
			proc.Note = ptrStr("Procedure completed without complications")
			return repo.Update(ctx, proc)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *clinical.ProcedureRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, proc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected Status=completed, got %s", fetched.Status)
		}
		if fetched.OutcomeCode == nil || *fetched.OutcomeCode != "385669000" {
			t.Errorf("expected OutcomeCode=385669000, got %v", fetched.OutcomeCode)
		}
		if fetched.OutcomeDisplay == nil || *fetched.OutcomeDisplay != "Successful" {
			t.Errorf("expected OutcomeDisplay=Successful, got %v", fetched.OutcomeDisplay)
		}
		if fetched.Note == nil || *fetched.Note != "Procedure completed without complications" {
			t.Errorf("expected Note='Procedure completed without complications', got %v", fetched.Note)
		}
	})

	t.Run("Update_With_Complication", func(t *testing.T) {
		var proc *clinical.ProcedureRecord
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			p := &clinical.ProcedureRecord{
				Status:      "completed",
				PatientID:   patient.ID,
				CodeValue:   "236211003",
				CodeDisplay: "Hernia repair",
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			proc = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			proc.ComplicationCode = ptrStr("131148009")
			proc.ComplicationDisp = ptrStr("Bleeding")
			return repo.Update(ctx, proc)
		})
		if err != nil {
			t.Fatalf("Update with complication: %v", err)
		}

		var fetched *clinical.ProcedureRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, proc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.ComplicationCode == nil || *fetched.ComplicationCode != "131148009" {
			t.Errorf("expected ComplicationCode=131148009, got %v", fetched.ComplicationCode)
		}
		if fetched.ComplicationDisp == nil || *fetched.ComplicationDisp != "Bleeding" {
			t.Errorf("expected ComplicationDisp=Bleeding, got %v", fetched.ComplicationDisp)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		// Create additional procedure for the same patient
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			p := &clinical.ProcedureRecord{
				Status:      "completed",
				PatientID:   patient.ID,
				CodeValue:   "274025005",
				CodeDisplay: "Blood transfusion",
			}
			return repo.Create(ctx, p)
		})
		if err != nil {
			t.Fatalf("Create extra procedure: %v", err)
		}

		var results []*clinical.ProcedureRecord
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total < 2 {
			t.Errorf("expected at least 2 procedures for patient, got %d", total)
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("AddPerformer_GetPerformers_RemovePerformer", func(t *testing.T) {
		// Create a procedure for performer tests
		var procID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			proc := &clinical.ProcedureRecord{
				Status:      "completed",
				PatientID:   patient.ID,
				CodeValue:   "18949003",
				CodeDisplay: "Change of dressing",
			}
			if err := repo.Create(ctx, proc); err != nil {
				return err
			}
			procID = proc.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create procedure: %v", err)
		}

		// Add performer
		var performerID uuid.UUID
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			pf := &clinical.ProcedurePerformer{
				ProcedureID:    procID,
				PractitionerID: practitioner.ID,
				RoleCode:       ptrStr("surgeon"),
				RoleDisplay:    ptrStr("Surgeon"),
			}
			if err := repo.AddPerformer(ctx, pf); err != nil {
				return err
			}
			performerID = pf.ID
			return nil
		})
		if err != nil {
			t.Fatalf("AddPerformer: %v", err)
		}
		if performerID == uuid.Nil {
			t.Fatal("expected non-nil performer ID")
		}

		// Get performers
		var performers []*clinical.ProcedurePerformer
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			var err error
			performers, err = repo.GetPerformers(ctx, procID)
			return err
		})
		if err != nil {
			t.Fatalf("GetPerformers: %v", err)
		}
		if len(performers) != 1 {
			t.Fatalf("expected 1 performer, got %d", len(performers))
		}
		if performers[0].PractitionerID != practitioner.ID {
			t.Errorf("expected PractitionerID=%s, got %s", practitioner.ID, performers[0].PractitionerID)
		}
		if performers[0].RoleCode == nil || *performers[0].RoleCode != "surgeon" {
			t.Errorf("expected RoleCode=surgeon, got %v", performers[0].RoleCode)
		}

		// Remove performer
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			return repo.RemovePerformer(ctx, performerID)
		})
		if err != nil {
			t.Fatalf("RemovePerformer: %v", err)
		}

		// Verify removal
		var performersAfter []*clinical.ProcedurePerformer
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			var err error
			performersAfter, err = repo.GetPerformers(ctx, procID)
			return err
		})
		if err != nil {
			t.Fatalf("GetPerformers after remove: %v", err)
		}
		if len(performersAfter) != 0 {
			t.Errorf("expected 0 performers after remove, got %d", len(performersAfter))
		}
	})

	t.Run("MultiplePerformers", func(t *testing.T) {
		consultant := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ConsultDoc", "Perf")

		var procID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			proc := &clinical.ProcedureRecord{
				Status:      "completed",
				PatientID:   patient.ID,
				CodeValue:   "73761001",
				CodeDisplay: "Colonoscopy procedure",
			}
			if err := repo.Create(ctx, proc); err != nil {
				return err
			}
			procID = proc.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		// Add two performers
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			pf1 := &clinical.ProcedurePerformer{
				ProcedureID:    procID,
				PractitionerID: practitioner.ID,
				RoleCode:       ptrStr("surgeon"),
			}
			if err := repo.AddPerformer(ctx, pf1); err != nil {
				return err
			}
			pf2 := &clinical.ProcedurePerformer{
				ProcedureID:    procID,
				PractitionerID: consultant.ID,
				RoleCode:       ptrStr("assistant"),
			}
			return repo.AddPerformer(ctx, pf2)
		})
		if err != nil {
			t.Fatalf("Add multiple performers: %v", err)
		}

		var performers []*clinical.ProcedurePerformer
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			var err error
			performers, err = repo.GetPerformers(ctx, procID)
			return err
		})
		if err != nil {
			t.Fatalf("GetPerformers: %v", err)
		}
		if len(performers) != 2 {
			t.Fatalf("expected 2 performers, got %d", len(performers))
		}
	})

	t.Run("Performer_FK_Violation", func(t *testing.T) {
		var procID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			proc := &clinical.ProcedureRecord{
				Status:      "completed",
				PatientID:   patient.ID,
				CodeValue:   "999999",
				CodeDisplay: "FK Test Procedure",
			}
			if err := repo.Create(ctx, proc); err != nil {
				return err
			}
			procID = proc.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			pf := &clinical.ProcedurePerformer{
				ProcedureID:    procID,
				PractitionerID: uuid.New(), // non-existent
				RoleCode:       ptrStr("surgeon"),
			}
			return repo.AddPerformer(ctx, pf)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent practitioner performer")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var procID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			proc := &clinical.ProcedureRecord{
				Status:      "completed",
				PatientID:   patient.ID,
				CodeValue:   "delete-test",
				CodeDisplay: "Delete Test Procedure",
			}
			if err := repo.Create(ctx, proc); err != nil {
				return err
			}
			procID = proc.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			return repo.Delete(ctx, procID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewProcedureRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, procID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted procedure")
		}
	})
}
