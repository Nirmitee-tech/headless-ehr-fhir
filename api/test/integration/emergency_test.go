package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/emergency"
	"github.com/google/uuid"
)

func TestTriageCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("triage")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "TriagePatient", "Test", "MRN-TRIAGE-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "TriageNurse", "Jones")
	enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

	t.Run("Create", func(t *testing.T) {
		var created *emergency.TriageRecord
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			tr := &emergency.TriageRecord{
				PatientID:        patient.ID,
				EncounterID:      enc.ID,
				TriageNurseID:    practitioner.ID,
				ArrivalTime:      &now,
				TriageTime:       &now,
				ChiefComplaint:   "Chest pain, sudden onset",
				AcuityLevel:      ptrInt(2),
				AcuitySystem:     ptrStr("ESI"),
				PainScale:        ptrInt(8),
				ArrivalMode:      ptrStr("ambulance"),
				HeartRate:        ptrInt(110),
				BloodPressureSys: ptrInt(150),
				BloodPressureDia: ptrInt(95),
				Temperature:      ptrFloat(37.2),
				RespiratoryRate:  ptrInt(22),
				OxygenSaturation: ptrInt(96),
				GlasgowComaScore: ptrInt(15),
				AllergyNote:      ptrStr("NKDA"),
				MedicationNote:   ptrStr("Aspirin 81mg daily"),
				Note:             ptrStr("Patient alert and oriented, diaphoretic"),
			}
			if err := repo.Create(ctx, tr); err != nil {
				return err
			}
			created = tr
			return nil
		})
		if err != nil {
			t.Fatalf("Create triage: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			tr := &emergency.TriageRecord{
				PatientID:      uuid.New(),
				EncounterID:    enc.ID,
				TriageNurseID:  practitioner.ID,
				ChiefComplaint: "Test",
			}
			return repo.Create(ctx, tr)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var triageID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			tr := &emergency.TriageRecord{
				PatientID:      patient.ID,
				EncounterID:    enc.ID,
				TriageNurseID:  practitioner.ID,
				ArrivalTime:    &now,
				ChiefComplaint: "Abdominal pain",
				AcuityLevel:    ptrInt(3),
			}
			if err := repo.Create(ctx, tr); err != nil {
				return err
			}
			triageID = tr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *emergency.TriageRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, triageID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.ChiefComplaint != "Abdominal pain" {
			t.Errorf("expected chief_complaint='Abdominal pain', got %s", fetched.ChiefComplaint)
		}
		if fetched.AcuityLevel == nil || *fetched.AcuityLevel != 3 {
			t.Errorf("expected acuity_level=3, got %v", fetched.AcuityLevel)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var triage *emergency.TriageRecord
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			tr := &emergency.TriageRecord{
				PatientID:      patient.ID,
				EncounterID:    enc.ID,
				TriageNurseID:  practitioner.ID,
				ArrivalTime:    &now,
				ChiefComplaint: "Shortness of breath",
				AcuityLevel:    ptrInt(3),
				PainScale:      ptrInt(4),
			}
			if err := repo.Create(ctx, tr); err != nil {
				return err
			}
			triage = tr
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			triage.AcuityLevel = ptrInt(2)
			triage.PainScale = ptrInt(7)
			triage.OxygenSaturation = ptrInt(90)
			triage.Note = ptrStr("Condition worsening, acuity upgraded")
			return repo.Update(ctx, triage)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *emergency.TriageRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, triage.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.AcuityLevel == nil || *fetched.AcuityLevel != 2 {
			t.Errorf("expected acuity_level=2, got %v", fetched.AcuityLevel)
		}
		if fetched.OxygenSaturation == nil || *fetched.OxygenSaturation != 90 {
			t.Errorf("expected oxygen_saturation=90, got %v", fetched.OxygenSaturation)
		}
		if fetched.Note == nil || *fetched.Note != "Condition worsening, acuity upgraded" {
			t.Errorf("expected updated note, got %v", fetched.Note)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*emergency.TriageRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 triage record")
		}
		_ = results
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*emergency.TriageRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 triage record for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*emergency.TriageRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient_id":  patient.ID.String(),
				"acuity_level": "2",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var triageID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			tr := &emergency.TriageRecord{
				PatientID:      patient.ID,
				EncounterID:    enc.ID,
				TriageNurseID:  practitioner.ID,
				ArrivalTime:    &now,
				ChiefComplaint: "Delete test",
			}
			if err := repo.Create(ctx, tr); err != nil {
				return err
			}
			triageID = tr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			return repo.Delete(ctx, triageID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTriageRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, triageID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted triage record")
		}
	})
}

func TestEDTrackingCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("edtrack")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "EDPatient", "Test", "MRN-ED-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "EDDoc", "Smith")
	nurse := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "EDNurse", "Williams")
	enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

	t.Run("Create", func(t *testing.T) {
		var created *emergency.EDTracking
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			ed := &emergency.EDTracking{
				PatientID:     patient.ID,
				EncounterID:   enc.ID,
				CurrentStatus: "waiting",
				BedAssignment: ptrStr("Bay 5"),
				AttendingID:   &practitioner.ID,
				NurseID:       &nurse.ID,
				ArrivalTime:   &now,
				Note:          ptrStr("Patient waiting for assessment"),
			}
			if err := repo.Create(ctx, ed); err != nil {
				return err
			}
			created = ed
			return nil
		})
		if err != nil {
			t.Fatalf("Create ED tracking: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			ed := &emergency.EDTracking{
				PatientID:     uuid.New(),
				EncounterID:   enc.ID,
				CurrentStatus: "waiting",
			}
			return repo.Create(ctx, ed)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var trackID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			ed := &emergency.EDTracking{
				PatientID:     patient.ID,
				EncounterID:   enc.ID,
				CurrentStatus: "in-treatment",
				BedAssignment: ptrStr("Room 3"),
				ArrivalTime:   &now,
			}
			if err := repo.Create(ctx, ed); err != nil {
				return err
			}
			trackID = ed.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *emergency.EDTracking
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, trackID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.CurrentStatus != "in-treatment" {
			t.Errorf("expected current_status=in-treatment, got %s", fetched.CurrentStatus)
		}
		if fetched.BedAssignment == nil || *fetched.BedAssignment != "Room 3" {
			t.Errorf("expected bed_assignment='Room 3', got %v", fetched.BedAssignment)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var tracking *emergency.EDTracking
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			ed := &emergency.EDTracking{
				PatientID:     patient.ID,
				EncounterID:   enc.ID,
				CurrentStatus: "in-treatment",
				BedAssignment: ptrStr("Bay 7"),
				ArrivalTime:   &now,
			}
			if err := repo.Create(ctx, ed); err != nil {
				return err
			}
			tracking = ed
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		dischTime := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			tracking.CurrentStatus = "discharged"
			tracking.DischargeTime = &dischTime
			tracking.Disposition = ptrStr("home")
			tracking.LengthOfStayMins = ptrInt(180)
			tracking.Note = ptrStr("Discharged with follow-up instructions")
			return repo.Update(ctx, tracking)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *emergency.EDTracking
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, tracking.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.CurrentStatus != "discharged" {
			t.Errorf("expected current_status=discharged, got %s", fetched.CurrentStatus)
		}
		if fetched.Disposition == nil || *fetched.Disposition != "home" {
			t.Errorf("expected disposition=home, got %v", fetched.Disposition)
		}
		if fetched.LengthOfStayMins == nil || *fetched.LengthOfStayMins != 180 {
			t.Errorf("expected length_of_stay_mins=180, got %v", fetched.LengthOfStayMins)
		}
		if fetched.DischargeTime == nil {
			t.Error("expected non-nil DischargeTime")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*emergency.EDTracking
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 ED tracking record")
		}
		_ = results
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*emergency.EDTracking
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 ED tracking record for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*emergency.EDTracking
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient_id":     patient.ID.String(),
				"current_status": "in-treatment",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.CurrentStatus != "in-treatment" {
				t.Errorf("expected current_status=in-treatment, got %s", r.CurrentStatus)
			}
		}
	})

	t.Run("StatusHistory", func(t *testing.T) {
		var trackID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			ed := &emergency.EDTracking{
				PatientID:     patient.ID,
				EncounterID:   enc.ID,
				CurrentStatus: "waiting",
				ArrivalTime:   &now,
			}
			if err := repo.Create(ctx, ed); err != nil {
				return err
			}
			trackID = ed.ID

			// Add status history entries
			h1 := &emergency.EDStatusHistory{
				EDTrackingID: trackID,
				Status:       "registered",
				ChangedAt:    now.Add(-30 * time.Minute),
				Note:         ptrStr("Patient registered at front desk"),
			}
			if err := repo.AddStatusHistory(ctx, h1); err != nil {
				return err
			}

			h2 := &emergency.EDStatusHistory{
				EDTrackingID: trackID,
				Status:       "waiting",
				ChangedAt:    now.Add(-20 * time.Minute),
				ChangedBy:    &practitioner.ID,
				Note:         ptrStr("Triaged and waiting for bed"),
			}
			if err := repo.AddStatusHistory(ctx, h2); err != nil {
				return err
			}

			h3 := &emergency.EDStatusHistory{
				EDTrackingID: trackID,
				Status:       "in-treatment",
				ChangedAt:    now,
				ChangedBy:    &practitioner.ID,
				Note:         ptrStr("Moved to treatment bay"),
			}
			return repo.AddStatusHistory(ctx, h3)
		})
		if err != nil {
			t.Fatalf("Create tracking with status history: %v", err)
		}

		var history []*emergency.EDStatusHistory
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			var err error
			history, err = repo.GetStatusHistory(ctx, trackID)
			return err
		})
		if err != nil {
			t.Fatalf("GetStatusHistory: %v", err)
		}
		if len(history) != 3 {
			t.Fatalf("expected 3 status history entries, got %d", len(history))
		}
		if history[0].Status != "registered" {
			t.Errorf("expected first status=registered, got %s", history[0].Status)
		}
		if history[1].Status != "waiting" {
			t.Errorf("expected second status=waiting, got %s", history[1].Status)
		}
		if history[2].Status != "in-treatment" {
			t.Errorf("expected third status=in-treatment, got %s", history[2].Status)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var trackID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			ed := &emergency.EDTracking{
				PatientID:     patient.ID,
				EncounterID:   enc.ID,
				CurrentStatus: "waiting",
				ArrivalTime:   &now,
			}
			if err := repo.Create(ctx, ed); err != nil {
				return err
			}
			trackID = ed.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			return repo.Delete(ctx, trackID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewEDTrackingRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, trackID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted ED tracking record")
		}
	})
}

func TestTraumaActivationCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("trauma")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "TraumaPatient", "Test", "MRN-TRAUMA-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "TraumaDoc", "Smith")
	teamLead := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "TraumaLead", "Brown")
	enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

	t.Run("Create", func(t *testing.T) {
		var created *emergency.TraumaActivation
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			ta := &emergency.TraumaActivation{
				PatientID:         patient.ID,
				EncounterID:       &enc.ID,
				ActivationLevel:   "level-1",
				ActivationTime:    now,
				MechanismOfInjury: ptrStr("Motor vehicle collision at high speed"),
				ActivatedBy:       &practitioner.ID,
				TeamLeadID:        &teamLead.ID,
				Note:              ptrStr("Full trauma team activation"),
			}
			if err := repo.Create(ctx, ta); err != nil {
				return err
			}
			created = ta
			return nil
		})
		if err != nil {
			t.Fatalf("Create trauma activation: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			ta := &emergency.TraumaActivation{
				PatientID:       uuid.New(),
				ActivationLevel: "level-2",
				ActivationTime:  now,
			}
			return repo.Create(ctx, ta)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var traumaID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			ta := &emergency.TraumaActivation{
				PatientID:         patient.ID,
				EncounterID:       &enc.ID,
				ActivationLevel:   "level-2",
				ActivationTime:    now,
				MechanismOfInjury: ptrStr("Fall from height"),
			}
			if err := repo.Create(ctx, ta); err != nil {
				return err
			}
			traumaID = ta.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *emergency.TraumaActivation
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, traumaID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.ActivationLevel != "level-2" {
			t.Errorf("expected activation_level=level-2, got %s", fetched.ActivationLevel)
		}
		if fetched.MechanismOfInjury == nil || *fetched.MechanismOfInjury != "Fall from height" {
			t.Errorf("expected mechanism_of_injury='Fall from height', got %v", fetched.MechanismOfInjury)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var trauma *emergency.TraumaActivation
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			ta := &emergency.TraumaActivation{
				PatientID:         patient.ID,
				EncounterID:       &enc.ID,
				ActivationLevel:   "level-1",
				ActivationTime:    now,
				MechanismOfInjury: ptrStr("Penetrating injury"),
				ActivatedBy:       &practitioner.ID,
			}
			if err := repo.Create(ctx, ta); err != nil {
				return err
			}
			trauma = ta
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		deactTime := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			trauma.DeactivationTime = &deactTime
			trauma.Outcome = ptrStr("admitted-icu")
			trauma.TeamLeadID = &teamLead.ID
			trauma.Note = ptrStr("Patient stabilized, transferred to ICU")
			return repo.Update(ctx, trauma)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *emergency.TraumaActivation
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, trauma.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Outcome == nil || *fetched.Outcome != "admitted-icu" {
			t.Errorf("expected outcome=admitted-icu, got %v", fetched.Outcome)
		}
		if fetched.DeactivationTime == nil {
			t.Error("expected non-nil DeactivationTime")
		}
		if fetched.TeamLeadID == nil || *fetched.TeamLeadID != teamLead.ID {
			t.Errorf("expected team_lead_id=%s, got %v", teamLead.ID, fetched.TeamLeadID)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*emergency.TraumaActivation
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 trauma activation")
		}
		_ = results
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*emergency.TraumaActivation
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 trauma activation for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*emergency.TraumaActivation
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient_id":       patient.ID.String(),
				"activation_level": "level-1",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.ActivationLevel != "level-1" {
				t.Errorf("expected activation_level=level-1, got %s", r.ActivationLevel)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var traumaID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			ta := &emergency.TraumaActivation{
				PatientID:       patient.ID,
				EncounterID:     &enc.ID,
				ActivationLevel: "level-2",
				ActivationTime:  now,
			}
			if err := repo.Create(ctx, ta); err != nil {
				return err
			}
			traumaID = ta.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			return repo.Delete(ctx, traumaID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := emergency.NewTraumaRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, traumaID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted trauma activation")
		}
	})
}
