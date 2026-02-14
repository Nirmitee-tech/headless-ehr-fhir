package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/clinical"
	"github.com/google/uuid"
)

func TestConditionCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("cond")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "CondPatient", "Test", "MRN-COND-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "CondDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *clinical.Condition
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewConditionRepoPG(globalDB.Pool)
			c := &clinical.Condition{
				PatientID:      patient.ID,
				RecorderID:     &practitioner.ID,
				ClinicalStatus: "active",
				CategoryCode:   ptrStr("encounter-diagnosis"),
				SeverityCode:   ptrStr("moderate"),
				SeverityDisplay: ptrStr("Moderate"),
				CodeSystem:     ptrStr("http://snomed.info/sct"),
				CodeValue:      "73211009",
				CodeDisplay:    "Diabetes mellitus",
				OnsetDatetime:  &now,
				Note:           ptrStr("Type 2 diabetes, newly diagnosed"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			created = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create condition: %v", err)
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
			repo := clinical.NewConditionRepoPG(globalDB.Pool)
			fakePatient := uuid.New()
			c := &clinical.Condition{
				PatientID:      fakePatient,
				ClinicalStatus: "active",
				CodeValue:      "38341003",
				CodeDisplay:    "Hypertension",
			}
			return repo.Create(ctx, c)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		cond := createTestCondition(t, ctx, globalDB.Pool, tenantID, patient.ID)

		var fetched *clinical.Condition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewConditionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, cond.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CodeValue != "38341003" {
			t.Errorf("expected code=38341003, got %s", fetched.CodeValue)
		}
		if fetched.ClinicalStatus != "active" {
			t.Errorf("expected status=active, got %s", fetched.ClinicalStatus)
		}
	})

	t.Run("Update", func(t *testing.T) {
		cond := createTestCondition(t, ctx, globalDB.Pool, tenantID, patient.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewConditionRepoPG(globalDB.Pool)
			cond.ClinicalStatus = "resolved"
			now := time.Now()
			cond.AbatementDatetime = &now
			cond.Note = ptrStr("Resolved with treatment")
			return repo.Update(ctx, cond)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *clinical.Condition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewConditionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, cond.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.ClinicalStatus != "resolved" {
			t.Errorf("expected status=resolved, got %s", fetched.ClinicalStatus)
		}
		if fetched.AbatementDatetime == nil {
			t.Error("expected non-nil AbatementDatetime")
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		// Create multiple conditions
		createTestCondition(t, ctx, globalDB.Pool, tenantID, patient.ID)

		var results []*clinical.Condition
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewConditionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 condition")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByCode", func(t *testing.T) {
		var results []*clinical.Condition
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewConditionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient":         patient.ID.String(),
				"code":            "38341003",
				"clinical-status": "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by code: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for code=38341003")
		}
		for _, r := range results {
			if r.CodeValue != "38341003" {
				t.Errorf("expected code=38341003, got %s", r.CodeValue)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		cond := createTestCondition(t, ctx, globalDB.Pool, tenantID, patient.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewConditionRepoPG(globalDB.Pool)
			return repo.Delete(ctx, cond.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewConditionRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, cond.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted condition")
		}
	})
}

func TestObservationCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("obs")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "ObsPatient", "Test", "MRN-OBS-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "ObsDoc", "Smith")

	t.Run("Create_VitalSign", func(t *testing.T) {
		var created *clinical.Observation
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			o := &clinical.Observation{
				Status:            "final",
				CategoryCode:      ptrStr("vital-signs"),
				CategoryDisplay:   ptrStr("Vital Signs"),
				CodeSystem:        ptrStr("http://loinc.org"),
				CodeValue:         "8867-4",
				CodeDisplay:       "Heart rate",
				PatientID:         patient.ID,
				PerformerID:       &practitioner.ID,
				EffectiveDatetime: &now,
				ValueQuantity:     ptrFloat(72),
				ValueUnit:         ptrStr("beats/minute"),
			}
			if err := repo.Create(ctx, o); err != nil {
				return err
			}
			created = o
			return nil
		})
		if err != nil {
			t.Fatalf("Create observation: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_WithComponents", func(t *testing.T) {
		now := time.Now()
		var obsID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			// Blood pressure with systolic/diastolic components
			o := &clinical.Observation{
				Status:            "final",
				CategoryCode:      ptrStr("vital-signs"),
				CategoryDisplay:   ptrStr("Vital Signs"),
				CodeSystem:        ptrStr("http://loinc.org"),
				CodeValue:         "85354-9",
				CodeDisplay:       "Blood pressure panel",
				PatientID:         patient.ID,
				PerformerID:       &practitioner.ID,
				EffectiveDatetime: &now,
			}
			if err := repo.Create(ctx, o); err != nil {
				return err
			}
			obsID = o.ID

			// Systolic component
			systolic := &clinical.ObservationComponent{
				ObservationID: o.ID,
				CodeSystem:    ptrStr("http://loinc.org"),
				CodeValue:     "8480-6",
				CodeDisplay:   "Systolic blood pressure",
				ValueQuantity: ptrFloat(120),
				ValueUnit:     ptrStr("mmHg"),
			}
			if err := repo.AddComponent(ctx, systolic); err != nil {
				return err
			}

			// Diastolic component
			diastolic := &clinical.ObservationComponent{
				ObservationID: o.ID,
				CodeSystem:    ptrStr("http://loinc.org"),
				CodeValue:     "8462-4",
				CodeDisplay:   "Diastolic blood pressure",
				ValueQuantity: ptrFloat(80),
				ValueUnit:     ptrStr("mmHg"),
			}
			return repo.AddComponent(ctx, diastolic)
		})
		if err != nil {
			t.Fatalf("Create observation with components: %v", err)
		}

		// Verify components
		var comps []*clinical.ObservationComponent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			var err error
			comps, err = repo.GetComponents(ctx, obsID)
			return err
		})
		if err != nil {
			t.Fatalf("GetComponents: %v", err)
		}
		if len(comps) != 2 {
			t.Fatalf("expected 2 components, got %d", len(comps))
		}
	})

	t.Run("Search_ByCategory", func(t *testing.T) {
		var results []*clinical.Observation
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient":  patient.ID.String(),
				"category": "vital-signs",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by category: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 vital-signs observation")
		}
		for _, r := range results {
			if r.CategoryCode == nil || *r.CategoryCode != "vital-signs" {
				t.Errorf("expected category=vital-signs, got %v", r.CategoryCode)
			}
		}
	})

	t.Run("Search_ByCode", func(t *testing.T) {
		var results []*clinical.Observation
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"code":    "8867-4",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by code: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for code=8867-4")
		}
		for _, r := range results {
			if r.CodeValue != "8867-4" {
				t.Errorf("expected code_value=8867-4, got %s", r.CodeValue)
			}
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var obs *clinical.Observation
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			o := &clinical.Observation{
				Status:            "preliminary",
				CodeValue:         "2339-0",
				CodeDisplay:       "Glucose",
				PatientID:         patient.ID,
				EffectiveDatetime: &now,
				ValueQuantity:     ptrFloat(100),
				ValueUnit:         ptrStr("mg/dL"),
			}
			if err := repo.Create(ctx, o); err != nil {
				return err
			}
			obs = o
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		// Update status to final
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			obs.Status = "final"
			obs.InterpretationCode = ptrStr("N")
			obs.InterpretationDisplay = ptrStr("Normal")
			return repo.Update(ctx, obs)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		// Verify
		var fetched *clinical.Observation
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, obs.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "final" {
			t.Errorf("expected status=final, got %s", fetched.Status)
		}
		if fetched.InterpretationCode == nil || *fetched.InterpretationCode != "N" {
			t.Errorf("expected interpretation=N, got %v", fetched.InterpretationCode)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var obs *clinical.Observation
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			o := &clinical.Observation{
				Status:            "final",
				CodeValue:         "29463-7",
				CodeDisplay:       "Body weight",
				PatientID:         patient.ID,
				EffectiveDatetime: &now,
				ValueQuantity:     ptrFloat(75.5),
				ValueUnit:         ptrStr("kg"),
			}
			if err := repo.Create(ctx, o); err != nil {
				return err
			}
			obs = o
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			return repo.Delete(ctx, obs.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewObservationRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, obs.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted observation")
		}
	})
}

func TestAllergyIntoleranceCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("allergy")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "AllergyPatient", "Test", "MRN-ALLERGY-001")

	t.Run("Create", func(t *testing.T) {
		var created *clinical.AllergyIntolerance
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			a := &clinical.AllergyIntolerance{
				PatientID:      patient.ID,
				ClinicalStatus: ptrStr("active"),
				Type:           ptrStr("allergy"),
				Category:       []string{"medication"},
				Criticality:    ptrStr("high"),
				CodeSystem:     ptrStr("http://snomed.info/sct"),
				CodeValue:      ptrStr("91936005"),
				CodeDisplay:    ptrStr("Penicillin allergy"),
				Note:           ptrStr("Severe reaction documented"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create allergy: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			a := &clinical.AllergyIntolerance{
				PatientID:      uuid.New(), // non-existent
				ClinicalStatus: ptrStr("active"),
				CodeValue:      ptrStr("test"),
				CodeDisplay:    ptrStr("Test"),
			}
			return repo.Create(ctx, a)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var allergyID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			a := &clinical.AllergyIntolerance{
				PatientID:      patient.ID,
				ClinicalStatus: ptrStr("active"),
				Type:           ptrStr("intolerance"),
				Category:       []string{"food"},
				Criticality:    ptrStr("low"),
				CodeValue:      ptrStr("102263004"),
				CodeDisplay:    ptrStr("Eggs allergy"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			allergyID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *clinical.AllergyIntolerance
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, allergyID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CodeValue == nil || *fetched.CodeValue != "102263004" {
			t.Errorf("expected code=102263004, got %v", fetched.CodeValue)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var allergy *clinical.AllergyIntolerance
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			a := &clinical.AllergyIntolerance{
				PatientID:      patient.ID,
				ClinicalStatus: ptrStr("active"),
				CodeValue:      ptrStr("111111"),
				CodeDisplay:    ptrStr("Test Allergy"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			allergy = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			allergy.ClinicalStatus = ptrStr("resolved")
			allergy.Note = ptrStr("Patient no longer exhibits symptoms")
			return repo.Update(ctx, allergy)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *clinical.AllergyIntolerance
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, allergy.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.ClinicalStatus == nil || *fetched.ClinicalStatus != "resolved" {
			t.Errorf("expected status=resolved, got %v", fetched.ClinicalStatus)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*clinical.AllergyIntolerance
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 allergy")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Reactions", func(t *testing.T) {
		var allergyID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			a := &clinical.AllergyIntolerance{
				PatientID:      patient.ID,
				ClinicalStatus: ptrStr("active"),
				CodeValue:      ptrStr("rxn-test"),
				CodeDisplay:    ptrStr("Reaction Test Allergy"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			allergyID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create allergy for reactions: %v", err)
		}

		// Add reaction
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			rx := &clinical.AllergyReaction{
				AllergyID:            allergyID,
				ManifestationCode:    "39579001",
				ManifestationDisplay: "Anaphylaxis",
				Severity:             ptrStr("severe"),
				Description:          ptrStr("Severe anaphylactic reaction"),
			}
			return repo.AddReaction(ctx, rx)
		})
		if err != nil {
			t.Fatalf("AddReaction: %v", err)
		}

		// Get reactions
		var reactions []*clinical.AllergyReaction
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			var err error
			reactions, err = repo.GetReactions(ctx, allergyID)
			return err
		})
		if err != nil {
			t.Fatalf("GetReactions: %v", err)
		}
		if len(reactions) != 1 {
			t.Fatalf("expected 1 reaction, got %d", len(reactions))
		}
		if reactions[0].ManifestationCode != "39579001" {
			t.Errorf("expected manifestation=39579001, got %s", reactions[0].ManifestationCode)
		}
		if reactions[0].Severity == nil || *reactions[0].Severity != "severe" {
			t.Errorf("expected severity=severe, got %v", reactions[0].Severity)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var allergyID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			a := &clinical.AllergyIntolerance{
				PatientID:      patient.ID,
				ClinicalStatus: ptrStr("active"),
				CodeValue:      ptrStr("delete-test"),
				CodeDisplay:    ptrStr("Delete Test"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			allergyID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			return repo.Delete(ctx, allergyID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := clinical.NewAllergyRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, allergyID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted allergy")
		}
	})
}
