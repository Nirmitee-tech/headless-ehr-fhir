package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/clinical"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Flag
// ---------------------------------------------------------------------------

func TestFlagCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("flag")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "FlagPatient", "Test", "MRN-FLAG-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "FlagDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *clinical.Flag
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			f := &clinical.Flag{
				Status:               "active",
				CategoryCode:         ptrStr("clinical"),
				CodeCode:             "SAFETY-001",
				CodeDisplay:          ptrStr("Fall risk"),
				SubjectPatientID:     ptrUUID(patient.ID),
				PeriodStart:          ptrTime(now),
				AuthorPractitionerID: &practitioner.ID,
			}
			if err := repo.Create(ctx, f); err != nil {
				return err
			}
			created = f
			return nil
		})
		if err != nil {
			t.Fatalf("Create flag: %v", err)
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
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			fakePatient := uuid.New()
			f := &clinical.Flag{
				Status:           "active",
				CodeCode:         "SAFETY-002",
				CodeDisplay:      ptrStr("Test flag"),
				SubjectPatientID: &fakePatient,
			}
			return repo.Create(ctx, f)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var flagID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			f := &clinical.Flag{
				Status:           "active",
				CodeCode:         "SAFETY-003",
				CodeDisplay:      ptrStr("Elopement risk"),
				SubjectPatientID: ptrUUID(patient.ID),
				PeriodStart:      ptrTime(time.Now()),
			}
			if err := repo.Create(ctx, f); err != nil {
				return err
			}
			flagID = f.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.Flag
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, flagID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CodeCode != "SAFETY-003" {
			t.Errorf("expected code=SAFETY-003, got %s", fetched.CodeCode)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			f := &clinical.Flag{
				Status:           "active",
				CodeCode:         "SAFETY-004",
				CodeDisplay:      ptrStr("Isolation required"),
				SubjectPatientID: ptrUUID(patient.ID),
				PeriodStart:      ptrTime(time.Now()),
			}
			if err := repo.Create(ctx, f); err != nil {
				return err
			}
			fhirID = f.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.Flag
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.FHIRID != fhirID {
			t.Errorf("expected fhir_id=%s, got %s", fhirID, fetched.FHIRID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var flag *clinical.Flag
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			f := &clinical.Flag{
				Status:           "active",
				CodeCode:         "SAFETY-005",
				CodeDisplay:      ptrStr("Aggressive behavior"),
				SubjectPatientID: ptrUUID(patient.ID),
				PeriodStart:      ptrTime(now),
			}
			if err := repo.Create(ctx, f); err != nil {
				return err
			}
			flag = f
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			flag.Status = "inactive"
			end := time.Now()
			flag.PeriodEnd = &end
			return repo.Update(ctx, flag)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *clinical.Flag
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, flag.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "inactive" {
			t.Errorf("expected status=inactive, got %s", fetched.Status)
		}
		if fetched.PeriodEnd == nil {
			t.Error("expected non-nil PeriodEnd")
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*clinical.Flag
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 flag")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var flagID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			f := &clinical.Flag{
				Status:           "active",
				CodeCode:         "SAFETY-DEL",
				CodeDisplay:      ptrStr("Delete test flag"),
				SubjectPatientID: ptrUUID(patient.ID),
			}
			if err := repo.Create(ctx, f); err != nil {
				return err
			}
			flagID = f.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			return repo.Delete(ctx, flagID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewFlagRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, flagID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted flag")
		}
	})
}

// ---------------------------------------------------------------------------
// DetectedIssue
// ---------------------------------------------------------------------------

func TestDetectedIssueCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("detiss")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "DIPatient", "Test", "MRN-DI-001")

	t.Run("Create", func(t *testing.T) {
		var created *clinical.DetectedIssue
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			d := &clinical.DetectedIssue{
				Status:         "final",
				CodeCode:       ptrStr("DRG-INT"),
				CodeDisplay:    ptrStr("Drug interaction"),
				Severity:       ptrStr("high"),
				PatientID:      ptrUUID(patient.ID),
				IdentifiedDate: ptrTime(now),
				Detail:         ptrStr("Potential interaction between warfarin and aspirin"),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			created = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create detected issue: %v", err)
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
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			fakePatient := uuid.New()
			d := &clinical.DetectedIssue{
				Status:      "final",
				CodeCode:    ptrStr("DRG-INT"),
				CodeDisplay: ptrStr("Drug interaction"),
				PatientID:   &fakePatient,
			}
			return repo.Create(ctx, d)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var issueID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			d := &clinical.DetectedIssue{
				Status:      "preliminary",
				CodeCode:    ptrStr("DUP-THER"),
				CodeDisplay: ptrStr("Duplicate therapy"),
				PatientID:   ptrUUID(patient.ID),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			issueID = d.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.DetectedIssue
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, issueID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CodeCode == nil || *fetched.CodeCode != "DUP-THER" {
			t.Errorf("expected code=DUP-THER, got %v", fetched.CodeCode)
		}
		if fetched.Status != "preliminary" {
			t.Errorf("expected status=preliminary, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			d := &clinical.DetectedIssue{
				Status:      "final",
				CodeCode:    ptrStr("DOSE-ERR"),
				CodeDisplay: ptrStr("Dosing error"),
				PatientID:   ptrUUID(patient.ID),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			fhirID = d.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.DetectedIssue
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.FHIRID != fhirID {
			t.Errorf("expected fhir_id=%s, got %s", fhirID, fetched.FHIRID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var issue *clinical.DetectedIssue
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			d := &clinical.DetectedIssue{
				Status:      "preliminary",
				CodeCode:    ptrStr("ALLRG-INT"),
				CodeDisplay: ptrStr("Allergy interaction"),
				PatientID:   ptrUUID(patient.ID),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			issue = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			issue.Status = "final"
			issue.Severity = ptrStr("moderate")
			issue.Detail = ptrStr("Mitigated by provider review")
			issue.MitigationAction = ptrStr("Prescriber notified")
			return repo.Update(ctx, issue)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *clinical.DetectedIssue
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, issue.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "final" {
			t.Errorf("expected status=final, got %s", fetched.Status)
		}
		if fetched.Severity == nil || *fetched.Severity != "moderate" {
			t.Errorf("expected severity=moderate, got %v", fetched.Severity)
		}
		if fetched.Detail == nil || *fetched.Detail != "Mitigated by provider review" {
			t.Errorf("expected detail updated, got %v", fetched.Detail)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*clinical.DetectedIssue
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "final",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 detected issue")
		}
		for _, r := range results {
			if r.Status != "final" {
				t.Errorf("expected status=final, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var issueID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			d := &clinical.DetectedIssue{
				Status:      "final",
				CodeCode:    ptrStr("DEL-TEST"),
				CodeDisplay: ptrStr("Delete test issue"),
				PatientID:   ptrUUID(patient.ID),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			issueID = d.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			return repo.Delete(ctx, issueID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewDetectedIssueRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, issueID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted detected issue")
		}
	})
}

// ---------------------------------------------------------------------------
// AdverseEvent
// ---------------------------------------------------------------------------

func TestClinicalAdverseEventCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("advevt")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "AEPatient", "Test", "MRN-AE-001")

	t.Run("Create", func(t *testing.T) {
		var created *clinical.AdverseEvent
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			a := &clinical.AdverseEvent{
				Actuality:        "actual",
				SubjectPatientID: patient.ID,
				CategoryCode:     ptrStr("medication-mishap"),
				EventCode:        ptrStr("MED-REACT"),
				Date:             ptrTime(now),
				SeriousnessCode:  ptrStr("serious"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create adverse event: %v", err)
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
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			a := &clinical.AdverseEvent{
				Actuality:        "actual",
				SubjectPatientID: uuid.New(), // non-existent
			}
			return repo.Create(ctx, a)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var aeID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			a := &clinical.AdverseEvent{
				Actuality:        "actual",
				SubjectPatientID: patient.ID,
				EventCode:        ptrStr("FALL"),
				EventDisplay:     ptrStr("Patient fall"),
				SeriousnessCode:  ptrStr("non-serious"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			aeID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.AdverseEvent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, aeID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Actuality != "actual" {
			t.Errorf("expected actuality=actual, got %s", fetched.Actuality)
		}
		if fetched.EventCode == nil || *fetched.EventCode != "FALL" {
			t.Errorf("expected event_code=FALL, got %v", fetched.EventCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			a := &clinical.AdverseEvent{
				Actuality:        "potential",
				SubjectPatientID: patient.ID,
				EventCode:        ptrStr("INFECT"),
				EventDisplay:     ptrStr("Infection"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			fhirID = a.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.AdverseEvent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.FHIRID != fhirID {
			t.Errorf("expected fhir_id=%s, got %s", fhirID, fetched.FHIRID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var ae *clinical.AdverseEvent
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			a := &clinical.AdverseEvent{
				Actuality:        "actual",
				SubjectPatientID: patient.ID,
				EventCode:        ptrStr("SKIN-RXN"),
				EventDisplay:     ptrStr("Skin reaction"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			ae = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			ae.SeriousnessCode = ptrStr("serious")
			ae.OutcomeCode = ptrStr("resolved")
			ae.Description = ptrStr("Resolved after discontinuation")
			return repo.Update(ctx, ae)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *clinical.AdverseEvent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ae.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.SeriousnessCode == nil || *fetched.SeriousnessCode != "serious" {
			t.Errorf("expected seriousness=serious, got %v", fetched.SeriousnessCode)
		}
		if fetched.OutcomeCode == nil || *fetched.OutcomeCode != "resolved" {
			t.Errorf("expected outcome=resolved, got %v", fetched.OutcomeCode)
		}
		if fetched.Description == nil || *fetched.Description != "Resolved after discontinuation" {
			t.Errorf("expected description updated, got %v", fetched.Description)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*clinical.AdverseEvent
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient":   patient.ID.String(),
				"actuality": "actual",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 adverse event")
		}
		for _, r := range results {
			if r.Actuality != "actual" {
				t.Errorf("expected actuality=actual, got %s", r.Actuality)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var aeID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			a := &clinical.AdverseEvent{
				Actuality:        "actual",
				SubjectPatientID: patient.ID,
				EventCode:        ptrStr("DEL-TEST"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			aeID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			return repo.Delete(ctx, aeID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAdverseEventRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, aeID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted adverse event")
		}
	})
}

// ---------------------------------------------------------------------------
// ClinicalImpression
// ---------------------------------------------------------------------------

func TestClinicalImpressionCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("climp")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "CIPatient", "Test", "MRN-CI-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "CIDoc", "Jones")

	t.Run("Create", func(t *testing.T) {
		var created *clinical.ClinicalImpression
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			ci := &clinical.ClinicalImpression{
				Status:           "in-progress",
				SubjectPatientID: patient.ID,
				Description:      ptrStr("Initial assessment of chest pain"),
				Date:             ptrTime(now),
				EncounterID:      nil,
				Summary:          ptrStr("Patient presents with acute chest pain"),
				AssessorID:       &practitioner.ID,
			}
			if err := repo.Create(ctx, ci); err != nil {
				return err
			}
			created = ci
			return nil
		})
		if err != nil {
			t.Fatalf("Create clinical impression: %v", err)
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
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			ci := &clinical.ClinicalImpression{
				Status:           "in-progress",
				SubjectPatientID: uuid.New(), // non-existent
				Description:      ptrStr("Test impression"),
			}
			return repo.Create(ctx, ci)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var ciID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			ci := &clinical.ClinicalImpression{
				Status:           "completed",
				SubjectPatientID: patient.ID,
				Description:      ptrStr("Follow-up assessment"),
				Summary:          ptrStr("Condition improving"),
			}
			if err := repo.Create(ctx, ci); err != nil {
				return err
			}
			ciID = ci.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.ClinicalImpression
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ciID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.Description == nil || *fetched.Description != "Follow-up assessment" {
			t.Errorf("expected description=Follow-up assessment, got %v", fetched.Description)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			ci := &clinical.ClinicalImpression{
				Status:           "in-progress",
				SubjectPatientID: patient.ID,
				Description:      ptrStr("FHIR ID lookup test"),
			}
			if err := repo.Create(ctx, ci); err != nil {
				return err
			}
			fhirID = ci.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.ClinicalImpression
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.FHIRID != fhirID {
			t.Errorf("expected fhir_id=%s, got %s", fhirID, fetched.FHIRID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var ci *clinical.ClinicalImpression
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			c := &clinical.ClinicalImpression{
				Status:           "in-progress",
				SubjectPatientID: patient.ID,
				Description:      ptrStr("Evaluation of fatigue"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			ci = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			ci.Status = "completed"
			ci.Summary = ptrStr("Iron deficiency anemia suspected")
			ci.PrognosisCode = ptrStr("good")
			ci.PrognosisDisplay = ptrStr("Good prognosis")
			ci.Note = ptrStr("Recommend iron supplementation")
			return repo.Update(ctx, ci)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *clinical.ClinicalImpression
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ci.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.Summary == nil || *fetched.Summary != "Iron deficiency anemia suspected" {
			t.Errorf("expected summary updated, got %v", fetched.Summary)
		}
		if fetched.PrognosisCode == nil || *fetched.PrognosisCode != "good" {
			t.Errorf("expected prognosis_code=good, got %v", fetched.PrognosisCode)
		}
		if fetched.Note == nil || *fetched.Note != "Recommend iron supplementation" {
			t.Errorf("expected note updated, got %v", fetched.Note)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*clinical.ClinicalImpression
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "completed",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 clinical impression")
		}
		for _, r := range results {
			if r.Status != "completed" {
				t.Errorf("expected status=completed, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var ciID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			ci := &clinical.ClinicalImpression{
				Status:           "in-progress",
				SubjectPatientID: patient.ID,
				Description:      ptrStr("Delete test impression"),
			}
			if err := repo.Create(ctx, ci); err != nil {
				return err
			}
			ciID = ci.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			return repo.Delete(ctx, ciID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewClinicalImpressionRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, ciID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted clinical impression")
		}
	})
}

// ---------------------------------------------------------------------------
// RiskAssessment
// ---------------------------------------------------------------------------

func TestRiskAssessmentCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("riskasmt")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "RAPatient", "Test", "MRN-RA-001")

	t.Run("Create", func(t *testing.T) {
		var created *clinical.RiskAssessment
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			ra := &clinical.RiskAssessment{
				Status:                "final",
				SubjectPatientID:      patient.ID,
				MethodCode:            ptrStr("MORSE"),
				CodeCode:              ptrStr("fall-risk"),
				OccurrenceDate:        ptrTime(now),
				PredictionOutcome:     ptrStr("Fall within 12 months"),
				PredictionProbability: ptrFloat(0.35),
			}
			if err := repo.Create(ctx, ra); err != nil {
				return err
			}
			created = ra
			return nil
		})
		if err != nil {
			t.Fatalf("Create risk assessment: %v", err)
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
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			ra := &clinical.RiskAssessment{
				Status:           "final",
				SubjectPatientID: uuid.New(), // non-existent
				CodeCode:         ptrStr("test-risk"),
			}
			return repo.Create(ctx, ra)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var raID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			ra := &clinical.RiskAssessment{
				Status:                "final",
				SubjectPatientID:      patient.ID,
				MethodCode:            ptrStr("BRADEN"),
				CodeCode:              ptrStr("pressure-ulcer-risk"),
				PredictionOutcome:     ptrStr("Pressure ulcer"),
				PredictionProbability: ptrFloat(0.20),
			}
			if err := repo.Create(ctx, ra); err != nil {
				return err
			}
			raID = ra.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.RiskAssessment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, raID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "final" {
			t.Errorf("expected status=final, got %s", fetched.Status)
		}
		if fetched.MethodCode == nil || *fetched.MethodCode != "BRADEN" {
			t.Errorf("expected method=BRADEN, got %v", fetched.MethodCode)
		}
		if fetched.PredictionProbability == nil || *fetched.PredictionProbability != 0.20 {
			t.Errorf("expected probability=0.20, got %v", fetched.PredictionProbability)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			ra := &clinical.RiskAssessment{
				Status:           "final",
				SubjectPatientID: patient.ID,
				CodeCode:         ptrStr("cardiac-risk"),
			}
			if err := repo.Create(ctx, ra); err != nil {
				return err
			}
			fhirID = ra.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.RiskAssessment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByFHIRID(ctx, fhirID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByFHIRID: %v", err)
		}
		if fetched.FHIRID != fhirID {
			t.Errorf("expected fhir_id=%s, got %s", fhirID, fetched.FHIRID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var ra *clinical.RiskAssessment
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			r := &clinical.RiskAssessment{
				Status:                "preliminary",
				SubjectPatientID:      patient.ID,
				MethodCode:            ptrStr("FRAMINGHAM"),
				CodeCode:              ptrStr("cv-risk"),
				PredictionOutcome:     ptrStr("Cardiovascular event"),
				PredictionProbability: ptrFloat(0.15),
			}
			if err := repo.Create(ctx, r); err != nil {
				return err
			}
			ra = r
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			ra.Status = "final"
			ra.PredictionProbability = ptrFloat(0.22)
			ra.PredictionQualitative = ptrStr("moderate")
			ra.Mitigation = ptrStr("Statin therapy initiated")
			ra.Note = ptrStr("Reassess in 6 months")
			return repo.Update(ctx, ra)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *clinical.RiskAssessment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ra.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "final" {
			t.Errorf("expected status=final, got %s", fetched.Status)
		}
		if fetched.PredictionProbability == nil || *fetched.PredictionProbability != 0.22 {
			t.Errorf("expected probability=0.22, got %v", fetched.PredictionProbability)
		}
		if fetched.PredictionQualitative == nil || *fetched.PredictionQualitative != "moderate" {
			t.Errorf("expected qualitative=moderate, got %v", fetched.PredictionQualitative)
		}
		if fetched.Mitigation == nil || *fetched.Mitigation != "Statin therapy initiated" {
			t.Errorf("expected mitigation updated, got %v", fetched.Mitigation)
		}
		if fetched.Note == nil || *fetched.Note != "Reassess in 6 months" {
			t.Errorf("expected note updated, got %v", fetched.Note)
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*clinical.RiskAssessment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "final",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 risk assessment")
		}
		for _, r := range results {
			if r.Status != "final" {
				t.Errorf("expected status=final, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var raID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			ra := &clinical.RiskAssessment{
				Status:           "final",
				SubjectPatientID: patient.ID,
				CodeCode:         ptrStr("del-test-risk"),
			}
			if err := repo.Create(ctx, ra); err != nil {
				return err
			}
			raID = ra.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			return repo.Delete(ctx, raID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewRiskAssessmentRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, raID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted risk assessment")
		}
	})
}
