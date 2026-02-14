package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/obstetrics"
	"github.com/google/uuid"
)

func TestPregnancyCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("preg")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "PregPatient", "Test", "MRN-PREG-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "OBDoc", "Williams")

	t.Run("Create", func(t *testing.T) {
		var created *obstetrics.Pregnancy
		lmp := time.Now().Add(-10 * 7 * 24 * time.Hour)
		edd := lmp.Add(280 * 24 * time.Hour)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			p := &obstetrics.Pregnancy{
				PatientID:         patient.ID,
				Status:            "active",
				OnsetDate:         &lmp,
				EstimatedDueDate:  &edd,
				LastMenstrualPeriod: &lmp,
				ConceptionMethod:  ptrStr("natural"),
				Gravida:           ptrInt(2),
				Para:              ptrInt(1),
				MultipleGestation: ptrBool(false),
				NumberOfFetuses:   ptrInt(1),
				RiskLevel:         ptrStr("low"),
				BloodType:         ptrStr("O"),
				RhFactor:          ptrStr("positive"),
				PrePregnancyWeight: ptrFloat(65.0),
				PrePregnancyBMI:   ptrFloat(23.5),
				PrimaryProviderID: &practitioner.ID,
				Note:              ptrStr("Routine pregnancy, no complications"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			created = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create pregnancy: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			p := &obstetrics.Pregnancy{
				PatientID: uuid.New(), // non-existent
				Status:    "active",
			}
			return repo.Create(ctx, p)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var pregID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			p := &obstetrics.Pregnancy{
				PatientID: patient.ID,
				Status:    "active",
				Gravida:   ptrInt(1),
				Para:      ptrInt(0),
				RiskLevel: ptrStr("high"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			pregID = p.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *obstetrics.Pregnancy
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pregID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.RiskLevel == nil || *fetched.RiskLevel != "high" {
			t.Errorf("expected risk_level=high, got %v", fetched.RiskLevel)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var preg *obstetrics.Pregnancy
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			p := &obstetrics.Pregnancy{
				PatientID: patient.ID,
				Status:    "active",
				RiskLevel: ptrStr("low"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			preg = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		outcomeDate := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			preg.Status = "completed"
			preg.RiskLevel = ptrStr("high")
			preg.RiskFactors = ptrStr("gestational diabetes")
			preg.OutcomeDate = &outcomeDate
			preg.OutcomeSummary = ptrStr("Healthy delivery, vaginal birth")
			preg.Note = ptrStr("Pregnancy completed without major complications")
			return repo.Update(ctx, preg)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *obstetrics.Pregnancy
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, preg.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.OutcomeDate == nil {
			t.Error("expected non-nil OutcomeDate")
		}
		if fetched.OutcomeSummary == nil || *fetched.OutcomeSummary != "Healthy delivery, vaginal birth" {
			t.Errorf("expected outcome_summary updated, got %v", fetched.OutcomeSummary)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*obstetrics.Pregnancy
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 pregnancy")
		}
		_ = results
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*obstetrics.Pregnancy
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 pregnancy")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var pregID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			p := &obstetrics.Pregnancy{
				PatientID: patient.ID,
				Status:    "ectopic",
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			pregID = p.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			return repo.Delete(ctx, pregID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, pregID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted pregnancy")
		}
	})
}

func TestPrenatalVisitCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("prenatal")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "PrenatalPatient", "Test", "MRN-PREN-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "PrenatalDoc", "Taylor")

	// Create pregnancy first
	var pregID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
		p := &obstetrics.Pregnancy{
			PatientID: patient.ID,
			Status:    "active",
			Gravida:   ptrInt(1),
			Para:      ptrInt(0),
		}
		if err := repo.Create(ctx, p); err != nil {
			return err
		}
		pregID = p.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create pregnancy: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *obstetrics.PrenatalVisit
		now := time.Now()
		nextVisit := now.Add(14 * 24 * time.Hour)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPrenatalVisitRepoPG(globalDB.Pool)
			v := &obstetrics.PrenatalVisit{
				PregnancyID:            pregID,
				VisitDate:              now,
				GestationalAgeWeeks:    ptrInt(12),
				GestationalAgeDays:     ptrInt(3),
				Weight:                 ptrFloat(68.5),
				BloodPressureSystolic:  ptrInt(118),
				BloodPressureDiastolic: ptrInt(72),
				FundalHeight:           ptrFloat(12.0),
				FetalHeartRate:         ptrInt(155),
				FetalPresentation:      ptrStr("cephalic"),
				FetalMovement:          ptrStr("positive"),
				UrineProtein:           ptrStr("negative"),
				UrineGlucose:           ptrStr("negative"),
				ProviderID:             &practitioner.ID,
				Note:                   ptrStr("Routine first trimester visit"),
				NextVisitDate:          &nextVisit,
			}
			if err := repo.Create(ctx, v); err != nil {
				return err
			}
			created = v
			return nil
		})
		if err != nil {
			t.Fatalf("Create prenatal visit: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var visitID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPrenatalVisitRepoPG(globalDB.Pool)
			v := &obstetrics.PrenatalVisit{
				PregnancyID:         pregID,
				VisitDate:           now,
				GestationalAgeWeeks: ptrInt(16),
				Weight:              ptrFloat(70.0),
				FetalHeartRate:      ptrInt(148),
			}
			if err := repo.Create(ctx, v); err != nil {
				return err
			}
			visitID = v.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *obstetrics.PrenatalVisit
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPrenatalVisitRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, visitID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.GestationalAgeWeeks == nil || *fetched.GestationalAgeWeeks != 16 {
			t.Errorf("expected gestational_age_weeks=16, got %v", fetched.GestationalAgeWeeks)
		}
		if fetched.PregnancyID != pregID {
			t.Errorf("expected pregnancy_id=%s, got %s", pregID, fetched.PregnancyID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var visit *obstetrics.PrenatalVisit
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPrenatalVisitRepoPG(globalDB.Pool)
			v := &obstetrics.PrenatalVisit{
				PregnancyID:         pregID,
				VisitDate:           now,
				GestationalAgeWeeks: ptrInt(20),
			}
			if err := repo.Create(ctx, v); err != nil {
				return err
			}
			visit = v
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		nextVisit := now.Add(14 * 24 * time.Hour)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPrenatalVisitRepoPG(globalDB.Pool)
			visit.Weight = ptrFloat(72.0)
			visit.BloodPressureSystolic = ptrInt(120)
			visit.BloodPressureDiastolic = ptrInt(75)
			visit.FundalHeight = ptrFloat(20.0)
			visit.FetalHeartRate = ptrInt(142)
			visit.FetalPresentation = ptrStr("cephalic")
			visit.UrineProtein = ptrStr("trace")
			visit.Note = ptrStr("Anatomy scan normal")
			visit.NextVisitDate = &nextVisit
			return repo.Update(ctx, visit)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *obstetrics.PrenatalVisit
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPrenatalVisitRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, visit.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Weight == nil || *fetched.Weight != 72.0 {
			t.Errorf("expected weight=72.0, got %v", fetched.Weight)
		}
		if fetched.UrineProtein == nil || *fetched.UrineProtein != "trace" {
			t.Errorf("expected urine_protein=trace, got %v", fetched.UrineProtein)
		}
	})

	t.Run("ListByPregnancy", func(t *testing.T) {
		var results []*obstetrics.PrenatalVisit
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPrenatalVisitRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPregnancy(ctx, pregID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPregnancy: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 prenatal visit")
		}
		for _, r := range results {
			if r.PregnancyID != pregID {
				t.Errorf("expected pregnancy_id=%s, got %s", pregID, r.PregnancyID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var visitID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPrenatalVisitRepoPG(globalDB.Pool)
			v := &obstetrics.PrenatalVisit{
				PregnancyID: pregID,
				VisitDate:   now,
			}
			if err := repo.Create(ctx, v); err != nil {
				return err
			}
			visitID = v.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPrenatalVisitRepoPG(globalDB.Pool)
			return repo.Delete(ctx, visitID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPrenatalVisitRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, visitID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted prenatal visit")
		}
	})
}

func TestLaborRecordCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("labor")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "LaborPatient", "Test", "MRN-LABOR-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "LaborDoc", "Anderson")

	// Create pregnancy first
	var pregID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
		p := &obstetrics.Pregnancy{
			PatientID: patient.ID,
			Status:    "active",
			Gravida:   ptrInt(1),
			Para:      ptrInt(0),
		}
		if err := repo.Create(ctx, p); err != nil {
			return err
		}
		pregID = p.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create pregnancy: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *obstetrics.LaborRecord
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			l := &obstetrics.LaborRecord{
				PregnancyID:        pregID,
				AdmissionDatetime:  &now,
				LaborOnsetDatetime: &now,
				LaborOnsetType:     ptrStr("spontaneous"),
				AmnioticFluidColor: ptrStr("clear"),
				AnesthesiaType:     ptrStr("epidural"),
				Status:             "active",
				AttendingProviderID: &practitioner.ID,
				Note:               ptrStr("Active labor, 5cm dilated on admission"),
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			created = l
			return nil
		})
		if err != nil {
			t.Fatalf("Create labor record: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var laborID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			l := &obstetrics.LaborRecord{
				PregnancyID:    pregID,
				LaborOnsetType: ptrStr("induced"),
				InductionMethod: ptrStr("oxytocin"),
				InductionReason: ptrStr("post-dates"),
				Status:         "active",
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			laborID = l.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *obstetrics.LaborRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, laborID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.LaborOnsetType == nil || *fetched.LaborOnsetType != "induced" {
			t.Errorf("expected labor_onset_type=induced, got %v", fetched.LaborOnsetType)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var labor *obstetrics.LaborRecord
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			l := &obstetrics.LaborRecord{
				PregnancyID: pregID,
				Status:      "active",
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			labor = l
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		anesthesiaStart := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			labor.Status = "completed"
			labor.AnesthesiaType = ptrStr("epidural")
			labor.AnesthesiaStart = &anesthesiaStart
			labor.Note = ptrStr("Delivered at 10cm, epidural effective")
			return repo.Update(ctx, labor)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *obstetrics.LaborRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, labor.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.AnesthesiaType == nil || *fetched.AnesthesiaType != "epidural" {
			t.Errorf("expected anesthesia_type=epidural, got %v", fetched.AnesthesiaType)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*obstetrics.LaborRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 labor record")
		}
		_ = results
	})

	t.Run("CervicalExams", func(t *testing.T) {
		now := time.Now()
		var laborID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			l := &obstetrics.LaborRecord{
				PregnancyID: pregID,
				Status:      "active",
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			laborID = l.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create labor: %v", err)
		}

		// Add cervical exams
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			e1 := &obstetrics.LaborCervicalExam{
				LaborRecordID: laborID,
				ExamDatetime:  now,
				DilationCM:    ptrFloat(5.0),
				EffacementPct: ptrInt(80),
				Station:       ptrStr("-1"),
				FetalPosition: ptrStr("LOA"),
				MembraneStatus: ptrStr("intact"),
				ExaminerID:    &practitioner.ID,
				Note:          ptrStr("Active labor"),
			}
			if err := repo.AddCervicalExam(ctx, e1); err != nil {
				return err
			}

			e2 := &obstetrics.LaborCervicalExam{
				LaborRecordID: laborID,
				ExamDatetime:  now.Add(2 * time.Hour),
				DilationCM:    ptrFloat(8.0),
				EffacementPct: ptrInt(100),
				Station:       ptrStr("0"),
				FetalPosition: ptrStr("LOA"),
				MembraneStatus: ptrStr("ruptured"),
				ExaminerID:    &practitioner.ID,
				Note:          ptrStr("Transition phase"),
			}
			return repo.AddCervicalExam(ctx, e2)
		})
		if err != nil {
			t.Fatalf("AddCervicalExam: %v", err)
		}

		var exams []*obstetrics.LaborCervicalExam
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			var err error
			exams, err = repo.GetCervicalExams(ctx, laborID)
			return err
		})
		if err != nil {
			t.Fatalf("GetCervicalExams: %v", err)
		}
		if len(exams) != 2 {
			t.Fatalf("expected 2 cervical exams, got %d", len(exams))
		}
		for _, e := range exams {
			if e.LaborRecordID != laborID {
				t.Errorf("expected labor_record_id=%s, got %s", laborID, e.LaborRecordID)
			}
		}
	})

	t.Run("FetalMonitoring", func(t *testing.T) {
		now := time.Now()
		var laborID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			l := &obstetrics.LaborRecord{
				PregnancyID: pregID,
				Status:      "active",
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			laborID = l.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create labor: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			f := &obstetrics.FetalMonitoring{
				LaborRecordID:       laborID,
				MonitoringDatetime:  now,
				MonitoringType:      ptrStr("continuous EFM"),
				FetalHeartRate:      ptrInt(145),
				BaselineRate:        ptrInt(140),
				Variability:         ptrStr("moderate"),
				Accelerations:       ptrStr("present"),
				Decelerations:       ptrStr("none"),
				ContractionFrequency: ptrStr("q3min"),
				ContractionDuration: ptrStr("60-90s"),
				ContractionIntensity: ptrStr("strong"),
				Interpretation:      ptrStr("reassuring"),
				Category:            ptrStr("I"),
				RecorderID:          &practitioner.ID,
				Note:                ptrStr("Category I tracing, reassuring"),
			}
			return repo.AddFetalMonitoring(ctx, f)
		})
		if err != nil {
			t.Fatalf("AddFetalMonitoring: %v", err)
		}

		var monitors []*obstetrics.FetalMonitoring
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			var err error
			monitors, err = repo.GetFetalMonitoring(ctx, laborID)
			return err
		})
		if err != nil {
			t.Fatalf("GetFetalMonitoring: %v", err)
		}
		if len(monitors) != 1 {
			t.Fatalf("expected 1 fetal monitoring record, got %d", len(monitors))
		}
		if monitors[0].FetalHeartRate == nil || *monitors[0].FetalHeartRate != 145 {
			t.Errorf("expected fetal_heart_rate=145, got %v", monitors[0].FetalHeartRate)
		}
		if monitors[0].Category == nil || *monitors[0].Category != "I" {
			t.Errorf("expected category=I, got %v", monitors[0].Category)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var laborID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			l := &obstetrics.LaborRecord{
				PregnancyID: pregID,
				Status:      "cancelled",
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			laborID = l.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			return repo.Delete(ctx, laborID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewLaborRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, laborID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted labor record")
		}
	})
}

func TestDeliveryRecordCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("delivery")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "DeliveryPatient", "Test", "MRN-DELIV-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "DeliveryDoc", "Baker")

	// Create pregnancy
	var pregID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
		p := &obstetrics.Pregnancy{
			PatientID: patient.ID,
			Status:    "active",
		}
		if err := repo.Create(ctx, p); err != nil {
			return err
		}
		pregID = p.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create pregnancy: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *obstetrics.DeliveryRecord
		now := time.Now()
		placentaTime := now.Add(15 * time.Minute)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
			d := &obstetrics.DeliveryRecord{
				PregnancyID:          pregID,
				PatientID:            patient.ID,
				DeliveryDatetime:     now,
				DeliveryMethod:       "vaginal",
				DeliveryType:         ptrStr("spontaneous"),
				DeliveringProviderID: practitioner.ID,
				BirthOrder:           ptrInt(1),
				PlacentaDelivery:     ptrStr("spontaneous"),
				PlacentaDatetime:     &placentaTime,
				PlacentaIntact:       ptrBool(true),
				CordVessels:          ptrInt(3),
				CordBloodCollected:   ptrBool(false),
				Episiotomy:           ptrBool(false),
				LacerationDegree:     ptrStr("second"),
				RepairMethod:         ptrStr("suture"),
				BloodLossML:          ptrInt(350),
				Note:                 ptrStr("Uncomplicated vaginal delivery"),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			created = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create delivery record: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var delivID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
			d := &obstetrics.DeliveryRecord{
				PregnancyID:          pregID,
				PatientID:            patient.ID,
				DeliveryDatetime:     now,
				DeliveryMethod:       "cesarean",
				DeliveryType:         ptrStr("primary"),
				DeliveringProviderID: practitioner.ID,
				BloodLossML:          ptrInt(800),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			delivID = d.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *obstetrics.DeliveryRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, delivID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.DeliveryMethod != "cesarean" {
			t.Errorf("expected delivery_method=cesarean, got %s", fetched.DeliveryMethod)
		}
		if fetched.BloodLossML == nil || *fetched.BloodLossML != 800 {
			t.Errorf("expected blood_loss_ml=800, got %v", fetched.BloodLossML)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var deliv *obstetrics.DeliveryRecord
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
			d := &obstetrics.DeliveryRecord{
				PregnancyID:          pregID,
				PatientID:            patient.ID,
				DeliveryDatetime:     now,
				DeliveryMethod:       "vaginal",
				DeliveringProviderID: practitioner.ID,
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			deliv = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
			deliv.DeliveryMethod = "vacuum-assisted"
			deliv.DeliveryType = ptrStr("operative vaginal")
			deliv.PlacentaDelivery = ptrStr("manual")
			deliv.PlacentaIntact = ptrBool(true)
			deliv.BloodLossML = ptrInt(500)
			deliv.Complications = ptrStr("Shoulder dystocia resolved with McRoberts")
			deliv.Note = ptrStr("Updated with delivery details")
			return repo.Update(ctx, deliv)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *obstetrics.DeliveryRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, deliv.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.DeliveryMethod != "vacuum-assisted" {
			t.Errorf("expected delivery_method=vacuum-assisted, got %s", fetched.DeliveryMethod)
		}
		if fetched.Complications == nil || *fetched.Complications != "Shoulder dystocia resolved with McRoberts" {
			t.Errorf("expected complications updated, got %v", fetched.Complications)
		}
	})

	t.Run("ListByPregnancy", func(t *testing.T) {
		var results []*obstetrics.DeliveryRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPregnancy(ctx, pregID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPregnancy: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 delivery record")
		}
		for _, r := range results {
			if r.PregnancyID != pregID {
				t.Errorf("expected pregnancy_id=%s, got %s", pregID, r.PregnancyID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var delivID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
			d := &obstetrics.DeliveryRecord{
				PregnancyID:          pregID,
				PatientID:            patient.ID,
				DeliveryDatetime:     now,
				DeliveryMethod:       "vaginal",
				DeliveringProviderID: practitioner.ID,
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			delivID = d.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
			return repo.Delete(ctx, delivID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, delivID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted delivery record")
		}
	})
}

func TestNewbornRecordCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("newborn")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "NewbornMother", "Test", "MRN-NEWB-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "NewbornDoc", "Carter")

	// Create pregnancy and delivery
	var pregID, delivID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		pregRepo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
		p := &obstetrics.Pregnancy{
			PatientID: patient.ID,
			Status:    "completed",
		}
		if err := pregRepo.Create(ctx, p); err != nil {
			return err
		}
		pregID = p.ID

		delivRepo := obstetrics.NewDeliveryRepoPG(globalDB.Pool)
		d := &obstetrics.DeliveryRecord{
			PregnancyID:          pregID,
			PatientID:            patient.ID,
			DeliveryDatetime:     time.Now(),
			DeliveryMethod:       "vaginal",
			DeliveringProviderID: practitioner.ID,
		}
		if err := delivRepo.Create(ctx, d); err != nil {
			return err
		}
		delivID = d.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create pregnancy and delivery: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *obstetrics.NewbornRecord
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewNewbornRepoPG(globalDB.Pool)
			n := &obstetrics.NewbornRecord{
				DeliveryID:          delivID,
				BirthDatetime:       now,
				Sex:                 ptrStr("female"),
				BirthWeightGrams:    ptrInt(3250),
				BirthLengthCM:       ptrFloat(50.5),
				HeadCircumferenceCM: ptrFloat(34.0),
				Apgar1Min:           ptrInt(8),
				Apgar5Min:           ptrInt(9),
				GestationalAgeWeeks: ptrInt(39),
				GestationalAgeDays:  ptrInt(2),
				BirthStatus:         ptrStr("liveborn"),
				NICUAdmission:       ptrBool(false),
				VitaminKGiven:       ptrBool(true),
				EyeProphylaxisGiven: ptrBool(true),
				HepatitisBGiven:     ptrBool(true),
				FeedingMethod:       ptrStr("breastfeeding"),
				Note:                ptrStr("Healthy newborn, good cry"),
			}
			if err := repo.Create(ctx, n); err != nil {
				return err
			}
			created = n
			return nil
		})
		if err != nil {
			t.Fatalf("Create newborn record: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var newbornID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewNewbornRepoPG(globalDB.Pool)
			n := &obstetrics.NewbornRecord{
				DeliveryID:       delivID,
				BirthDatetime:    now,
				Sex:              ptrStr("male"),
				BirthWeightGrams: ptrInt(2800),
				Apgar1Min:        ptrInt(6),
				Apgar5Min:        ptrInt(8),
				BirthStatus:      ptrStr("liveborn"),
				NICUAdmission:    ptrBool(true),
				NICUReason:       ptrStr("low birth weight"),
			}
			if err := repo.Create(ctx, n); err != nil {
				return err
			}
			newbornID = n.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *obstetrics.NewbornRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewNewbornRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, newbornID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Sex == nil || *fetched.Sex != "male" {
			t.Errorf("expected sex=male, got %v", fetched.Sex)
		}
		if fetched.NICUAdmission == nil || !*fetched.NICUAdmission {
			t.Error("expected nicu_admission=true")
		}
		if fetched.DeliveryID != delivID {
			t.Errorf("expected delivery_id=%s, got %s", delivID, fetched.DeliveryID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var newborn *obstetrics.NewbornRecord
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewNewbornRepoPG(globalDB.Pool)
			n := &obstetrics.NewbornRecord{
				DeliveryID:       delivID,
				BirthDatetime:    now,
				BirthWeightGrams: ptrInt(3100),
				BirthStatus:      ptrStr("liveborn"),
			}
			if err := repo.Create(ctx, n); err != nil {
				return err
			}
			newborn = n
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewNewbornRepoPG(globalDB.Pool)
			newborn.Sex = ptrStr("female")
			newborn.BirthWeightGrams = ptrInt(3150)
			newborn.BirthLengthCM = ptrFloat(49.0)
			newborn.Apgar1Min = ptrInt(9)
			newborn.Apgar5Min = ptrInt(9)
			newborn.Apgar10Min = ptrInt(10)
			newborn.FeedingMethod = ptrStr("formula")
			newborn.Note = ptrStr("Updated with complete measurements")
			return repo.Update(ctx, newborn)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *obstetrics.NewbornRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewNewbornRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, newborn.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Sex == nil || *fetched.Sex != "female" {
			t.Errorf("expected sex=female, got %v", fetched.Sex)
		}
		if fetched.BirthWeightGrams == nil || *fetched.BirthWeightGrams != 3150 {
			t.Errorf("expected birth_weight_grams=3150, got %v", fetched.BirthWeightGrams)
		}
		if fetched.FeedingMethod == nil || *fetched.FeedingMethod != "formula" {
			t.Errorf("expected feeding_method=formula, got %v", fetched.FeedingMethod)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*obstetrics.NewbornRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewNewbornRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 newborn record")
		}
		_ = results
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var newbornID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewNewbornRepoPG(globalDB.Pool)
			n := &obstetrics.NewbornRecord{
				DeliveryID:    delivID,
				BirthDatetime: now,
				BirthStatus:   ptrStr("liveborn"),
			}
			if err := repo.Create(ctx, n); err != nil {
				return err
			}
			newbornID = n.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewNewbornRepoPG(globalDB.Pool)
			return repo.Delete(ctx, newbornID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewNewbornRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, newbornID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted newborn record")
		}
	})
}

func TestPostpartumRecordCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("postpartum")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "PostpartumPatient", "Test", "MRN-POST-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "PostpartumDoc", "Davis")

	// Create pregnancy
	var pregID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := obstetrics.NewPregnancyRepoPG(globalDB.Pool)
		p := &obstetrics.Pregnancy{
			PatientID: patient.ID,
			Status:    "completed",
		}
		if err := repo.Create(ctx, p); err != nil {
			return err
		}
		pregID = p.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create pregnancy: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *obstetrics.PostpartumRecord
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPostpartumRepoPG(globalDB.Pool)
			p := &obstetrics.PostpartumRecord{
				PregnancyID:            pregID,
				PatientID:              patient.ID,
				VisitDate:              now,
				DaysPostpartum:         ptrInt(2),
				UterineInvolution:      ptrStr("firm, at umbilicus"),
				LochiaType:             ptrStr("rubra"),
				LochiaAmount:           ptrStr("moderate"),
				PerineumStatus:         ptrStr("intact, mild edema"),
				BreastStatus:           ptrStr("engorged"),
				BreastfeedingStatus:    ptrStr("establishing"),
				MoodScreeningScore:     ptrInt(5),
				MoodScreeningTool:      ptrStr("PHQ-9"),
				DepressionRisk:         ptrStr("low"),
				BloodPressureSystolic:  ptrInt(115),
				BloodPressureDiastolic: ptrInt(70),
				Weight:                 ptrFloat(72.0),
				ProviderID:             &practitioner.ID,
				Note:                   ptrStr("Day 2 postpartum, recovering well"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			created = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create postpartum record: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var recID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPostpartumRepoPG(globalDB.Pool)
			p := &obstetrics.PostpartumRecord{
				PregnancyID:    pregID,
				PatientID:      patient.ID,
				VisitDate:      now,
				WeeksPostpartum: ptrInt(6),
				MoodScreeningScore: ptrInt(3),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			recID = p.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *obstetrics.PostpartumRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPostpartumRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, recID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.WeeksPostpartum == nil || *fetched.WeeksPostpartum != 6 {
			t.Errorf("expected weeks_postpartum=6, got %v", fetched.WeeksPostpartum)
		}
		if fetched.PregnancyID != pregID {
			t.Errorf("expected pregnancy_id=%s, got %s", pregID, fetched.PregnancyID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var rec *obstetrics.PostpartumRecord
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPostpartumRepoPG(globalDB.Pool)
			p := &obstetrics.PostpartumRecord{
				PregnancyID: pregID,
				PatientID:   patient.ID,
				VisitDate:   now,
				DaysPostpartum: ptrInt(14),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			rec = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPostpartumRepoPG(globalDB.Pool)
			rec.UterineInvolution = ptrStr("firm, 2cm below umbilicus")
			rec.LochiaType = ptrStr("serosa")
			rec.LochiaAmount = ptrStr("scant")
			rec.PerineumStatus = ptrStr("healing well")
			rec.BreastStatus = ptrStr("soft, non-tender")
			rec.BreastfeedingStatus = ptrStr("established")
			rec.ContraceptionPlan = ptrStr("IUD planned at 6 weeks")
			rec.MoodScreeningScore = ptrInt(2)
			rec.MoodScreeningTool = ptrStr("EPDS")
			rec.DepressionRisk = ptrStr("low")
			rec.BloodPressureSystolic = ptrInt(110)
			rec.BloodPressureDiastolic = ptrInt(68)
			rec.Weight = ptrFloat(69.0)
			rec.Note = ptrStr("2-week follow-up, recovering well")
			return repo.Update(ctx, rec)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *obstetrics.PostpartumRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPostpartumRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, rec.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.LochiaType == nil || *fetched.LochiaType != "serosa" {
			t.Errorf("expected lochia_type=serosa, got %v", fetched.LochiaType)
		}
		if fetched.BreastfeedingStatus == nil || *fetched.BreastfeedingStatus != "established" {
			t.Errorf("expected breastfeeding_status=established, got %v", fetched.BreastfeedingStatus)
		}
		if fetched.ContraceptionPlan == nil || *fetched.ContraceptionPlan != "IUD planned at 6 weeks" {
			t.Errorf("expected contraception_plan updated, got %v", fetched.ContraceptionPlan)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*obstetrics.PostpartumRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPostpartumRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 postpartum record")
		}
		_ = results
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var recID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPostpartumRepoPG(globalDB.Pool)
			p := &obstetrics.PostpartumRecord{
				PregnancyID: pregID,
				PatientID:   patient.ID,
				VisitDate:   now,
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			recID = p.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPostpartumRepoPG(globalDB.Pool)
			return repo.Delete(ctx, recID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := obstetrics.NewPostpartumRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, recID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted postpartum record")
		}
	})
}
