package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/fhirlist"
	"github.com/google/uuid"
)

func TestFHIRListCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("flist")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ListPatient", "Test", "MRN-LIST-001")

	t.Run("Create", func(t *testing.T) {
		var created *fhirlist.FHIRList
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			l := &fhirlist.FHIRList{
				Status:           "current",
				Mode:             "working",
				Title:            ptrStr("Problem List"),
				CodeCode:         ptrStr("problems"),
				CodeDisplay:      ptrStr("Problem List"),
				SubjectPatientID: &patient.ID,
				Date:             &now,
				OrderedBy:        ptrStr("priority"),
				Note:             ptrStr("Active problem list for patient"),
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			created = l
			return nil
		})
		if err != nil {
			t.Fatalf("Create FHIR list: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		fakePatient := uuid.New()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			l := &fhirlist.FHIRList{
				Status:           "current",
				Mode:             "working",
				SubjectPatientID: &fakePatient,
			}
			return repo.Create(ctx, l)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var listID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			l := &fhirlist.FHIRList{
				Status:           "current",
				Mode:             "working",
				Title:            ptrStr("Medication List"),
				CodeCode:         ptrStr("medications"),
				CodeDisplay:      ptrStr("Medication List"),
				SubjectPatientID: &patient.ID,
				Date:             &now,
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			listID = l.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *fhirlist.FHIRList
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, listID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "current" {
			t.Errorf("expected status=current, got %s", fetched.Status)
		}
		if fetched.Mode != "working" {
			t.Errorf("expected mode=working, got %s", fetched.Mode)
		}
		if fetched.Title == nil || *fetched.Title != "Medication List" {
			t.Errorf("expected title=Medication List, got %v", fetched.Title)
		}
		if fetched.CodeCode == nil || *fetched.CodeCode != "medications" {
			t.Errorf("expected code_code=medications, got %v", fetched.CodeCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			l := &fhirlist.FHIRList{
				Status:           "current",
				Mode:             "snapshot",
				Title:            ptrStr("Allergy List"),
				SubjectPatientID: &patient.ID,
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			fhirID = l.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *fhirlist.FHIRList
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
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
		var fl *fhirlist.FHIRList
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			l := &fhirlist.FHIRList{
				Status:           "current",
				Mode:             "working",
				Title:            ptrStr("Original Title"),
				SubjectPatientID: &patient.ID,
				Date:             &now,
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			fl = l
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			fl.Status = "retired"
			fl.Title = ptrStr("Updated Title")
			fl.Note = ptrStr("List retired and replaced by new version")
			fl.OrderedBy = ptrStr("entry-date")
			fl.CodeCode = ptrStr("allergies")
			fl.CodeDisplay = ptrStr("Allergy List")
			return repo.Update(ctx, fl)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *fhirlist.FHIRList
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, fl.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "retired" {
			t.Errorf("expected status=retired, got %s", fetched.Status)
		}
		if fetched.Title == nil || *fetched.Title != "Updated Title" {
			t.Errorf("expected title=Updated Title, got %v", fetched.Title)
		}
		if fetched.Note == nil || *fetched.Note != "List retired and replaced by new version" {
			t.Errorf("expected note set, got %v", fetched.Note)
		}
		if fetched.OrderedBy == nil || *fetched.OrderedBy != "entry-date" {
			t.Errorf("expected ordered_by=entry-date, got %v", fetched.OrderedBy)
		}
		if fetched.CodeCode == nil || *fetched.CodeCode != "allergies" {
			t.Errorf("expected code_code=allergies, got %v", fetched.CodeCode)
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*fhirlist.FHIRList
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "current",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for status=current")
		}
		for _, r := range results {
			if r.Status != "current" {
				t.Errorf("expected status=current, got %s", r.Status)
			}
		}
	})

	t.Run("Search_ByCode", func(t *testing.T) {
		var results []*fhirlist.FHIRList
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"code":    "problems",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by code: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.CodeCode == nil || *r.CodeCode != "problems" {
				t.Errorf("expected code_code=problems, got %v", r.CodeCode)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var listID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			l := &fhirlist.FHIRList{
				Status:           "current",
				Mode:             "working",
				SubjectPatientID: &patient.ID,
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			listID = l.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			return repo.Delete(ctx, listID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := fhirlist.NewFHIRListRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, listID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted FHIR list")
		}
	})
}
