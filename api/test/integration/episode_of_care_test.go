package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/episodeofcare"
	"github.com/google/uuid"
)

func TestEpisodeOfCareCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("eoc")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "EocPatient", "Test", "MRN-EOC-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "EocDoc", "Smith")
	orgID := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *episodeofcare.EpisodeOfCare
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			e := &episodeofcare.EpisodeOfCare{
				Status:        "active",
				TypeCode:      ptrStr("hacc"),
				TypeDisplay:   ptrStr("Home and Community Care"),
				PatientID:     patient.ID,
				ManagingOrgID: ptrUUID(orgID),
				PeriodStart:   ptrTime(now),
				CareManagerID: ptrUUID(practitioner.ID),
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			created = e
			return nil
		})
		if err != nil {
			t.Fatalf("Create episode of care: %v", err)
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
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			fakePatient := uuid.New()
			e := &episodeofcare.EpisodeOfCare{
				Status:    "active",
				PatientID: fakePatient,
			}
			return repo.Create(ctx, e)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var eocID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			e := &episodeofcare.EpisodeOfCare{
				Status:        "active",
				TypeCode:      ptrStr("hacc"),
				TypeDisplay:   ptrStr("Home and Community Care"),
				PatientID:     patient.ID,
				ManagingOrgID: ptrUUID(orgID),
				PeriodStart:   ptrTime(now),
				CareManagerID: ptrUUID(practitioner.ID),
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			eocID = e.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *episodeofcare.EpisodeOfCare
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, eocID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.TypeCode == nil || *fetched.TypeCode != "hacc" {
			t.Errorf("expected type_code=hacc, got %v", fetched.TypeCode)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
		if fetched.ManagingOrgID == nil || *fetched.ManagingOrgID != orgID {
			t.Errorf("expected managing_org_id=%s, got %v", orgID, fetched.ManagingOrgID)
		}
		if fetched.CareManagerID == nil || *fetched.CareManagerID != practitioner.ID {
			t.Errorf("expected care_manager_id=%s, got %v", practitioner.ID, fetched.CareManagerID)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			e := &episodeofcare.EpisodeOfCare{
				Status:    "active",
				PatientID: patient.ID,
				TypeCode:  ptrStr("cacp"),
				TypeDisplay: ptrStr("Community Aged Care Packages"),
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			fhirID = e.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *episodeofcare.EpisodeOfCare
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
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
		if fetched.TypeCode == nil || *fetched.TypeCode != "cacp" {
			t.Errorf("expected type_code=cacp, got %v", fetched.TypeCode)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var eoc *episodeofcare.EpisodeOfCare
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			e := &episodeofcare.EpisodeOfCare{
				Status:      "planned",
				TypeCode:    ptrStr("hacc"),
				TypeDisplay: ptrStr("Home and Community Care"),
				PatientID:   patient.ID,
				PeriodStart: ptrTime(now),
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			eoc = e
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		// Update status to active and add care manager
		endTime := now.Add(30 * 24 * time.Hour)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			eoc.Status = "active"
			eoc.CareManagerID = ptrUUID(practitioner.ID)
			eoc.ManagingOrgID = ptrUUID(orgID)
			eoc.PeriodEnd = ptrTime(endTime)
			return repo.Update(ctx, eoc)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *episodeofcare.EpisodeOfCare
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, eoc.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.CareManagerID == nil || *fetched.CareManagerID != practitioner.ID {
			t.Errorf("expected care_manager_id=%s, got %v", practitioner.ID, fetched.CareManagerID)
		}
		if fetched.ManagingOrgID == nil || *fetched.ManagingOrgID != orgID {
			t.Errorf("expected managing_org_id=%s, got %v", orgID, fetched.ManagingOrgID)
		}
		if fetched.PeriodEnd == nil {
			t.Error("expected non-nil PeriodEnd")
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		// Create an episode with a specific status for searching
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			e := &episodeofcare.EpisodeOfCare{
				Status:    "waitlist",
				PatientID: patient.ID,
				TypeCode:  ptrStr("da"),
				TypeDisplay: ptrStr("Drug and Alcohol"),
			}
			return repo.Create(ctx, e)
		})
		if err != nil {
			t.Fatalf("Create for search: %v", err)
		}

		var results []*episodeofcare.EpisodeOfCare
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "waitlist",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for status=waitlist")
		}
		for _, r := range results {
			if r.Status != "waitlist" {
				t.Errorf("expected status=waitlist, got %s", r.Status)
			}
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByPatient", func(t *testing.T) {
		var results []*episodeofcare.EpisodeOfCare
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
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
			t.Error("expected at least 1 episode for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var eocID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			e := &episodeofcare.EpisodeOfCare{
				Status:    "active",
				PatientID: patient.ID,
				TypeCode:  ptrStr("hacc"),
				TypeDisplay: ptrStr("Home and Community Care"),
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			eocID = e.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			return repo.Delete(ctx, eocID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := episodeofcare.NewEpisodeOfCareRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, eocID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted episode of care")
		}
	})
}
