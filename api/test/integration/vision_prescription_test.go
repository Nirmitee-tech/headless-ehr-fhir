package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/visionprescription"
	"github.com/google/uuid"
)

func TestVisionPrescriptionCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("vp")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "VPPatient", "Test", "MRN-VP-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "VPDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *visionprescription.VisionPrescription
		now := time.Now()
		dateWritten := time.Now().AddDate(0, 0, -1)
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			vp := &visionprescription.VisionPrescription{
				Status:       "active",
				Created:      ptrTime(now),
				PatientID:    patient.ID,
				EncounterID:  nil,
				DateWritten:  ptrTime(dateWritten),
				PrescriberID: ptrUUID(practitioner.ID),
			}
			if err := repo.Create(ctx, vp); err != nil {
				return err
			}
			created = vp
			return nil
		})
		if err != nil {
			t.Fatalf("Create vision prescription: %v", err)
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
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			fakePatient := uuid.New()
			vp := &visionprescription.VisionPrescription{
				Status:    "active",
				PatientID: fakePatient,
			}
			return repo.Create(ctx, vp)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("Create_FK_Violation_Prescriber", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			fakePractitioner := uuid.New()
			vp := &visionprescription.VisionPrescription{
				Status:       "active",
				PatientID:    patient.ID,
				PrescriberID: ptrUUID(fakePractitioner),
			}
			return repo.Create(ctx, vp)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent prescriber")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var vpID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			vp := &visionprescription.VisionPrescription{
				Status:       "active",
				Created:      ptrTime(now),
				PatientID:    patient.ID,
				DateWritten:  ptrTime(now),
				PrescriberID: ptrUUID(practitioner.ID),
			}
			if err := repo.Create(ctx, vp); err != nil {
				return err
			}
			vpID = vp.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *visionprescription.VisionPrescription
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, vpID)
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
		if fetched.PrescriberID == nil || *fetched.PrescriberID != practitioner.ID {
			t.Errorf("expected prescriber_id=%s, got %v", practitioner.ID, fetched.PrescriberID)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			vp := &visionprescription.VisionPrescription{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, vp); err != nil {
				return err
			}
			fhirID = vp.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *visionprescription.VisionPrescription
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
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
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var vp *visionprescription.VisionPrescription
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			v := &visionprescription.VisionPrescription{
				Status:       "active",
				Created:      ptrTime(now),
				PatientID:    patient.ID,
				PrescriberID: ptrUUID(practitioner.ID),
			}
			if err := repo.Create(ctx, v); err != nil {
				return err
			}
			vp = v
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		// Update status and add date_written
		newDateWritten := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			vp.Status = "cancelled"
			vp.DateWritten = ptrTime(newDateWritten)
			return repo.Update(ctx, vp)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *visionprescription.VisionPrescription
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, vp.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "cancelled" {
			t.Errorf("expected status=cancelled, got %s", fetched.Status)
		}
		if fetched.DateWritten == nil {
			t.Error("expected non-nil DateWritten after update")
		}
	})

	t.Run("Search_ByPatient", func(t *testing.T) {
		// Create a prescription to ensure at least one result
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			vp := &visionprescription.VisionPrescription{
				Status:    "active",
				PatientID: patient.ID,
			}
			return repo.Create(ctx, vp)
		})
		if err != nil {
			t.Fatalf("Create for search: %v", err)
		}

		var results []*visionprescription.VisionPrescription
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
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
			t.Error("expected at least 1 vision prescription for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*visionprescription.VisionPrescription
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
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
		if total == 0 {
			t.Error("expected at least 1 active vision prescription")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var vpID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			vp := &visionprescription.VisionPrescription{
				Status:    "active",
				PatientID: patient.ID,
			}
			if err := repo.Create(ctx, vp); err != nil {
				return err
			}
			vpID = vp.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			return repo.Delete(ctx, vpID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, vpID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted vision prescription")
		}
	})
}

func TestVisionPrescriptionLensSpec(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("vp")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "VPLensPatient", "Test", "MRN-VP-002")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "VPLensDoc", "Jones")

	// Create a prescription to attach lens specs to
	var prescriptionID uuid.UUID
	now := time.Now()
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
		vp := &visionprescription.VisionPrescription{
			Status:       "active",
			Created:      ptrTime(now),
			PatientID:    patient.ID,
			DateWritten:  ptrTime(now),
			PrescriberID: ptrUUID(practitioner.ID),
		}
		if err := repo.Create(ctx, vp); err != nil {
			return err
		}
		prescriptionID = vp.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create prescription for lens spec tests: %v", err)
	}

	t.Run("AddLensSpec", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			ls := &visionprescription.VisionPrescriptionLensSpec{
				PrescriptionID: prescriptionID,
				ProductCode:    "lens-single-vision",
				ProductDisplay: ptrStr("Single Vision Lenses"),
				Eye:            "right",
				Sphere:         ptrFloat(-2.00),
				Cylinder:       ptrFloat(-0.50),
				Axis:           ptrInt(180),
				AddPower:       ptrFloat(1.75),
			}
			return repo.AddLensSpec(ctx, ls)
		})
		if err != nil {
			t.Fatalf("AddLensSpec: %v", err)
		}
	})

	t.Run("AddLensSpec_BothEyes", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			ls := &visionprescription.VisionPrescriptionLensSpec{
				PrescriptionID: prescriptionID,
				ProductCode:    "lens-single-vision",
				ProductDisplay: ptrStr("Single Vision Lenses"),
				Eye:            "left",
				Sphere:         ptrFloat(-1.75),
				Cylinder:       ptrFloat(-0.25),
				Axis:           ptrInt(170),
				AddPower:       ptrFloat(1.75),
			}
			return repo.AddLensSpec(ctx, ls)
		})
		if err != nil {
			t.Fatalf("AddLensSpec left eye: %v", err)
		}
	})

	t.Run("GetLensSpecs", func(t *testing.T) {
		var specs []*visionprescription.VisionPrescriptionLensSpec
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			var err error
			specs, err = repo.GetLensSpecs(ctx, prescriptionID)
			return err
		})
		if err != nil {
			t.Fatalf("GetLensSpecs: %v", err)
		}
		if len(specs) < 2 {
			t.Fatalf("expected at least 2 lens specs, got %d", len(specs))
		}
		for _, s := range specs {
			if s.PrescriptionID != prescriptionID {
				t.Errorf("expected prescription_id=%s, got %s", prescriptionID, s.PrescriptionID)
			}
			if s.ProductCode != "lens-single-vision" {
				t.Errorf("expected product_code=lens-single-vision, got %s", s.ProductCode)
			}
			if s.Eye != "right" && s.Eye != "left" {
				t.Errorf("expected eye=right or left, got %s", s.Eye)
			}
		}
	})

	t.Run("AddLensSpec_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := visionprescription.NewVisionPrescriptionRepoPG(globalDB.Pool)
			ls := &visionprescription.VisionPrescriptionLensSpec{
				PrescriptionID: uuid.New(), // non-existent prescription
				ProductCode:    "lens-contact",
				Eye:            "right",
			}
			return repo.AddLensSpec(ctx, ls)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent prescription")
		}
	})
}
