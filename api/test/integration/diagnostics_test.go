package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/clinical"
	"github.com/ehr/ehr/internal/domain/diagnostics"
	"github.com/google/uuid"
)

func TestServiceRequestCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("srvreq")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "SRPatient", "Test", "MRN-SR-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SRDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *diagnostics.ServiceRequest
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			sr := &diagnostics.ServiceRequest{
				PatientID:      patient.ID,
				RequesterID:    practitioner.ID,
				Status:         "active",
				Intent:         "order",
				Priority:       ptrStr("routine"),
				CategoryCode:   ptrStr("108252007"),
				CategoryDisplay: ptrStr("Laboratory procedure"),
				CodeSystem:     ptrStr("http://loinc.org"),
				CodeValue:      "2951-2",
				CodeDisplay:    "Sodium [Moles/volume] in Serum or Plasma",
				AuthoredOn:     &now,
				Note:           ptrStr("Fasting specimen preferred"),
			}
			if err := repo.Create(ctx, sr); err != nil {
				return err
			}
			created = sr
			return nil
		})
		if err != nil {
			t.Fatalf("Create service request: %v", err)
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
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			sr := &diagnostics.ServiceRequest{
				PatientID:   uuid.New(), // non-existent
				RequesterID: practitioner.ID,
				Status:      "active",
				Intent:      "order",
				CodeValue:   "test",
				CodeDisplay: "Test",
			}
			return repo.Create(ctx, sr)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var srID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			sr := &diagnostics.ServiceRequest{
				PatientID:   patient.ID,
				RequesterID: practitioner.ID,
				Status:      "active",
				Intent:      "order",
				CodeValue:   "2823-3",
				CodeDisplay: "Potassium [Moles/volume] in Serum or Plasma",
			}
			if err := repo.Create(ctx, sr); err != nil {
				return err
			}
			srID = sr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *diagnostics.ServiceRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, srID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CodeValue != "2823-3" {
			t.Errorf("expected code=2823-3, got %s", fetched.CodeValue)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			sr := &diagnostics.ServiceRequest{
				PatientID:   patient.ID,
				RequesterID: practitioner.ID,
				Status:      "active",
				Intent:      "order",
				CodeValue:   "6690-2",
				CodeDisplay: "Leukocytes",
			}
			if err := repo.Create(ctx, sr); err != nil {
				return err
			}
			fhirID = sr.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *diagnostics.ServiceRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
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
		var sr *diagnostics.ServiceRequest
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			s := &diagnostics.ServiceRequest{
				PatientID:   patient.ID,
				RequesterID: practitioner.ID,
				Status:      "active",
				Intent:      "order",
				CodeValue:   "789-8",
				CodeDisplay: "Erythrocytes",
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			sr = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			sr.Status = "completed"
			sr.Note = ptrStr("Results received")
			return repo.Update(ctx, sr)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *diagnostics.ServiceRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, sr.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.Note == nil || *fetched.Note != "Results received" {
			t.Errorf("expected note='Results received', got %v", fetched.Note)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*diagnostics.ServiceRequest
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 service request")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*diagnostics.ServiceRequest
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
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
		_ = total
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var srID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			sr := &diagnostics.ServiceRequest{
				PatientID:   patient.ID,
				RequesterID: practitioner.ID,
				Status:      "draft",
				Intent:      "order",
				CodeValue:   "del-test",
				CodeDisplay: "Delete Test",
			}
			if err := repo.Create(ctx, sr); err != nil {
				return err
			}
			srID = sr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			return repo.Delete(ctx, srID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewServiceRequestRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, srID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted service request")
		}
	})
}

func TestSpecimenCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("spec")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "SpecPatient", "Test", "MRN-SPEC-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SpecDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *diagnostics.Specimen
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			sp := &diagnostics.Specimen{
				PatientID:           patient.ID,
				Status:              "available",
				AccessionID:         ptrStr("ACC-001"),
				TypeCode:            ptrStr("119297000"),
				TypeDisplay:         ptrStr("Blood specimen"),
				CollectionCollector: &practitioner.ID,
				CollectionDatetime:  &now,
				CollectionQuantity:  ptrFloat(5.0),
				CollectionUnit:      ptrStr("mL"),
				CollectionMethod:    ptrStr("129300006"),
				CollectionBodySite:  ptrStr("368209003"),
				ContainerDesc:       ptrStr("Red-top tube"),
				ContainerType:       ptrStr("tube"),
				Note:                ptrStr("Fasting specimen"),
			}
			if err := repo.Create(ctx, sp); err != nil {
				return err
			}
			created = sp
			return nil
		})
		if err != nil {
			t.Fatalf("Create specimen: %v", err)
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
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			sp := &diagnostics.Specimen{
				PatientID: uuid.New(), // non-existent
				Status:    "available",
			}
			return repo.Create(ctx, sp)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var spID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			sp := &diagnostics.Specimen{
				PatientID:   patient.ID,
				Status:      "available",
				TypeCode:    ptrStr("122555007"),
				TypeDisplay: ptrStr("Venous blood specimen"),
			}
			if err := repo.Create(ctx, sp); err != nil {
				return err
			}
			spID = sp.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *diagnostics.Specimen
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, spID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.TypeCode == nil || *fetched.TypeCode != "122555007" {
			t.Errorf("expected type_code=122555007, got %v", fetched.TypeCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			sp := &diagnostics.Specimen{
				PatientID: patient.ID,
				Status:    "available",
				TypeCode:  ptrStr("urine-type"),
			}
			if err := repo.Create(ctx, sp); err != nil {
				return err
			}
			fhirID = sp.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *diagnostics.Specimen
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
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
		var sp *diagnostics.Specimen
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			s := &diagnostics.Specimen{
				PatientID: patient.ID,
				Status:    "available",
				TypeCode:  ptrStr("119361006"),
				TypeDisplay: ptrStr("Plasma specimen"),
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			sp = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			sp.Status = "unavailable"
			sp.ConditionCode = ptrStr("HEMOLIZED")
			sp.ConditionDisplay = ptrStr("Hemolized specimen")
			sp.Note = ptrStr("Specimen hemolized, redrawn needed")
			return repo.Update(ctx, sp)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *diagnostics.Specimen
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, sp.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "unavailable" {
			t.Errorf("expected status=unavailable, got %s", fetched.Status)
		}
		if fetched.ConditionCode == nil || *fetched.ConditionCode != "HEMOLIZED" {
			t.Errorf("expected condition_code=HEMOLIZED, got %v", fetched.ConditionCode)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*diagnostics.Specimen
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 specimen")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByType", func(t *testing.T) {
		var results []*diagnostics.Specimen
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "available",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Status != "available" {
				t.Errorf("expected status=available, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var spID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			sp := &diagnostics.Specimen{
				PatientID: patient.ID,
				Status:    "available",
			}
			if err := repo.Create(ctx, sp); err != nil {
				return err
			}
			spID = sp.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			return repo.Delete(ctx, spID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewSpecimenRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, spID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted specimen")
		}
	})
}

func TestDiagnosticReportCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("dxrpt")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "DxRptPatient", "Test", "MRN-DXRPT-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "DxRptDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *diagnostics.DiagnosticReport
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			dr := &diagnostics.DiagnosticReport{
				PatientID:         patient.ID,
				PerformerID:       &practitioner.ID,
				Status:            "final",
				CategoryCode:      ptrStr("LAB"),
				CategoryDisplay:   ptrStr("Laboratory"),
				CodeSystem:        ptrStr("http://loinc.org"),
				CodeValue:         "58410-2",
				CodeDisplay:       "Complete blood count",
				EffectiveDatetime: &now,
				Issued:            &now,
				Conclusion:        ptrStr("All values within normal range"),
			}
			if err := repo.Create(ctx, dr); err != nil {
				return err
			}
			created = dr
			return nil
		})
		if err != nil {
			t.Fatalf("Create diagnostic report: %v", err)
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
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			dr := &diagnostics.DiagnosticReport{
				PatientID:   uuid.New(), // non-existent
				Status:      "final",
				CodeValue:   "test",
				CodeDisplay: "Test",
			}
			return repo.Create(ctx, dr)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var drID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			dr := &diagnostics.DiagnosticReport{
				PatientID:   patient.ID,
				Status:      "final",
				CodeValue:   "24323-8",
				CodeDisplay: "Comprehensive metabolic panel",
			}
			if err := repo.Create(ctx, dr); err != nil {
				return err
			}
			drID = dr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *diagnostics.DiagnosticReport
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, drID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CodeValue != "24323-8" {
			t.Errorf("expected code=24323-8, got %s", fetched.CodeValue)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			dr := &diagnostics.DiagnosticReport{
				PatientID:   patient.ID,
				Status:      "preliminary",
				CodeValue:   "fhirid-test",
				CodeDisplay: "FHIR ID Test",
			}
			if err := repo.Create(ctx, dr); err != nil {
				return err
			}
			fhirID = dr.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *diagnostics.DiagnosticReport
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
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
		var dr *diagnostics.DiagnosticReport
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			d := &diagnostics.DiagnosticReport{
				PatientID:   patient.ID,
				Status:      "preliminary",
				CodeValue:   "4548-4",
				CodeDisplay: "Hemoglobin A1c",
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			dr = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			dr.Status = "final"
			dr.Conclusion = ptrStr("HbA1c 5.7% - within normal limits")
			dr.ConclusionCode = ptrStr("N")
			dr.ConclusionDisplay = ptrStr("Normal")
			dr.Note = ptrStr("Reviewed by lab director")
			return repo.Update(ctx, dr)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *diagnostics.DiagnosticReport
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, dr.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "final" {
			t.Errorf("expected status=final, got %s", fetched.Status)
		}
		if fetched.Conclusion == nil || *fetched.Conclusion != "HbA1c 5.7% - within normal limits" {
			t.Errorf("expected conclusion set, got %v", fetched.Conclusion)
		}
		if fetched.ConclusionCode == nil || *fetched.ConclusionCode != "N" {
			t.Errorf("expected conclusion_code=N, got %v", fetched.ConclusionCode)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*diagnostics.DiagnosticReport
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 diagnostic report")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByCategory", func(t *testing.T) {
		var results []*diagnostics.DiagnosticReport
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient":  patient.ID.String(),
				"category": "LAB",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by category: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.CategoryCode == nil || *r.CategoryCode != "LAB" {
				t.Errorf("expected category=LAB, got %v", r.CategoryCode)
			}
		}
	})

	t.Run("Results", func(t *testing.T) {
		// Create a diagnostic report
		var drID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			dr := &diagnostics.DiagnosticReport{
				PatientID:   patient.ID,
				Status:      "final",
				CodeValue:   "results-test",
				CodeDisplay: "Results Test Report",
			}
			if err := repo.Create(ctx, dr); err != nil {
				return err
			}
			drID = dr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create report: %v", err)
		}

		// Create an observation to link as a result
		var obsID uuid.UUID
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			obsRepo := clinical.NewObservationRepoPG(globalDB.Pool)
			obs := &clinical.Observation{
				Status:        "final",
				CodeValue:     "2951-2",
				CodeDisplay:   "Sodium",
				PatientID:     patient.ID,
				ValueQuantity: ptrFloat(140),
				ValueUnit:     ptrStr("mmol/L"),
			}
			if err := obsRepo.Create(ctx, obs); err != nil {
				return err
			}
			obsID = obs.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create observation: %v", err)
		}

		// Add result
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			return repo.AddResult(ctx, drID, obsID)
		})
		if err != nil {
			t.Fatalf("AddResult: %v", err)
		}

		// Get results
		var resultIDs []uuid.UUID
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			var err error
			resultIDs, err = repo.GetResults(ctx, drID)
			return err
		})
		if err != nil {
			t.Fatalf("GetResults: %v", err)
		}
		if len(resultIDs) != 1 {
			t.Fatalf("expected 1 result, got %d", len(resultIDs))
		}
		if resultIDs[0] != obsID {
			t.Errorf("expected observation_id=%s, got %s", obsID, resultIDs[0])
		}

		// Remove result
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			return repo.RemoveResult(ctx, drID, obsID)
		})
		if err != nil {
			t.Fatalf("RemoveResult: %v", err)
		}

		// Verify removal
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			var err error
			resultIDs, err = repo.GetResults(ctx, drID)
			return err
		})
		if err != nil {
			t.Fatalf("GetResults after remove: %v", err)
		}
		if len(resultIDs) != 0 {
			t.Errorf("expected 0 results after remove, got %d", len(resultIDs))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var drID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			dr := &diagnostics.DiagnosticReport{
				PatientID:   patient.ID,
				Status:      "final",
				CodeValue:   "del-test",
				CodeDisplay: "Delete Test",
			}
			if err := repo.Create(ctx, dr); err != nil {
				return err
			}
			drID = dr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			return repo.Delete(ctx, drID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewDiagnosticReportRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, drID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted diagnostic report")
		}
	})
}

func TestImagingStudyCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("imgstd")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ImgPatient", "Test", "MRN-IMG-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ImgDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *diagnostics.ImagingStudy
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			is := &diagnostics.ImagingStudy{
				PatientID:         patient.ID,
				ReferrerID:        &practitioner.ID,
				Status:            "available",
				ModalityCode:      ptrStr("CT"),
				ModalityDisplay:   ptrStr("Computed Tomography"),
				StudyUID:          ptrStr("1.2.840.113619.2.55.3.2831211900"),
				NumberOfSeries:    ptrInt(3),
				NumberOfInstances: ptrInt(120),
				Description:       ptrStr("CT Head without contrast"),
				Started:           &now,
				ReasonCode:        ptrStr("25064002"),
				ReasonDisplay:     ptrStr("Headache"),
				Note:              ptrStr("Urgent study"),
			}
			if err := repo.Create(ctx, is); err != nil {
				return err
			}
			created = is
			return nil
		})
		if err != nil {
			t.Fatalf("Create imaging study: %v", err)
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
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			is := &diagnostics.ImagingStudy{
				PatientID: uuid.New(), // non-existent
				Status:    "available",
			}
			return repo.Create(ctx, is)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var isID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			is := &diagnostics.ImagingStudy{
				PatientID:    patient.ID,
				Status:       "available",
				ModalityCode: ptrStr("MR"),
				ModalityDisplay: ptrStr("Magnetic Resonance"),
				Description:  ptrStr("MRI Brain"),
			}
			if err := repo.Create(ctx, is); err != nil {
				return err
			}
			isID = is.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *diagnostics.ImagingStudy
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, isID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.ModalityCode == nil || *fetched.ModalityCode != "MR" {
			t.Errorf("expected modality_code=MR, got %v", fetched.ModalityCode)
		}
		if fetched.Description == nil || *fetched.Description != "MRI Brain" {
			t.Errorf("expected description='MRI Brain', got %v", fetched.Description)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			is := &diagnostics.ImagingStudy{
				PatientID:    patient.ID,
				Status:       "available",
				ModalityCode: ptrStr("XR"),
			}
			if err := repo.Create(ctx, is); err != nil {
				return err
			}
			fhirID = is.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *diagnostics.ImagingStudy
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
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
		var is *diagnostics.ImagingStudy
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			s := &diagnostics.ImagingStudy{
				PatientID: patient.ID,
				Status:    "registered",
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			is = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			is.Status = "available"
			is.NumberOfSeries = ptrInt(2)
			is.NumberOfInstances = ptrInt(45)
			is.Description = ptrStr("Chest X-ray PA and Lateral")
			is.Note = ptrStr("Study completed successfully")
			return repo.Update(ctx, is)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *diagnostics.ImagingStudy
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, is.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "available" {
			t.Errorf("expected status=available, got %s", fetched.Status)
		}
		if fetched.NumberOfSeries == nil || *fetched.NumberOfSeries != 2 {
			t.Errorf("expected number_of_series=2, got %v", fetched.NumberOfSeries)
		}
		if fetched.NumberOfInstances == nil || *fetched.NumberOfInstances != 45 {
			t.Errorf("expected number_of_instances=45, got %v", fetched.NumberOfInstances)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*diagnostics.ImagingStudy
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 imaging study")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByModality", func(t *testing.T) {
		var results []*diagnostics.ImagingStudy
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient":  patient.ID.String(),
				"modality": "CT",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by modality: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.ModalityCode == nil || *r.ModalityCode != "CT" {
				t.Errorf("expected modality=CT, got %v", r.ModalityCode)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var isID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			is := &diagnostics.ImagingStudy{
				PatientID: patient.ID,
				Status:    "available",
			}
			if err := repo.Create(ctx, is); err != nil {
				return err
			}
			isID = is.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			return repo.Delete(ctx, isID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := diagnostics.NewImagingStudyRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, isID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted imaging study")
		}
	})
}
