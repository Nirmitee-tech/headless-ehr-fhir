package integration

import (
	"context"
	"testing"

	"github.com/ehr/ehr/internal/domain/healthcareservice"
	"github.com/google/uuid"
)

func TestHealthcareServiceCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("hcsvc")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	orgID := createTestOrganization(t, ctx, globalDB.Pool, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *healthcareservice.HealthcareService
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			hs := &healthcareservice.HealthcareService{
				Active:              true,
				Name:                "Cardiology Clinic",
				CategoryCode:        ptrStr("35"),
				CategoryDisplay:     ptrStr("Hospital"),
				TypeCode:            ptrStr("394579002"),
				TypeDisplay:         ptrStr("Cardiology"),
				Comment:             ptrStr("Provides comprehensive cardiac care"),
				TelecomPhone:        ptrStr("555-0100"),
				AppointmentRequired: true,
				ProvidedByOrgID:     &orgID,
			}
			if err := repo.Create(ctx, hs); err != nil {
				return err
			}
			created = hs
			return nil
		})
		if err != nil {
			t.Fatalf("Create healthcare service: %v", err)
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
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			fakeOrg := uuid.New()
			hs := &healthcareservice.HealthcareService{
				Active:          true,
				Name:            "Fake Org Service",
				ProvidedByOrgID: &fakeOrg,
			}
			return repo.Create(ctx, hs)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent organization")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var hsID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			hs := &healthcareservice.HealthcareService{
				Active:       true,
				Name:         "Dermatology Clinic",
				TypeCode:     ptrStr("394582007"),
				TypeDisplay:  ptrStr("Dermatology"),
				TelecomPhone: ptrStr("555-0200"),
			}
			if err := repo.Create(ctx, hs); err != nil {
				return err
			}
			hsID = hs.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *healthcareservice.HealthcareService
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, hsID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "Dermatology Clinic" {
			t.Errorf("expected name=Dermatology Clinic, got %s", fetched.Name)
		}
		if fetched.TypeCode == nil || *fetched.TypeCode != "394582007" {
			t.Errorf("expected type_code=394582007, got %v", fetched.TypeCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			hs := &healthcareservice.HealthcareService{
				Active: true,
				Name:   "Radiology Services",
			}
			if err := repo.Create(ctx, hs); err != nil {
				return err
			}
			fhirID = hs.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *healthcareservice.HealthcareService
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
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
		var hs *healthcareservice.HealthcareService
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			s := &healthcareservice.HealthcareService{
				Active:              true,
				Name:                "General Practice",
				Comment:             ptrStr("Original comment"),
				AppointmentRequired: false,
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			hs = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			hs.Active = false
			hs.Comment = ptrStr("Service temporarily unavailable")
			hs.AppointmentRequired = true
			hs.CategoryCode = ptrStr("17")
			hs.CategoryDisplay = ptrStr("General Practice")
			return repo.Update(ctx, hs)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *healthcareservice.HealthcareService
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, hs.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Active != false {
			t.Errorf("expected active=false, got %v", fetched.Active)
		}
		if fetched.Comment == nil || *fetched.Comment != "Service temporarily unavailable" {
			t.Errorf("expected updated comment, got %v", fetched.Comment)
		}
		if fetched.AppointmentRequired != true {
			t.Errorf("expected appointment_required=true, got %v", fetched.AppointmentRequired)
		}
	})

	t.Run("Search_ByName", func(t *testing.T) {
		var results []*healthcareservice.HealthcareService
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"name": "Cardiology",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by name: %v", err)
		}
		_ = total
		_ = results
	})

	t.Run("Search_ByActive", func(t *testing.T) {
		var results []*healthcareservice.HealthcareService
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"active": "true",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by active: %v", err)
		}
		_ = total
		for _, r := range results {
			if !r.Active {
				t.Errorf("expected active=true, got false")
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var hsID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			hs := &healthcareservice.HealthcareService{
				Active: true,
				Name:   "Delete Test Service",
			}
			if err := repo.Create(ctx, hs); err != nil {
				return err
			}
			hsID = hs.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			return repo.Delete(ctx, hsID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := healthcareservice.NewHealthcareServiceRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, hsID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted healthcare service")
		}
	})
}
