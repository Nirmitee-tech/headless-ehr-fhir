package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/encounter"
	"github.com/google/uuid"
)

func TestEncounterLifecycle(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("enc")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	// Create prerequisite data
	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "EncPatient", "Test", "MRN-ENC-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "EncDoc", "Smith")

	t.Run("Create_Encounter", func(t *testing.T) {
		var created *encounter.Encounter
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			now := time.Now()
			enc := &encounter.Encounter{
				Status:                "planned",
				ClassCode:             "AMB",
				ClassDisplay:          ptrStr("ambulatory"),
				PatientID:             patient.ID,
				PrimaryPractitionerID: &practitioner.ID,
				PeriodStart:           now,
				ReasonText:            ptrStr("Annual checkup"),
				IsTelehealth:          false,
			}
			if err := repo.Create(ctx, enc); err != nil {
				return err
			}
			created = enc
			return nil
		})
		if err != nil {
			t.Fatalf("Create encounter: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_With_FK_Validation", func(t *testing.T) {
		// Attempt to create encounter with non-existent patient ID
		fakePatientID := uuid.New()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			enc := &encounter.Encounter{
				Status:      "planned",
				ClassCode:   "AMB",
				PatientID:   fakePatientID,
				PeriodStart: time.Now(),
			}
			return repo.Create(ctx, enc)
		})
		if err == nil {
			t.Fatal("expected FK violation error for non-existent patient, got nil")
		}
	})

	t.Run("GetByID_and_Update", func(t *testing.T) {
		enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

		// Get
		var fetched *encounter.Encounter
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, enc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "in-progress" {
			t.Errorf("expected status=in-progress, got %s", fetched.Status)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}

		// Update to finished
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			fetched.Status = "finished"
			now := time.Now()
			fetched.PeriodEnd = &now
			lengthMin := 30
			fetched.LengthMinutes = &lengthMin
			return repo.Update(ctx, fetched)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		// Verify update
		var updated *encounter.Encounter
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			var err error
			updated, err = repo.GetByID(ctx, enc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if updated.Status != "finished" {
			t.Errorf("expected status=finished, got %s", updated.Status)
		}
		if updated.PeriodEnd == nil {
			t.Error("expected non-nil PeriodEnd after update")
		}
		if updated.LengthMinutes == nil || *updated.LengthMinutes != 30 {
			t.Errorf("expected LengthMinutes=30, got %v", updated.LengthMinutes)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		// Create a few encounters for the same patient
		createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)
		createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

		var results []*encounter.Encounter
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total < 2 {
			t.Errorf("expected at least 2 encounters for patient, got %d", total)
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*encounter.Encounter
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "in-progress",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 in-progress encounter")
		}
		for _, r := range results {
			if r.Status != "in-progress" {
				t.Errorf("expected status=in-progress, got %s", r.Status)
			}
		}
	})

	t.Run("Search_ByClass", func(t *testing.T) {
		var results []*encounter.Encounter
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"class": "AMB",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by class: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 AMB encounter")
		}
		for _, r := range results {
			if r.ClassCode != "AMB" {
				t.Errorf("expected class_code=AMB, got %s", r.ClassCode)
			}
		}
	})

	t.Run("Participants", func(t *testing.T) {
		enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)
		consultant := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ConsultDoc", "Jones")

		// Add participant
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			p := &encounter.EncounterParticipant{
				EncounterID:    enc.ID,
				PractitionerID: consultant.ID,
				TypeCode:       "CON",
				TypeDisplay:    ptrStr("consultant"),
			}
			return repo.AddParticipant(ctx, p)
		})
		if err != nil {
			t.Fatalf("AddParticipant: %v", err)
		}

		// Get participants
		var parts []*encounter.EncounterParticipant
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			var err error
			parts, err = repo.GetParticipants(ctx, enc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetParticipants: %v", err)
		}
		if len(parts) != 1 {
			t.Fatalf("expected 1 participant, got %d", len(parts))
		}
		if parts[0].TypeCode != "CON" {
			t.Errorf("expected type_code=CON, got %s", parts[0].TypeCode)
		}
		if parts[0].PractitionerID != consultant.ID {
			t.Errorf("expected practitioner_id=%s, got %s", consultant.ID, parts[0].PractitionerID)
		}

		// FK violation: add participant with non-existent practitioner
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			p := &encounter.EncounterParticipant{
				EncounterID:    enc.ID,
				PractitionerID: uuid.New(), // non-existent
				TypeCode:       "REF",
			}
			return repo.AddParticipant(ctx, p)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent practitioner participant")
		}
	})

	t.Run("StatusHistory", func(t *testing.T) {
		enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			sh := &encounter.EncounterStatusHistory{
				EncounterID: enc.ID,
				Status:      "planned",
				PeriodStart: now.Add(-1 * time.Hour),
				PeriodEnd:   &now,
			}
			return repo.AddStatusHistory(ctx, sh)
		})
		if err != nil {
			t.Fatalf("AddStatusHistory: %v", err)
		}

		var history []*encounter.EncounterStatusHistory
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			var err error
			history, err = repo.GetStatusHistory(ctx, enc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetStatusHistory: %v", err)
		}
		if len(history) != 1 {
			t.Fatalf("expected 1 status history entry, got %d", len(history))
		}
		if history[0].Status != "planned" {
			t.Errorf("expected status=planned, got %s", history[0].Status)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			return repo.Delete(ctx, enc.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := encounter.NewRepo(globalDB.Pool)
			_, err := repo.GetByID(ctx, enc.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted encounter")
		}
	})
}
