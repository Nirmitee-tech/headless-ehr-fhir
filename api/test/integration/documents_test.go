package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/documents"
	"github.com/google/uuid"
)

func TestConsentCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("consent")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ConsentPatient", "Test", "MRN-CONSENT-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ConsentDoc", "Smith")
	org := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *documents.Consent
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			c := &documents.Consent{
				Status:          "active",
				Scope:           ptrStr("patient-privacy"),
				CategoryCode:    ptrStr("59284-0"),
				CategoryDisplay: ptrStr("Consent Document"),
				PatientID:       patient.ID,
				PerformerID:     &practitioner.ID,
				OrganizationID:  &org,
				PolicyAuthority: ptrStr("http://example.org"),
				PolicyURI:       ptrStr("http://example.org/policy"),
				ProvisionType:   ptrStr("permit"),
				ProvisionStart:  &now,
				ProvisionAction: ptrStr("access"),
				HIPAAAuth:       ptrBool(true),
				DateTime:        &now,
				Note:            ptrStr("Patient consented to data sharing"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			created = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create consent: %v", err)
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
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			c := &documents.Consent{
				Status:    "active",
				PatientID: uuid.New(),
			}
			return repo.Create(ctx, c)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var consentID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			c := &documents.Consent{
				Status:    "active",
				Scope:     ptrStr("patient-privacy"),
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			consentID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *documents.Consent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, consentID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			c := &documents.Consent{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			fhirID = c.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *documents.Consent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
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
		var consent *documents.Consent
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			c := &documents.Consent{
				Status:    "active",
				PatientID: patient.ID,
				Note:      ptrStr("Original note"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			consent = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			consent.Status = "inactive"
			consent.Note = ptrStr("Consent revoked")
			consent.HIPAAAuth = ptrBool(false)
			return repo.Update(ctx, consent)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *documents.Consent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, consent.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "inactive" {
			t.Errorf("expected status=inactive, got %s", fetched.Status)
		}
		if fetched.Note == nil || *fetched.Note != "Consent revoked" {
			t.Errorf("expected note='Consent revoked', got %v", fetched.Note)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		// Create another consent
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			c := &documents.Consent{
				Status:    "active",
				PatientID: patient.ID,
			}
			return repo.Create(ctx, c)
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var results []*documents.Consent
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 consent")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*documents.Consent
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
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
			t.Error("expected at least 1 active consent")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var consentID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			c := &documents.Consent{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			consentID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			return repo.Delete(ctx, consentID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewConsentRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, consentID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted consent")
		}
	})
}

func TestDocumentReferenceCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("docref")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "DocRefPatient", "Test", "MRN-DOCREF-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "DocRefDoc", "Smith")
	org := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *documents.DocumentReference
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			d := &documents.DocumentReference{
				Status:          "current",
				DocStatus:       ptrStr("final"),
				TypeCode:        ptrStr("34133-9"),
				TypeDisplay:     ptrStr("Summarization of Episode Note"),
				CategoryCode:    ptrStr("clinical-note"),
				CategoryDisplay: ptrStr("Clinical Note"),
				PatientID:       patient.ID,
				AuthorID:        &practitioner.ID,
				CustodianID:     &org,
				Date:            &now,
				Description:     ptrStr("Annual physical exam summary"),
				SecurityLabel:   ptrStr("N"),
				ContentType:     ptrStr("application/pdf"),
				ContentURL:      ptrStr("https://example.com/doc/123.pdf"),
				ContentSize:     ptrInt(45678),
				ContentHash:     ptrStr("abc123hash"),
				ContentTitle:    ptrStr("Physical Exam Report"),
				FormatCode:      ptrStr("urn:ihe:iti:xds:2017:mimeTypeSufficient"),
				FormatDisplay:   ptrStr("mimeType Sufficient"),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			created = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create document reference: %v", err)
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
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			d := &documents.DocumentReference{
				Status:    "current",
				PatientID: uuid.New(),
			}
			return repo.Create(ctx, d)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var docID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			d := &documents.DocumentReference{
				Status:      "current",
				PatientID:   patient.ID,
				ContentType: ptrStr("text/plain"),
				ContentURL:  ptrStr("https://example.com/doc/text.txt"),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			docID = d.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *documents.DocumentReference
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, docID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "current" {
			t.Errorf("expected status=current, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			d := &documents.DocumentReference{
				Status:    "current",
				PatientID: patient.ID,
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

		var fetched *documents.DocumentReference
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
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
		var doc *documents.DocumentReference
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			d := &documents.DocumentReference{
				Status:      "current",
				PatientID:   patient.ID,
				Description: ptrStr("Original description"),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			doc = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			doc.Status = "superseded"
			doc.Description = ptrStr("Replaced by newer version")
			doc.SecurityLabel = ptrStr("R")
			return repo.Update(ctx, doc)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *documents.DocumentReference
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, doc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "superseded" {
			t.Errorf("expected status=superseded, got %s", fetched.Status)
		}
		if fetched.Description == nil || *fetched.Description != "Replaced by newer version" {
			t.Errorf("expected updated description, got %v", fetched.Description)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*documents.DocumentReference
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 document reference")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*documents.DocumentReference
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "current",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Status != "current" {
				t.Errorf("expected status=current, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var docID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			d := &documents.DocumentReference{
				Status:    "current",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			docID = d.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			return repo.Delete(ctx, docID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewDocumentReferenceRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, docID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted document reference")
		}
	})
}

func TestClinicalNoteCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("clinnote")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "NotePatient", "Test", "MRN-NOTE-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "NoteDoc", "Smith")
	enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

	t.Run("Create", func(t *testing.T) {
		var created *documents.ClinicalNote
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			n := &documents.ClinicalNote{
				PatientID:   patient.ID,
				EncounterID: &enc.ID,
				AuthorID:    practitioner.ID,
				NoteType:    "progress",
				Status:      "in-progress",
				Title:       ptrStr("Progress Note"),
				Subjective:  ptrStr("Patient reports feeling better"),
				Objective:   ptrStr("Vitals stable"),
				Assessment:  ptrStr("Improving condition"),
				Plan:        ptrStr("Continue current treatment"),
				NoteText:    ptrStr("Full text of progress note"),
			}
			if err := repo.Create(ctx, n); err != nil {
				return err
			}
			created = n
			return nil
		})
		if err != nil {
			t.Fatalf("Create clinical note: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			n := &documents.ClinicalNote{
				PatientID: uuid.New(),
				AuthorID:  practitioner.ID,
				NoteType:  "progress",
				Status:    "in-progress",
			}
			return repo.Create(ctx, n)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var noteID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			n := &documents.ClinicalNote{
				PatientID:   patient.ID,
				EncounterID: &enc.ID,
				AuthorID:    practitioner.ID,
				NoteType:    "discharge",
				Status:      "final",
				Title:       ptrStr("Discharge Summary"),
				NoteText:    ptrStr("Patient discharged in good condition"),
			}
			if err := repo.Create(ctx, n); err != nil {
				return err
			}
			noteID = n.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *documents.ClinicalNote
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, noteID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.NoteType != "discharge" {
			t.Errorf("expected note_type=discharge, got %s", fetched.NoteType)
		}
		if fetched.Status != "final" {
			t.Errorf("expected status=final, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var note *documents.ClinicalNote
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			n := &documents.ClinicalNote{
				PatientID: patient.ID,
				AuthorID:  practitioner.ID,
				NoteType:  "progress",
				Status:    "in-progress",
				Title:     ptrStr("Draft Note"),
			}
			if err := repo.Create(ctx, n); err != nil {
				return err
			}
			note = n
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		now := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			note.Status = "final"
			note.SignedBy = &practitioner.ID
			note.SignedAt = &now
			note.NoteText = ptrStr("Finalized note text")
			return repo.Update(ctx, note)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *documents.ClinicalNote
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, note.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "final" {
			t.Errorf("expected status=final, got %s", fetched.Status)
		}
		if fetched.SignedBy == nil || *fetched.SignedBy != practitioner.ID {
			t.Errorf("expected signed_by=%s, got %v", practitioner.ID, fetched.SignedBy)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*documents.ClinicalNote
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 clinical note")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("ListByEncounter", func(t *testing.T) {
		var results []*documents.ClinicalNote
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByEncounter(ctx, enc.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByEncounter: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 clinical note for encounter")
		}
		for _, r := range results {
			if r.EncounterID == nil || *r.EncounterID != enc.ID {
				t.Errorf("expected encounter_id=%s, got %v", enc.ID, r.EncounterID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var noteID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			n := &documents.ClinicalNote{
				PatientID: patient.ID,
				AuthorID:  practitioner.ID,
				NoteType:  "progress",
				Status:    "in-progress",
			}
			if err := repo.Create(ctx, n); err != nil {
				return err
			}
			noteID = n.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			return repo.Delete(ctx, noteID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewClinicalNoteRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, noteID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted clinical note")
		}
	})
}

func TestCompositionCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("comp")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "CompPatient", "Test", "MRN-COMP-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "CompDoc", "Smith")
	org := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *documents.Composition
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			c := &documents.Composition{
				Status:          "final",
				TypeCode:        ptrStr("11488-4"),
				TypeDisplay:     ptrStr("Consult note"),
				CategoryCode:    ptrStr("LP173421-1"),
				CategoryDisplay: ptrStr("Report"),
				PatientID:       patient.ID,
				Date:            &now,
				AuthorID:        &practitioner.ID,
				Title:           ptrStr("Consultation Note"),
				Confidentiality: ptrStr("N"),
				CustodianID:     &org,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			created = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create composition: %v", err)
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
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			c := &documents.Composition{
				Status:    "final",
				PatientID: uuid.New(),
			}
			return repo.Create(ctx, c)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var compID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			c := &documents.Composition{
				Status:    "preliminary",
				PatientID: patient.ID,
				Title:     ptrStr("Test Composition"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			compID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *documents.Composition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, compID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "preliminary" {
			t.Errorf("expected status=preliminary, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			c := &documents.Composition{
				Status:    "final",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			fhirID = c.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *documents.Composition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
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
		var comp *documents.Composition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			c := &documents.Composition{
				Status:    "preliminary",
				PatientID: patient.ID,
				Title:     ptrStr("Draft Composition"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			comp = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			comp.Status = "final"
			comp.Title = ptrStr("Final Composition")
			comp.Confidentiality = ptrStr("R")
			return repo.Update(ctx, comp)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *documents.Composition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, comp.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "final" {
			t.Errorf("expected status=final, got %s", fetched.Status)
		}
		if fetched.Title == nil || *fetched.Title != "Final Composition" {
			t.Errorf("expected title='Final Composition', got %v", fetched.Title)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*documents.Composition
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 composition")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Sections", func(t *testing.T) {
		var compID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			c := &documents.Composition{
				Status:    "final",
				PatientID: patient.ID,
				Title:     ptrStr("Composition with Sections"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			compID = c.ID

			// Add sections
			s1 := &documents.CompositionSection{
				CompositionID:  compID,
				Title:          ptrStr("Chief Complaint"),
				CodeValue:      ptrStr("10154-3"),
				CodeDisplay:    ptrStr("Chief complaint"),
				TextStatus:     ptrStr("generated"),
				TextDiv:        ptrStr("<div>Headache for 3 days</div>"),
				Mode:           ptrStr("snapshot"),
				SortOrder:      ptrInt(1),
			}
			if err := repo.AddSection(ctx, s1); err != nil {
				return err
			}

			s2 := &documents.CompositionSection{
				CompositionID:  compID,
				Title:          ptrStr("History of Present Illness"),
				CodeValue:      ptrStr("10164-2"),
				CodeDisplay:    ptrStr("History of present illness"),
				TextStatus:     ptrStr("generated"),
				TextDiv:        ptrStr("<div>Patient describes intermittent headache</div>"),
				SortOrder:      ptrInt(2),
			}
			return repo.AddSection(ctx, s2)
		})
		if err != nil {
			t.Fatalf("Create composition with sections: %v", err)
		}

		// Verify sections
		var sections []*documents.CompositionSection
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			var err error
			sections, err = repo.GetSections(ctx, compID)
			return err
		})
		if err != nil {
			t.Fatalf("GetSections: %v", err)
		}
		if len(sections) != 2 {
			t.Fatalf("expected 2 sections, got %d", len(sections))
		}
		if sections[0].Title == nil || *sections[0].Title != "Chief Complaint" {
			t.Errorf("expected first section title='Chief Complaint', got %v", sections[0].Title)
		}
		if sections[1].Title == nil || *sections[1].Title != "History of Present Illness" {
			t.Errorf("expected second section title='History of Present Illness', got %v", sections[1].Title)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var compID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			c := &documents.Composition{
				Status:    "final",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			compID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			return repo.Delete(ctx, compID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := documents.NewCompositionRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, compID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted composition")
		}
	})
}
