package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/medication"
	"github.com/google/uuid"
)

func TestMedicationRequestLifecycle(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("medreq")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	// Create prerequisite data
	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "MedPatient", "Test", "MRN-MED-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "MedDoc", "Smith")
	med := createTestMedication(t, ctx, globalDB.Pool, tenantID, "723", "Amoxicillin 500mg")

	t.Run("Create_Medication", func(t *testing.T) {
		var created *medication.Medication
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRepoPG(globalDB.Pool)
			m := &medication.Medication{
				CodeSystem:  ptrStr("http://www.nlm.nih.gov/research/umls/rxnorm"),
				CodeValue:   "197696",
				CodeDisplay: "Lisinopril 10 MG Oral Tablet",
				Status:      "active",
				FormCode:    ptrStr("tablet"),
				FormDisplay: ptrStr("Tablet"),
				Schedule:    ptrStr("OTC"),
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			created = m
			return nil
		})
		if err != nil {
			t.Fatalf("Create medication: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Medication_Search", func(t *testing.T) {
		var results []*medication.Medication
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{"status": "active"}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search medications: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 active medication")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Medication_Ingredients", func(t *testing.T) {
		// Add ingredient to existing medication
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRepoPG(globalDB.Pool)
			ing := &medication.MedicationIngredient{
				MedicationID:        med.ID,
				ItemCode:            ptrStr("723"),
				ItemDisplay:         "Amoxicillin",
				StrengthNumerator:   ptrFloat(500),
				StrengthNumeratorUnit: ptrStr("mg"),
				StrengthDenominator: ptrFloat(1),
				StrengthDenominatorUnit: ptrStr("tablet"),
				IsActive:            ptrBool(true),
			}
			return repo.AddIngredient(ctx, ing)
		})
		if err != nil {
			t.Fatalf("AddIngredient: %v", err)
		}

		var ings []*medication.MedicationIngredient
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRepoPG(globalDB.Pool)
			var err error
			ings, err = repo.GetIngredients(ctx, med.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetIngredients: %v", err)
		}
		if len(ings) != 1 {
			t.Fatalf("expected 1 ingredient, got %d", len(ings))
		}
		if ings[0].ItemDisplay != "Amoxicillin" {
			t.Errorf("expected item=Amoxicillin, got %s", ings[0].ItemDisplay)
		}
	})

	t.Run("Create_MedicationRequest", func(t *testing.T) {
		now := time.Now()
		var created *medication.MedicationRequest
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			mr := &medication.MedicationRequest{
				Status:              "active",
				Intent:              "order",
				CategoryCode:        ptrStr("outpatient"),
				Priority:            ptrStr("routine"),
				MedicationID:        med.ID,
				PatientID:           patient.ID,
				RequesterID:         practitioner.ID,
				DosageText:          ptrStr("Take 1 capsule three times daily"),
				DosageTimingCode:    ptrStr("TID"),
				DosageRouteCode:     ptrStr("PO"),
				DosageRouteDisplay:  ptrStr("Oral"),
				DoseQuantity:        ptrFloat(500),
				DoseUnit:            ptrStr("mg"),
				QuantityValue:       ptrFloat(30),
				QuantityUnit:        ptrStr("capsule"),
				DaysSupply:          ptrInt(10),
				RefillsAllowed:      ptrInt(2),
				SubstitutionAllowed: ptrBool(true),
				AuthoredOn:          &now,
				Note:                ptrStr("For bacterial infection"),
			}
			if err := repo.Create(ctx, mr); err != nil {
				return err
			}
			created = mr
			return nil
		})
		if err != nil {
			t.Fatalf("Create medication request: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.Status != "active" {
			t.Errorf("expected status=active, got %s", created.Status)
		}
	})

	t.Run("MedicationRequest_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			mr := &medication.MedicationRequest{
				Status:       "active",
				Intent:       "order",
				MedicationID: uuid.New(), // non-existent
				PatientID:    patient.ID,
				RequesterID:  practitioner.ID,
			}
			return repo.Create(ctx, mr)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent medication")
		}
	})

	t.Run("MedicationRequest_GetByID_and_Update", func(t *testing.T) {
		now := time.Now()
		var reqID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			mr := &medication.MedicationRequest{
				Status:       "active",
				Intent:       "order",
				MedicationID: med.ID,
				PatientID:    patient.ID,
				RequesterID:  practitioner.ID,
				DosageText:   ptrStr("1 tablet daily"),
				DoseQuantity: ptrFloat(10),
				DoseUnit:     ptrStr("mg"),
				AuthoredOn:   &now,
			}
			if err := repo.Create(ctx, mr); err != nil {
				return err
			}
			reqID = mr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		// Get
		var fetched *medication.MedicationRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, reqID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.MedicationID != med.ID {
			t.Errorf("expected medication_id=%s, got %s", med.ID, fetched.MedicationID)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
		if fetched.RequesterID != practitioner.ID {
			t.Errorf("expected requester_id=%s, got %s", practitioner.ID, fetched.RequesterID)
		}

		// Update
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			fetched.Status = "completed"
			fetched.Note = ptrStr("Course completed")
			return repo.Update(ctx, fetched)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		// Verify
		var updated *medication.MedicationRequest
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			var err error
			updated, err = repo.GetByID(ctx, reqID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if updated.Status != "completed" {
			t.Errorf("expected status=completed, got %s", updated.Status)
		}
		if updated.Note == nil || *updated.Note != "Course completed" {
			t.Errorf("expected note='Course completed', got %v", updated.Note)
		}
	})

	t.Run("MedicationRequest_ListByPatient", func(t *testing.T) {
		var results []*medication.MedicationRequest
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 medication request")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("MedicationRequest_Search", func(t *testing.T) {
		var results []*medication.MedicationRequest
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
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
		// We may have 0 active ones now since we updated one to completed
		_ = total
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("MedicationRequest_Delete", func(t *testing.T) {
		now := time.Now()
		var reqID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			mr := &medication.MedicationRequest{
				Status:       "draft",
				Intent:       "order",
				MedicationID: med.ID,
				PatientID:    patient.ID,
				RequesterID:  practitioner.ID,
				AuthoredOn:   &now,
			}
			if err := repo.Create(ctx, mr); err != nil {
				return err
			}
			reqID = mr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			return repo.Delete(ctx, reqID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, reqID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted medication request")
		}
	})

	t.Run("Full_Medication_Workflow", func(t *testing.T) {
		// 1. Create medication request (prescription)
		now := time.Now()
		var reqID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationRequestRepoPG(globalDB.Pool)
			mr := &medication.MedicationRequest{
				Status:       "active",
				Intent:       "order",
				MedicationID: med.ID,
				PatientID:    patient.ID,
				RequesterID:  practitioner.ID,
				DosageText:   ptrStr("500mg TID x 10 days"),
				DoseQuantity: ptrFloat(500),
				DoseUnit:     ptrStr("mg"),
				DaysSupply:   ptrInt(10),
				AuthoredOn:   &now,
			}
			if err := repo.Create(ctx, mr); err != nil {
				return err
			}
			reqID = mr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create request: %v", err)
		}

		// 2. Create medication administration (nurse gives it)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationAdministrationRepoPG(globalDB.Pool)
			ma := &medication.MedicationAdministration{
				Status:              "completed",
				MedicationID:        med.ID,
				PatientID:           patient.ID,
				MedicationRequestID: &reqID,
				PerformerID:         &practitioner.ID,
				EffectiveDatetime:   &now,
				DoseQuantity:        ptrFloat(500),
				DoseUnit:            ptrStr("mg"),
				DosageRouteCode:     ptrStr("PO"),
				DosageRouteDisplay:  ptrStr("Oral"),
			}
			return repo.Create(ctx, ma)
		})
		if err != nil {
			t.Fatalf("Create administration: %v", err)
		}

		// 3. Verify administration is linked to request
		var admins []*medication.MedicationAdministration
		var adminTotal int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationAdministrationRepoPG(globalDB.Pool)
			var err error
			admins, adminTotal, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient admins: %v", err)
		}
		if adminTotal == 0 {
			t.Error("expected at least 1 administration")
		}

		foundLinked := false
		for _, a := range admins {
			if a.MedicationRequestID != nil && *a.MedicationRequestID == reqID {
				foundLinked = true
				break
			}
		}
		if !foundLinked {
			t.Error("expected to find administration linked to request")
		}

		// 4. Create dispense record
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationDispenseRepoPG(globalDB.Pool)
			md := &medication.MedicationDispense{
				Status:              "completed",
				MedicationID:        med.ID,
				PatientID:           patient.ID,
				MedicationRequestID: &reqID,
				PerformerID:         &practitioner.ID,
				QuantityValue:       ptrFloat(30),
				QuantityUnit:        ptrStr("capsule"),
				DaysSupply:          ptrInt(10),
				WhenPrepared:        &now,
				WhenHandedOver:      &now,
			}
			return repo.Create(ctx, md)
		})
		if err != nil {
			t.Fatalf("Create dispense: %v", err)
		}

		// 5. Verify dispense by patient
		var dispenses []*medication.MedicationDispense
		var dispTotal int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := medication.NewMedicationDispenseRepoPG(globalDB.Pool)
			var err error
			dispenses, dispTotal, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient dispenses: %v", err)
		}
		if dispTotal == 0 {
			t.Error("expected at least 1 dispense")
		}

		foundDispLinked := false
		for _, d := range dispenses {
			if d.MedicationRequestID != nil && *d.MedicationRequestID == reqID {
				foundDispLinked = true
				break
			}
		}
		if !foundDispLinked {
			t.Error("expected to find dispense linked to request")
		}
	})
}
