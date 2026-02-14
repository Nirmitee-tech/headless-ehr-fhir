package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/oncology"
	"github.com/google/uuid"
)

func TestCancerDiagnosisCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("oncdx")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "OncPatient", "Test", "MRN-ONC-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "OncDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *oncology.CancerDiagnosis
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewCancerDiagnosisRepoPG(globalDB.Pool)
			d := &oncology.CancerDiagnosis{
				PatientID:            patient.ID,
				DiagnosisDate:        time.Now(),
				CancerType:           ptrStr("breast"),
				CancerSite:           ptrStr("C50.9"),
				HistologyCode:        ptrStr("8500/3"),
				HistologyDisplay:     ptrStr("Invasive ductal carcinoma"),
				StagingSystem:        ptrStr("AJCC 8th"),
				StageGroup:           ptrStr("IIA"),
				TStage:               ptrStr("T2"),
				NStage:               ptrStr("N0"),
				MStage:               ptrStr("M0"),
				Grade:                ptrStr("2"),
				Laterality:           ptrStr("left"),
				CurrentStatus:        "active",
				DiagnosingProviderID: &practitioner.ID,
				ManagingProviderID:   &practitioner.ID,
				ICD10Code:            ptrStr("C50.912"),
				ICD10Display:         ptrStr("Malignant neoplasm of unspecified site of left female breast"),
				Note:                 ptrStr("Initial diagnosis"),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			created = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create cancer diagnosis: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		dx := createTestCancerDiagnosis(t, ctx, tenantID, patient.ID)

		var fetched *oncology.CancerDiagnosis
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewCancerDiagnosisRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, dx.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
		if fetched.CurrentStatus != "active" {
			t.Errorf("expected status=active, got %s", fetched.CurrentStatus)
		}
	})

	t.Run("Update", func(t *testing.T) {
		dx := createTestCancerDiagnosis(t, ctx, tenantID, patient.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewCancerDiagnosisRepoPG(globalDB.Pool)
			dx.CurrentStatus = "remission"
			dx.StageGroup = ptrStr("IIB")
			dx.Note = ptrStr("Responding well to treatment")
			return repo.Update(ctx, dx)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *oncology.CancerDiagnosis
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewCancerDiagnosisRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, dx.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.CurrentStatus != "remission" {
			t.Errorf("expected status=remission, got %s", fetched.CurrentStatus)
		}
		if fetched.StageGroup == nil || *fetched.StageGroup != "IIB" {
			t.Errorf("expected stage_group=IIB, got %v", fetched.StageGroup)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		dx := createTestCancerDiagnosis(t, ctx, tenantID, patient.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewCancerDiagnosisRepoPG(globalDB.Pool)
			return repo.Delete(ctx, dx.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewCancerDiagnosisRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, dx.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted cancer diagnosis")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*oncology.CancerDiagnosis
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewCancerDiagnosisRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 cancer diagnosis in list")
		}
		if len(results) != total {
			t.Errorf("expected results count=%d to match total=%d", len(results), total)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		createTestCancerDiagnosis(t, ctx, tenantID, patient.ID)

		var results []*oncology.CancerDiagnosis
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewCancerDiagnosisRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 diagnosis for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})
}

func TestTreatmentProtocolCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("oncproto")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ProtoPatient", "Test", "MRN-PROTO-001")
	dx := createTestCancerDiagnosis(t, ctx, tenantID, patient.ID)

	t.Run("Create", func(t *testing.T) {
		var created *oncology.TreatmentProtocol
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
			p := &oncology.TreatmentProtocol{
				CancerDiagnosisID: dx.ID,
				ProtocolName:      "AC-T",
				ProtocolCode:      ptrStr("AC-T-001"),
				ProtocolType:      ptrStr("chemotherapy"),
				Intent:            ptrStr("curative"),
				NumberOfCycles:    ptrInt(8),
				CycleLengthDays:   ptrInt(21),
				StartDate:         &now,
				Status:            "active",
				Note:              ptrStr("Adjuvant chemotherapy"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			created = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create treatment protocol: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		proto := createTestTreatmentProtocol(t, ctx, tenantID, dx.ID)

		var fetched *oncology.TreatmentProtocol
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, proto.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CancerDiagnosisID != dx.ID {
			t.Errorf("expected cancer_diagnosis_id=%s, got %s", dx.ID, fetched.CancerDiagnosisID)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		proto := createTestTreatmentProtocol(t, ctx, tenantID, dx.ID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
			proto.Status = "completed"
			proto.EndDate = &now
			proto.Note = ptrStr("Protocol completed successfully")
			return repo.Update(ctx, proto)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *oncology.TreatmentProtocol
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, proto.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.EndDate == nil {
			t.Error("expected non-nil EndDate")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		proto := createTestTreatmentProtocol(t, ctx, tenantID, dx.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
			return repo.Delete(ctx, proto.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, proto.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted protocol")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*oncology.TreatmentProtocol
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 protocol")
		}
		_ = results
	})

	t.Run("AddDrug_and_GetDrugs", func(t *testing.T) {
		proto := createTestTreatmentProtocol(t, ctx, tenantID, dx.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
			drug := &oncology.TreatmentProtocolDrug{
				ProtocolID:            proto.ID,
				DrugName:              "Doxorubicin",
				DrugCode:              ptrStr("3639"),
				DrugCodeSystem:        ptrStr("http://www.nlm.nih.gov/research/umls/rxnorm"),
				Route:                 ptrStr("IV"),
				DoseValue:             ptrFloat(60),
				DoseUnit:              ptrStr("mg/m2"),
				DoseCalculationMethod: ptrStr("BSA"),
				Frequency:             ptrStr("every 21 days"),
				AdministrationDay:     ptrStr("Day 1"),
				InfusionDurationMin:   ptrInt(30),
				SequenceOrder:         ptrInt(1),
				Note:                  ptrStr("Administer with antiemetics"),
			}
			return repo.AddDrug(ctx, drug)
		})
		if err != nil {
			t.Fatalf("AddDrug: %v", err)
		}

		// Add second drug
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
			drug := &oncology.TreatmentProtocolDrug{
				ProtocolID:    proto.ID,
				DrugName:      "Cyclophosphamide",
				Route:         ptrStr("IV"),
				DoseValue:     ptrFloat(600),
				DoseUnit:      ptrStr("mg/m2"),
				SequenceOrder: ptrInt(2),
			}
			return repo.AddDrug(ctx, drug)
		})
		if err != nil {
			t.Fatalf("AddDrug (second): %v", err)
		}

		var drugs []*oncology.TreatmentProtocolDrug
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
			var err error
			drugs, err = repo.GetDrugs(ctx, proto.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetDrugs: %v", err)
		}
		if len(drugs) != 2 {
			t.Fatalf("expected 2 drugs, got %d", len(drugs))
		}
		if drugs[0].DrugName != "Doxorubicin" {
			t.Errorf("expected first drug=Doxorubicin, got %s", drugs[0].DrugName)
		}
	})
}

func TestChemoCycleCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("oncchemo")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ChemoPatient", "Test", "MRN-CHEMO-001")
	dx := createTestCancerDiagnosis(t, ctx, tenantID, patient.ID)
	proto := createTestTreatmentProtocol(t, ctx, tenantID, dx.ID)

	t.Run("Create", func(t *testing.T) {
		var created *oncology.ChemoCycle
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewChemoCycleRepoPG(globalDB.Pool)
			c := &oncology.ChemoCycle{
				ProtocolID:       proto.ID,
				CycleNumber:      1,
				PlannedStartDate: &now,
				Status:           "planned",
				BSAM2:            ptrFloat(1.8),
				WeightKG:         ptrFloat(70),
				HeightCM:         ptrFloat(170),
				Note:             ptrStr("First cycle"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			created = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create chemo cycle: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		cycle := createTestChemoCycle(t, ctx, tenantID, proto.ID, 2)

		var fetched *oncology.ChemoCycle
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewChemoCycleRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, cycle.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CycleNumber != 2 {
			t.Errorf("expected cycle_number=2, got %d", fetched.CycleNumber)
		}
		if fetched.ProtocolID != proto.ID {
			t.Errorf("expected protocol_id=%s, got %s", proto.ID, fetched.ProtocolID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		cycle := createTestChemoCycle(t, ctx, tenantID, proto.ID, 3)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewChemoCycleRepoPG(globalDB.Pool)
			cycle.Status = "in-progress"
			cycle.ActualStartDate = &now
			cycle.DoseReductionPct = ptrFloat(10)
			cycle.DoseReductionReason = ptrStr("Neutropenia")
			return repo.Update(ctx, cycle)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *oncology.ChemoCycle
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewChemoCycleRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, cycle.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "in-progress" {
			t.Errorf("expected status=in-progress, got %s", fetched.Status)
		}
		if fetched.DoseReductionPct == nil || *fetched.DoseReductionPct != 10 {
			t.Errorf("expected dose_reduction_pct=10, got %v", fetched.DoseReductionPct)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		cycle := createTestChemoCycle(t, ctx, tenantID, proto.ID, 4)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewChemoCycleRepoPG(globalDB.Pool)
			return repo.Delete(ctx, cycle.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewChemoCycleRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, cycle.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted chemo cycle")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*oncology.ChemoCycle
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewChemoCycleRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 chemo cycle")
		}
		_ = results
	})

	t.Run("AddAdministration_and_GetAdministrations", func(t *testing.T) {
		cycle := createTestChemoCycle(t, ctx, tenantID, proto.ID, 5)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewChemoCycleRepoPG(globalDB.Pool)
			admin := &oncology.ChemoAdministration{
				CycleID:                cycle.ID,
				DrugName:               "Doxorubicin",
				AdministrationDatetime: now,
				DoseGiven:              ptrFloat(108),
				DoseUnit:               ptrStr("mg"),
				Route:                  ptrStr("IV"),
				InfusionDurationMin:    ptrInt(30),
				Site:                   ptrStr("Right arm port"),
				SequenceNumber:         ptrInt(1),
				Note:                   ptrStr("Administered without issues"),
			}
			return repo.AddAdministration(ctx, admin)
		})
		if err != nil {
			t.Fatalf("AddAdministration: %v", err)
		}

		var admins []*oncology.ChemoAdministration
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewChemoCycleRepoPG(globalDB.Pool)
			var err error
			admins, err = repo.GetAdministrations(ctx, cycle.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetAdministrations: %v", err)
		}
		if len(admins) != 1 {
			t.Fatalf("expected 1 administration, got %d", len(admins))
		}
		if admins[0].DrugName != "Doxorubicin" {
			t.Errorf("expected drug=Doxorubicin, got %s", admins[0].DrugName)
		}
	})
}

func TestRadiationTherapyCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("oncrad")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "RadPatient", "Test", "MRN-RAD-001")
	dx := createTestCancerDiagnosis(t, ctx, tenantID, patient.ID)

	t.Run("Create", func(t *testing.T) {
		var created *oncology.RadiationTherapy
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
			rt := &oncology.RadiationTherapy{
				CancerDiagnosisID:  dx.ID,
				TherapyType:        ptrStr("external"),
				Modality:           ptrStr("IMRT"),
				Technique:          ptrStr("VMAT"),
				TargetSite:         ptrStr("Left breast"),
				TotalDoseCGY:       ptrFloat(5000),
				DosePerFractionCGY: ptrFloat(200),
				PlannedFractions:   ptrInt(25),
				StartDate:          &now,
				Status:             "active",
				EnergyType:         ptrStr("photon"),
				EnergyValue:        ptrStr("6MV"),
				Note:               ptrStr("Standard fractionation"),
			}
			if err := repo.Create(ctx, rt); err != nil {
				return err
			}
			created = rt
			return nil
		})
		if err != nil {
			t.Fatalf("Create radiation therapy: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		rt := createTestRadiationTherapy(t, ctx, tenantID, dx.ID)

		var fetched *oncology.RadiationTherapy
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, rt.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CancerDiagnosisID != dx.ID {
			t.Errorf("expected cancer_diagnosis_id=%s, got %s", dx.ID, fetched.CancerDiagnosisID)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		rt := createTestRadiationTherapy(t, ctx, tenantID, dx.ID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
			rt.Status = "completed"
			rt.CompletedFractions = ptrInt(25)
			rt.EndDate = &now
			rt.Note = ptrStr("Treatment completed without complications")
			return repo.Update(ctx, rt)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *oncology.RadiationTherapy
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, rt.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.CompletedFractions == nil || *fetched.CompletedFractions != 25 {
			t.Errorf("expected completed_fractions=25, got %v", fetched.CompletedFractions)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		rt := createTestRadiationTherapy(t, ctx, tenantID, dx.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
			return repo.Delete(ctx, rt.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, rt.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted radiation therapy")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*oncology.RadiationTherapy
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 radiation therapy")
		}
		_ = results
	})

	t.Run("AddSession_and_GetSessions", func(t *testing.T) {
		rt := createTestRadiationTherapy(t, ctx, tenantID, dx.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
			session := &oncology.RadiationSession{
				RadiationTherapyID: rt.ID,
				SessionNumber:      1,
				SessionDate:        time.Now(),
				DoseDeliveredCGY:   ptrFloat(200),
				FieldName:          ptrStr("AP/PA"),
				SetupVerified:      ptrBool(true),
				ImagingType:        ptrStr("CBCT"),
				SkinReactionGrade:  ptrInt(0),
				FatigueGrade:       ptrInt(0),
				MachineID:          ptrStr("LINAC-01"),
				Note:               ptrStr("First fraction delivered"),
			}
			return repo.AddSession(ctx, session)
		})
		if err != nil {
			t.Fatalf("AddSession: %v", err)
		}

		// Add second session
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
			session := &oncology.RadiationSession{
				RadiationTherapyID: rt.ID,
				SessionNumber:      2,
				SessionDate:        time.Now().Add(24 * time.Hour),
				DoseDeliveredCGY:   ptrFloat(200),
				SetupVerified:      ptrBool(true),
				SkinReactionGrade:  ptrInt(1),
			}
			return repo.AddSession(ctx, session)
		})
		if err != nil {
			t.Fatalf("AddSession (second): %v", err)
		}

		var sessions []*oncology.RadiationSession
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
			var err error
			sessions, err = repo.GetSessions(ctx, rt.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetSessions: %v", err)
		}
		if len(sessions) != 2 {
			t.Fatalf("expected 2 sessions, got %d", len(sessions))
		}
		if sessions[0].SessionNumber != 1 {
			t.Errorf("expected first session number=1, got %d", sessions[0].SessionNumber)
		}
	})
}

func TestTumorMarkerCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("oncmarker")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "MarkerPatient", "Test", "MRN-MARKER-001")

	t.Run("Create", func(t *testing.T) {
		var created *oncology.TumorMarker
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorMarkerRepoPG(globalDB.Pool)
			m := &oncology.TumorMarker{
				PatientID:           patient.ID,
				MarkerName:          "CA-125",
				MarkerCode:          ptrStr("10334-1"),
				MarkerCodeSystem:    ptrStr("http://loinc.org"),
				ValueQuantity:       ptrFloat(35.0),
				ValueUnit:           ptrStr("U/mL"),
				ValueInterpretation: ptrStr("normal"),
				ReferenceRangeLow:   ptrFloat(0),
				ReferenceRangeHigh:  ptrFloat(35),
				SpecimenType:        ptrStr("serum"),
				CollectionDatetime:  &now,
				ResultDatetime:      &now,
				PerformingLab:       ptrStr("Hospital Lab"),
				Note:                ptrStr("Baseline measurement"),
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			created = m
			return nil
		})
		if err != nil {
			t.Fatalf("Create tumor marker: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		marker := createTestTumorMarker(t, ctx, tenantID, patient.ID)

		var fetched *oncology.TumorMarker
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorMarkerRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, marker.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.MarkerName != "PSA" {
			t.Errorf("expected marker_name=PSA, got %s", fetched.MarkerName)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		marker := createTestTumorMarker(t, ctx, tenantID, patient.ID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorMarkerRepoPG(globalDB.Pool)
			marker.ValueQuantity = ptrFloat(8.5)
			marker.ValueInterpretation = ptrStr("elevated")
			marker.ResultDatetime = &now
			marker.Note = ptrStr("Rising trend")
			return repo.Update(ctx, marker)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *oncology.TumorMarker
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorMarkerRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, marker.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.ValueQuantity == nil || *fetched.ValueQuantity != 8.5 {
			t.Errorf("expected value_quantity=8.5, got %v", fetched.ValueQuantity)
		}
		if fetched.ValueInterpretation == nil || *fetched.ValueInterpretation != "elevated" {
			t.Errorf("expected interpretation=elevated, got %v", fetched.ValueInterpretation)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		marker := createTestTumorMarker(t, ctx, tenantID, patient.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorMarkerRepoPG(globalDB.Pool)
			return repo.Delete(ctx, marker.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorMarkerRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, marker.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted tumor marker")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*oncology.TumorMarker
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorMarkerRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 tumor marker")
		}
		_ = results
	})
}

func TestTumorBoardCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("oncboard")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "BoardPatient", "Test", "MRN-BOARD-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "BoardDoc", "Smith")
	dx := createTestCancerDiagnosis(t, ctx, tenantID, patient.ID)

	t.Run("Create", func(t *testing.T) {
		var created *oncology.TumorBoardReview
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorBoardRepoPG(globalDB.Pool)
			b := &oncology.TumorBoardReview{
				CancerDiagnosisID:      dx.ID,
				PatientID:              patient.ID,
				ReviewDate:             time.Now(),
				ReviewType:             ptrStr("initial"),
				PresentingProviderID:   &practitioner.ID,
				Attendees:              ptrStr("Dr. Smith, Dr. Jones, Dr. Williams"),
				ClinicalSummary:        ptrStr("52yo female with Stage IIA breast cancer"),
				PathologySummary:       ptrStr("IDC, ER+/PR+/HER2-"),
				Recommendations:        ptrStr("Recommend AC-T followed by endocrine therapy"),
				TreatmentDecision:      ptrStr("Proceed with chemotherapy"),
				ClinicalTrialDiscussed: ptrBool(true),
				Note:                   ptrStr("Unanimous consensus"),
			}
			if err := repo.Create(ctx, b); err != nil {
				return err
			}
			created = b
			return nil
		})
		if err != nil {
			t.Fatalf("Create tumor board: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		board := createTestTumorBoard(t, ctx, tenantID, dx.ID, patient.ID)

		var fetched *oncology.TumorBoardReview
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorBoardRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, board.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CancerDiagnosisID != dx.ID {
			t.Errorf("expected cancer_diagnosis_id=%s, got %s", dx.ID, fetched.CancerDiagnosisID)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		board := createTestTumorBoard(t, ctx, tenantID, dx.ID, patient.ID)

		nextReview := time.Now().Add(90 * 24 * time.Hour)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorBoardRepoPG(globalDB.Pool)
			board.Recommendations = ptrStr("Continue current treatment plan")
			board.TreatmentDecision = ptrStr("Maintain AC-T protocol")
			board.NextReviewDate = &nextReview
			board.Note = ptrStr("Good response, review in 3 months")
			return repo.Update(ctx, board)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *oncology.TumorBoardReview
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorBoardRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, board.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.NextReviewDate == nil {
			t.Error("expected non-nil NextReviewDate")
		}
		if fetched.Recommendations == nil || *fetched.Recommendations != "Continue current treatment plan" {
			t.Errorf("expected updated recommendations, got %v", fetched.Recommendations)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		board := createTestTumorBoard(t, ctx, tenantID, dx.ID, patient.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorBoardRepoPG(globalDB.Pool)
			return repo.Delete(ctx, board.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorBoardRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, board.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted tumor board")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*oncology.TumorBoardReview
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := oncology.NewTumorBoardRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 tumor board review")
		}
		_ = results
	})
}

// =========== Test Helpers ===========

func createTestCancerDiagnosis(t *testing.T, ctx context.Context, tenantID string, patientID uuid.UUID) *oncology.CancerDiagnosis {
	t.Helper()
	var result *oncology.CancerDiagnosis
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := oncology.NewCancerDiagnosisRepoPG(globalDB.Pool)
		d := &oncology.CancerDiagnosis{
			PatientID:     patientID,
			DiagnosisDate: time.Now(),
			CancerType:    ptrStr("lung"),
			CancerSite:    ptrStr("C34.1"),
			CurrentStatus: "active",
			ICD10Code:     ptrStr("C34.10"),
			ICD10Display:  ptrStr("Malignant neoplasm of upper lobe, bronchus or lung"),
		}
		if err := repo.Create(ctx, d); err != nil {
			return err
		}
		result = d
		return nil
	})
	if err != nil {
		t.Fatalf("create test cancer diagnosis: %v", err)
	}
	return result
}

func createTestTreatmentProtocol(t *testing.T, ctx context.Context, tenantID string, dxID uuid.UUID) *oncology.TreatmentProtocol {
	t.Helper()
	var result *oncology.TreatmentProtocol
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := oncology.NewTreatmentProtocolRepoPG(globalDB.Pool)
		p := &oncology.TreatmentProtocol{
			CancerDiagnosisID: dxID,
			ProtocolName:      "FOLFOX",
			ProtocolType:      ptrStr("chemotherapy"),
			Intent:            ptrStr("curative"),
			NumberOfCycles:    ptrInt(12),
			CycleLengthDays:   ptrInt(14),
			Status:            "active",
		}
		if err := repo.Create(ctx, p); err != nil {
			return err
		}
		result = p
		return nil
	})
	if err != nil {
		t.Fatalf("create test treatment protocol: %v", err)
	}
	return result
}

func createTestChemoCycle(t *testing.T, ctx context.Context, tenantID string, protocolID uuid.UUID, cycleNum int) *oncology.ChemoCycle {
	t.Helper()
	var result *oncology.ChemoCycle
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := oncology.NewChemoCycleRepoPG(globalDB.Pool)
		c := &oncology.ChemoCycle{
			ProtocolID:  protocolID,
			CycleNumber: cycleNum,
			Status:      "planned",
			BSAM2:       ptrFloat(1.8),
			WeightKG:    ptrFloat(75),
		}
		if err := repo.Create(ctx, c); err != nil {
			return err
		}
		result = c
		return nil
	})
	if err != nil {
		t.Fatalf("create test chemo cycle: %v", err)
	}
	return result
}

func createTestRadiationTherapy(t *testing.T, ctx context.Context, tenantID string, dxID uuid.UUID) *oncology.RadiationTherapy {
	t.Helper()
	var result *oncology.RadiationTherapy
	now := time.Now()
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := oncology.NewRadiationTherapyRepoPG(globalDB.Pool)
		rt := &oncology.RadiationTherapy{
			CancerDiagnosisID:  dxID,
			TherapyType:        ptrStr("external"),
			Modality:           ptrStr("3D-CRT"),
			TargetSite:         ptrStr("Left breast"),
			TotalDoseCGY:       ptrFloat(5040),
			DosePerFractionCGY: ptrFloat(180),
			PlannedFractions:   ptrInt(28),
			StartDate:          &now,
			Status:             "active",
		}
		if err := repo.Create(ctx, rt); err != nil {
			return err
		}
		result = rt
		return nil
	})
	if err != nil {
		t.Fatalf("create test radiation therapy: %v", err)
	}
	return result
}

func createTestTumorMarker(t *testing.T, ctx context.Context, tenantID string, patientID uuid.UUID) *oncology.TumorMarker {
	t.Helper()
	var result *oncology.TumorMarker
	now := time.Now()
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := oncology.NewTumorMarkerRepoPG(globalDB.Pool)
		m := &oncology.TumorMarker{
			PatientID:          patientID,
			MarkerName:         "PSA",
			MarkerCode:         ptrStr("2857-1"),
			ValueQuantity:      ptrFloat(4.0),
			ValueUnit:          ptrStr("ng/mL"),
			SpecimenType:       ptrStr("serum"),
			CollectionDatetime: &now,
		}
		if err := repo.Create(ctx, m); err != nil {
			return err
		}
		result = m
		return nil
	})
	if err != nil {
		t.Fatalf("create test tumor marker: %v", err)
	}
	return result
}

func createTestTumorBoard(t *testing.T, ctx context.Context, tenantID string, dxID, patientID uuid.UUID) *oncology.TumorBoardReview {
	t.Helper()
	var result *oncology.TumorBoardReview
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := oncology.NewTumorBoardRepoPG(globalDB.Pool)
		b := &oncology.TumorBoardReview{
			CancerDiagnosisID: dxID,
			PatientID:         patientID,
			ReviewDate:        time.Now(),
			ReviewType:        ptrStr("follow-up"),
			ClinicalSummary:   ptrStr("Ongoing treatment review"),
			Recommendations:   ptrStr("Continue treatment"),
		}
		if err := repo.Create(ctx, b); err != nil {
			return err
		}
		result = b
		return nil
	})
	if err != nil {
		t.Fatalf("create test tumor board: %v", err)
	}
	return result
}
