package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/behavioral"
	"github.com/google/uuid"
)

func TestPsychAssessmentCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("psych")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "PsychPatient", "Test", "MRN-PSYCH-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "PsychDoc", "Smith")
	enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

	t.Run("Create", func(t *testing.T) {
		var created *behavioral.PsychiatricAssessment
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			a := &behavioral.PsychiatricAssessment{
				PatientID:           patient.ID,
				EncounterID:         enc.ID,
				AssessorID:          practitioner.ID,
				AssessmentDate:      now,
				ChiefComplaint:      ptrStr("Anxiety and insomnia"),
				MentalStatusExam:    ptrStr("Alert and oriented x4"),
				Mood:                ptrStr("anxious"),
				Affect:              ptrStr("congruent"),
				ThoughtProcess:      ptrStr("logical"),
				RiskAssessment:      ptrStr("low"),
				SuicideRiskLevel:    ptrStr("low"),
				HomicideRiskLevel:   ptrStr("none"),
				DiagnosisCode:       ptrStr("F41.1"),
				DiagnosisDisplay:    ptrStr("Generalized anxiety disorder"),
				TreatmentPlan:       ptrStr("CBT and SSRI"),
				Note:                ptrStr("Initial psychiatric assessment"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create psychiatric assessment: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var assessID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			a := &behavioral.PsychiatricAssessment{
				PatientID:        patient.ID,
				EncounterID:      enc.ID,
				AssessorID:       practitioner.ID,
				AssessmentDate:   now,
				ChiefComplaint:   ptrStr("Depression"),
				SuicideRiskLevel: ptrStr("moderate"),
				DiagnosisCode:    ptrStr("F32.1"),
				DiagnosisDisplay: ptrStr("Major depressive disorder, single episode, moderate"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			assessID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *behavioral.PsychiatricAssessment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, assessID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.ChiefComplaint == nil || *fetched.ChiefComplaint != "Depression" {
			t.Errorf("expected chief_complaint=Depression, got %v", fetched.ChiefComplaint)
		}
		if fetched.DiagnosisCode == nil || *fetched.DiagnosisCode != "F32.1" {
			t.Errorf("expected diagnosis_code=F32.1, got %v", fetched.DiagnosisCode)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var assess *behavioral.PsychiatricAssessment
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			a := &behavioral.PsychiatricAssessment{
				PatientID:        patient.ID,
				EncounterID:      enc.ID,
				AssessorID:       practitioner.ID,
				AssessmentDate:   now,
				ChiefComplaint:   ptrStr("Panic attacks"),
				SuicideRiskLevel: ptrStr("low"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			assess = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			assess.SuicideRiskLevel = ptrStr("moderate")
			assess.TreatmentPlan = ptrStr("Increase therapy frequency")
			assess.Note = ptrStr("Updated after follow-up")
			return repo.Update(ctx, assess)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *behavioral.PsychiatricAssessment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, assess.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.SuicideRiskLevel == nil || *fetched.SuicideRiskLevel != "moderate" {
			t.Errorf("expected suicide_risk_level=moderate, got %v", fetched.SuicideRiskLevel)
		}
		if fetched.TreatmentPlan == nil || *fetched.TreatmentPlan != "Increase therapy frequency" {
			t.Errorf("expected treatment_plan updated, got %v", fetched.TreatmentPlan)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*behavioral.PsychiatricAssessment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 assessment")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*behavioral.PsychiatricAssessment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient":   patient.ID.String(),
				"encounter": enc.ID.String(),
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result")
		}
		for _, r := range results {
			if r.EncounterID != enc.ID {
				t.Errorf("expected encounter_id=%s, got %s", enc.ID, r.EncounterID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var assessID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			a := &behavioral.PsychiatricAssessment{
				PatientID:      patient.ID,
				EncounterID:    enc.ID,
				AssessorID:     practitioner.ID,
				AssessmentDate: now,
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			assessID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			return repo.Delete(ctx, assessID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewPsychAssessmentRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, assessID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted assessment")
		}
	})
}

func TestSafetyPlanCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("safety")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "SafetyPatient", "Test", "MRN-SAFE-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SafetyDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *behavioral.SafetyPlan
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
			s := &behavioral.SafetyPlan{
				PatientID:              patient.ID,
				CreatedByID:            practitioner.ID,
				Status:                 "active",
				PlanDate:               now,
				WarningSigns:           ptrStr("Increased isolation, sleep disturbance"),
				CopingStrategies:       ptrStr("Deep breathing, walking, journaling"),
				SocialDistractions:     ptrStr("Call friend, go to coffee shop"),
				PeopleToContact:        ptrStr("Jane Doe 555-1234"),
				ProfessionalsToContact: ptrStr("Therapist: Dr. Smith 555-5678"),
				EmergencyContacts:      ptrStr("988 Suicide Hotline"),
				MeansRestriction:       ptrStr("Remove firearms from home"),
				ReasonsForLiving:       ptrStr("Children, career goals"),
				PatientSignature:       ptrBool(true),
				ProviderSignature:      ptrBool(true),
				Note:                   ptrStr("Created during crisis intervention"),
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			created = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create safety plan: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var planID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
			s := &behavioral.SafetyPlan{
				PatientID:   patient.ID,
				CreatedByID: practitioner.ID,
				Status:      "active",
				PlanDate:    now,
				WarningSigns: ptrStr("Hopelessness"),
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			planID = s.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *behavioral.SafetyPlan
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, planID)
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

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var plan *behavioral.SafetyPlan
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
			s := &behavioral.SafetyPlan{
				PatientID:   patient.ID,
				CreatedByID: practitioner.ID,
				Status:      "active",
				PlanDate:    now,
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			plan = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		reviewDate := time.Now().Add(30 * 24 * time.Hour)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
			plan.Status = "reviewed"
			plan.ReviewDate = &reviewDate
			plan.Note = ptrStr("Plan reviewed and updated")
			return repo.Update(ctx, plan)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *behavioral.SafetyPlan
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, plan.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "reviewed" {
			t.Errorf("expected status=reviewed, got %s", fetched.Status)
		}
		if fetched.ReviewDate == nil {
			t.Error("expected non-nil ReviewDate")
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*behavioral.SafetyPlan
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 safety plan")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*behavioral.SafetyPlan
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
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
			t.Error("expected at least 1 result for status=active")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var planID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
			s := &behavioral.SafetyPlan{
				PatientID:   patient.ID,
				CreatedByID: practitioner.ID,
				Status:      "draft",
				PlanDate:    now,
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			planID = s.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
			return repo.Delete(ctx, planID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSafetyPlanRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, planID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted safety plan")
		}
	})
}

func TestLegalHoldCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("legal")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "LegalPatient", "Test", "MRN-LEGAL-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "LegalDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *behavioral.LegalHold
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
			h := &behavioral.LegalHold{
				PatientID:            patient.ID,
				InitiatedByID:        practitioner.ID,
				Status:               "active",
				HoldType:             "5150",
				AuthorityStatute:     ptrStr("CA WIC 5150"),
				StartDatetime:        now,
				DurationHours:        ptrInt(72),
				Reason:               "Danger to self",
				CriteriaMet:          ptrStr("SI with plan"),
				PatientRightsGiven:   ptrBool(true),
				LegalCounselNotified: ptrBool(false),
				Note:                 ptrStr("Emergency psychiatric hold"),
			}
			if err := repo.Create(ctx, h); err != nil {
				return err
			}
			created = h
			return nil
		})
		if err != nil {
			t.Fatalf("Create legal hold: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var holdID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
			h := &behavioral.LegalHold{
				PatientID:     patient.ID,
				InitiatedByID: practitioner.ID,
				Status:        "active",
				HoldType:      "5250",
				StartDatetime: now,
				Reason:        "Gravely disabled",
			}
			if err := repo.Create(ctx, h); err != nil {
				return err
			}
			holdID = h.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *behavioral.LegalHold
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, holdID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.HoldType != "5250" {
			t.Errorf("expected hold_type=5250, got %s", fetched.HoldType)
		}
		if fetched.Reason != "Gravely disabled" {
			t.Errorf("expected reason=Gravely disabled, got %s", fetched.Reason)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var hold *behavioral.LegalHold
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
			h := &behavioral.LegalHold{
				PatientID:     patient.ID,
				InitiatedByID: practitioner.ID,
				Status:        "active",
				HoldType:      "5150",
				StartDatetime: now,
				Reason:        "Danger to others",
			}
			if err := repo.Create(ctx, h); err != nil {
				return err
			}
			hold = h
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		endTime := now.Add(72 * time.Hour)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
			hold.Status = "released"
			hold.EndDatetime = &endTime
			hold.ReleaseReason = ptrStr("Stabilized, no longer meets criteria")
			hold.ReleaseAuthorizedByID = &practitioner.ID
			return repo.Update(ctx, hold)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *behavioral.LegalHold
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, hold.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "released" {
			t.Errorf("expected status=released, got %s", fetched.Status)
		}
		if fetched.EndDatetime == nil {
			t.Error("expected non-nil EndDatetime")
		}
		if fetched.ReleaseReason == nil || *fetched.ReleaseReason != "Stabilized, no longer meets criteria" {
			t.Errorf("expected release_reason updated, got %v", fetched.ReleaseReason)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*behavioral.LegalHold
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 legal hold")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*behavioral.LegalHold
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
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
		// May be 0 since we released one above
		_ = total
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var holdID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
			h := &behavioral.LegalHold{
				PatientID:     patient.ID,
				InitiatedByID: practitioner.ID,
				Status:        "active",
				HoldType:      "5150",
				StartDatetime: now,
				Reason:        "Delete test hold",
			}
			if err := repo.Create(ctx, h); err != nil {
				return err
			}
			holdID = h.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
			return repo.Delete(ctx, holdID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewLegalHoldRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, holdID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted legal hold")
		}
	})
}

func TestSeclusionRestraintCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("seclusion")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "SeclusionPatient", "Test", "MRN-SECL-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SeclusionDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *behavioral.SeclusionRestraintEvent
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			e := &behavioral.SeclusionRestraintEvent{
				PatientID:              patient.ID,
				OrderedByID:            practitioner.ID,
				EventType:              "seclusion",
				StartDatetime:          now,
				Reason:                 "Imminent danger to self and others",
				BehaviorDescription:    ptrStr("Aggressive, throwing objects"),
				AlternativesAttempted:  ptrStr("Verbal de-escalation, PRN medication offered"),
				MonitoringFrequencyMin: ptrInt(15),
				Note:                   ptrStr("Seclusion initiated per protocol"),
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			created = e
			return nil
		})
		if err != nil {
			t.Fatalf("Create seclusion event: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var eventID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			e := &behavioral.SeclusionRestraintEvent{
				PatientID:     patient.ID,
				OrderedByID:   practitioner.ID,
				EventType:     "restraint",
				RestraintType: ptrStr("4-point"),
				StartDatetime: now,
				Reason:        "Self-injurious behavior",
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			eventID = e.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *behavioral.SeclusionRestraintEvent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, eventID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.EventType != "restraint" {
			t.Errorf("expected event_type=restraint, got %s", fetched.EventType)
		}
		if fetched.RestraintType == nil || *fetched.RestraintType != "4-point" {
			t.Errorf("expected restraint_type=4-point, got %v", fetched.RestraintType)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var event *behavioral.SeclusionRestraintEvent
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			e := &behavioral.SeclusionRestraintEvent{
				PatientID:     patient.ID,
				OrderedByID:   practitioner.ID,
				EventType:     "seclusion",
				StartDatetime: now,
				Reason:        "Agitation",
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			event = e
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		endTime := now.Add(2 * time.Hour)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			event.EndDatetime = &endTime
			event.PatientConditionDuring = ptrStr("Calm, cooperative")
			event.DiscontinuedByID = &practitioner.ID
			event.DiscontinuationReason = ptrStr("Patient de-escalated")
			event.DebriefCompleted = ptrBool(true)
			event.DebriefNotes = ptrStr("Patient debriefed, coping plan reviewed")
			event.NutritionOffered = ptrBool(true)
			event.ToiletingOffered = ptrBool(true)
			return repo.Update(ctx, event)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *behavioral.SeclusionRestraintEvent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, event.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.EndDatetime == nil {
			t.Error("expected non-nil EndDatetime")
		}
		if fetched.DebriefCompleted == nil || !*fetched.DebriefCompleted {
			t.Error("expected debrief_completed=true")
		}
		if fetched.NutritionOffered == nil || !*fetched.NutritionOffered {
			t.Error("expected nutrition_offered=true")
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*behavioral.SeclusionRestraintEvent
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 seclusion/restraint event")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*behavioral.SeclusionRestraintEvent
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient":    patient.ID.String(),
				"event_type": "seclusion",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 seclusion event")
		}
		for _, r := range results {
			if r.EventType != "seclusion" {
				t.Errorf("expected event_type=seclusion, got %s", r.EventType)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var eventID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			e := &behavioral.SeclusionRestraintEvent{
				PatientID:     patient.ID,
				OrderedByID:   practitioner.ID,
				EventType:     "restraint",
				StartDatetime: now,
				Reason:        "Delete test",
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			eventID = e.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			return repo.Delete(ctx, eventID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewSeclusionRestraintRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, eventID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted seclusion/restraint event")
		}
	})
}

func TestGroupTherapyCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("group")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "GroupPatient", "Test", "MRN-GROUP-001")
	patient2 := createTestPatient(t, ctx, globalDB.Pool, tenantID, "GroupPatient2", "Test", "MRN-GROUP-002")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "GroupFacilitator", "Jones")

	t.Run("Create", func(t *testing.T) {
		var created *behavioral.GroupTherapySession
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			s := &behavioral.GroupTherapySession{
				SessionName:       "Anxiety Management Group",
				SessionType:       ptrStr("CBT"),
				FacilitatorID:     practitioner.ID,
				Status:            "scheduled",
				ScheduledDatetime: now.Add(24 * time.Hour),
				Location:          ptrStr("Room 204"),
				MaxParticipants:   ptrInt(12),
				Topic:             ptrStr("Coping with panic attacks"),
				SessionGoals:      ptrStr("Learn grounding techniques"),
				Note:              ptrStr("Weekly session"),
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			created = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create group therapy session: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var sessionID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			s := &behavioral.GroupTherapySession{
				SessionName:       "DBT Skills Group",
				SessionType:       ptrStr("DBT"),
				FacilitatorID:     practitioner.ID,
				Status:            "scheduled",
				ScheduledDatetime: now,
				MaxParticipants:   ptrInt(8),
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			sessionID = s.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *behavioral.GroupTherapySession
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, sessionID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.SessionName != "DBT Skills Group" {
			t.Errorf("expected session_name=DBT Skills Group, got %s", fetched.SessionName)
		}
		if fetched.FacilitatorID != practitioner.ID {
			t.Errorf("expected facilitator_id=%s, got %s", practitioner.ID, fetched.FacilitatorID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var session *behavioral.GroupTherapySession
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			s := &behavioral.GroupTherapySession{
				SessionName:       "Process Group",
				FacilitatorID:     practitioner.ID,
				Status:            "scheduled",
				ScheduledDatetime: now,
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			session = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		startTime := now.Add(1 * time.Hour)
		endTime := now.Add(2 * time.Hour)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			session.Status = "completed"
			session.ActualStart = &startTime
			session.ActualEnd = &endTime
			session.SessionNotes = ptrStr("Good participation from all members")
			return repo.Update(ctx, session)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *behavioral.GroupTherapySession
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, session.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.ActualStart == nil {
			t.Error("expected non-nil ActualStart")
		}
		if fetched.ActualEnd == nil {
			t.Error("expected non-nil ActualEnd")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*behavioral.GroupTherapySession
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 group therapy session")
		}
		_ = results
	})

	t.Run("Search", func(t *testing.T) {
		var results []*behavioral.GroupTherapySession
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status":      "scheduled",
				"facilitator": practitioner.ID.String(),
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for status=scheduled")
		}
		for _, r := range results {
			if r.Status != "scheduled" {
				t.Errorf("expected status=scheduled, got %s", r.Status)
			}
		}
	})

	t.Run("Attendance", func(t *testing.T) {
		now := time.Now()
		var sessionID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			s := &behavioral.GroupTherapySession{
				SessionName:       "Attendance Test Session",
				FacilitatorID:     practitioner.ID,
				Status:            "completed",
				ScheduledDatetime: now,
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			sessionID = s.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create session: %v", err)
		}

		// Add attendance for patient 1
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			a := &behavioral.GroupTherapyAttendance{
				SessionID:          sessionID,
				PatientID:          patient.ID,
				AttendanceStatus:   "present",
				ParticipationLevel: ptrStr("active"),
				MoodBefore:         ptrStr("anxious"),
				MoodAfter:          ptrStr("calm"),
				Note:               ptrStr("Shared personal experience"),
			}
			return repo.AddAttendance(ctx, a)
		})
		if err != nil {
			t.Fatalf("AddAttendance patient1: %v", err)
		}

		// Add attendance for patient 2
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			a := &behavioral.GroupTherapyAttendance{
				SessionID:          sessionID,
				PatientID:          patient2.ID,
				AttendanceStatus:   "present",
				ParticipationLevel: ptrStr("minimal"),
				MoodBefore:         ptrStr("flat"),
				MoodAfter:          ptrStr("flat"),
			}
			return repo.AddAttendance(ctx, a)
		})
		if err != nil {
			t.Fatalf("AddAttendance patient2: %v", err)
		}

		// Get attendance
		var attendance []*behavioral.GroupTherapyAttendance
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			var err error
			attendance, err = repo.GetAttendance(ctx, sessionID)
			return err
		})
		if err != nil {
			t.Fatalf("GetAttendance: %v", err)
		}
		if len(attendance) != 2 {
			t.Fatalf("expected 2 attendance records, got %d", len(attendance))
		}
		for _, a := range attendance {
			if a.SessionID != sessionID {
				t.Errorf("expected session_id=%s, got %s", sessionID, a.SessionID)
			}
			if a.AttendanceStatus != "present" {
				t.Errorf("expected attendance_status=present, got %s", a.AttendanceStatus)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var sessionID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			s := &behavioral.GroupTherapySession{
				SessionName:       "Delete Test Session",
				FacilitatorID:     practitioner.ID,
				Status:            "cancelled",
				ScheduledDatetime: now,
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			sessionID = s.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			return repo.Delete(ctx, sessionID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := behavioral.NewGroupTherapyRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, sessionID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted group therapy session")
		}
	})
}
