package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/measurereport"
	"github.com/google/uuid"
)

func TestMeasureReportCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("mrpt")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "MRPatient", "Test", "MRN-MR-001")

	t.Run("Create", func(t *testing.T) {
		var created *measurereport.MeasureReport
		now := time.Now()
		periodStart := now.Add(-30 * 24 * time.Hour)
		periodEnd := now
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			mr := &measurereport.MeasureReport{
				Status:              "complete",
				Type:                "individual",
				MeasureURL:          ptrStr("http://example.org/fhir/Measure/diabetes-screening"),
				SubjectPatientID:    &patient.ID,
				Date:                &now,
				PeriodStart:         periodStart,
				PeriodEnd:           periodEnd,
				GroupCode:           ptrStr("diabetes-group"),
				GroupPopulationCode: ptrStr("initial-population"),
				GroupPopulationCount: ptrInt(1),
				GroupMeasureScore:   ptrFloat(0.85),
			}
			if err := repo.Create(ctx, mr); err != nil {
				return err
			}
			created = mr
			return nil
		})
		if err != nil {
			t.Fatalf("Create measure report: %v", err)
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
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			mr := &measurereport.MeasureReport{
				Status:           "complete",
				Type:             "individual",
				SubjectPatientID: &fakePatient,
				PeriodStart:      now.Add(-24 * time.Hour),
				PeriodEnd:        now,
			}
			return repo.Create(ctx, mr)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var mrID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			mr := &measurereport.MeasureReport{
				Status:           "complete",
				Type:             "individual",
				SubjectPatientID: &patient.ID,
				MeasureURL:       ptrStr("http://example.org/fhir/Measure/bp-control"),
				PeriodStart:      now.Add(-7 * 24 * time.Hour),
				PeriodEnd:        now,
				GroupCode:        ptrStr("bp-group"),
			}
			if err := repo.Create(ctx, mr); err != nil {
				return err
			}
			mrID = mr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *measurereport.MeasureReport
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, mrID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "complete" {
			t.Errorf("expected status=complete, got %s", fetched.Status)
		}
		if fetched.Type != "individual" {
			t.Errorf("expected type=individual, got %s", fetched.Type)
		}
		if fetched.MeasureURL == nil || *fetched.MeasureURL != "http://example.org/fhir/Measure/bp-control" {
			t.Errorf("expected measure_url set, got %v", fetched.MeasureURL)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			mr := &measurereport.MeasureReport{
				Status:           "complete",
				Type:             "summary",
				SubjectPatientID: &patient.ID,
				PeriodStart:      now.Add(-14 * 24 * time.Hour),
				PeriodEnd:        now,
			}
			if err := repo.Create(ctx, mr); err != nil {
				return err
			}
			fhirID = mr.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *measurereport.MeasureReport
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
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
		var mr *measurereport.MeasureReport
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			m := &measurereport.MeasureReport{
				Status:           "complete",
				Type:             "individual",
				SubjectPatientID: &patient.ID,
				PeriodStart:      now.Add(-30 * 24 * time.Hour),
				PeriodEnd:        now,
				GroupCode:        ptrStr("quality-group"),
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			mr = m
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			mr.GroupPopulationCode = ptrStr("denominator")
			mr.GroupPopulationCount = ptrInt(50)
			mr.GroupMeasureScore = ptrFloat(0.92)
			mr.MeasureURL = ptrStr("http://example.org/fhir/Measure/quality-measure")
			return repo.Update(ctx, mr)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *measurereport.MeasureReport
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, mr.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.GroupPopulationCode == nil || *fetched.GroupPopulationCode != "denominator" {
			t.Errorf("expected group_population_code=denominator, got %v", fetched.GroupPopulationCode)
		}
		if fetched.GroupPopulationCount == nil || *fetched.GroupPopulationCount != 50 {
			t.Errorf("expected group_population_count=50, got %v", fetched.GroupPopulationCount)
		}
		if fetched.GroupMeasureScore == nil || *fetched.GroupMeasureScore != 0.92 {
			t.Errorf("expected group_measure_score=0.92, got %v", fetched.GroupMeasureScore)
		}
		if fetched.MeasureURL == nil || *fetched.MeasureURL != "http://example.org/fhir/Measure/quality-measure" {
			t.Errorf("expected updated measure_url, got %v", fetched.MeasureURL)
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*measurereport.MeasureReport
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "complete",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for status=complete")
		}
		for _, r := range results {
			if r.Status != "complete" {
				t.Errorf("expected status=complete, got %s", r.Status)
			}
		}
	})

	t.Run("Search_ByType", func(t *testing.T) {
		var results []*measurereport.MeasureReport
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"type":    "individual",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by type: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Type != "individual" {
				t.Errorf("expected type=individual, got %s", r.Type)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var mrID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			mr := &measurereport.MeasureReport{
				Status:           "complete",
				Type:             "individual",
				SubjectPatientID: &patient.ID,
				PeriodStart:      now.Add(-24 * time.Hour),
				PeriodEnd:        now,
			}
			if err := repo.Create(ctx, mr); err != nil {
				return err
			}
			mrID = mr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			return repo.Delete(ctx, mrID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := measurereport.NewMeasureReportRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, mrID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted measure report")
		}
	})
}
