package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/research"
	"github.com/google/uuid"
)

func TestResearchStudyCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("study")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "StudyPI", "Johnson")

	t.Run("Create", func(t *testing.T) {
		var created *research.ResearchStudy
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			s := &research.ResearchStudy{
				Title:                    "Phase III Clinical Trial for Drug XYZ",
				ProtocolNumber:           "PROTO-2024-001",
				Status:                   "active-recruiting",
				Phase:                    ptrStr("phase-3"),
				Category:                 ptrStr("interventional"),
				Focus:                    ptrStr("oncology"),
				Description:              ptrStr("A randomized, double-blind study evaluating Drug XYZ"),
				SponsorName:              ptrStr("PharmaCorp Inc."),
				SponsorContact:           ptrStr("sponsor@pharmacorp.com"),
				PrincipalInvestigatorID:  &practitioner.ID,
				SiteName:                 ptrStr("City General Hospital"),
				SiteContact:              ptrStr("research@citygeneral.org"),
				IRBNumber:                ptrStr("IRB-2024-0456"),
				IRBApprovalDate:          &now,
				StartDate:                &now,
				EnrollmentTarget:         ptrInt(200),
				PrimaryEndpoint:          ptrStr("Overall survival at 12 months"),
				SecondaryEndpoints:       ptrStr("Progression-free survival, QoL scores"),
				InclusionCriteria:        ptrStr("Age 18-75, confirmed diagnosis"),
				ExclusionCriteria:        ptrStr("Prior chemotherapy within 6 months"),
				Note:                     ptrStr("Study initiated per sponsor request"),
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			created = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create research study: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var studyID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			s := &research.ResearchStudy{
				Title:          "GetByID Test Study",
				ProtocolNumber: "PROTO-GET-001",
				Status:         "approved",
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			studyID = s.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *research.ResearchStudy
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, studyID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Title != "GetByID Test Study" {
			t.Errorf("expected title='GetByID Test Study', got %s", fetched.Title)
		}
		if fetched.Status != "approved" {
			t.Errorf("expected status=approved, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			s := &research.ResearchStudy{
				Title:          "FHIRID Test Study",
				ProtocolNumber: "PROTO-FHIR-001",
				Status:         "in-review",
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			fhirID = s.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *research.ResearchStudy
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
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
		var study *research.ResearchStudy
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			s := &research.ResearchStudy{
				Title:          "Update Test Study",
				ProtocolNumber: "PROTO-UPD-001",
				Status:         "approved",
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			study = s
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		now := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			study.Status = "active-recruiting"
			study.Title = "Updated Study Title"
			study.StartDate = &now
			study.EnrollmentTarget = ptrInt(150)
			study.Note = ptrStr("Study activated, recruiting started")
			return repo.Update(ctx, study)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *research.ResearchStudy
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, study.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "active-recruiting" {
			t.Errorf("expected status=active-recruiting, got %s", fetched.Status)
		}
		if fetched.Title != "Updated Study Title" {
			t.Errorf("expected title='Updated Study Title', got %s", fetched.Title)
		}
		if fetched.EnrollmentTarget == nil || *fetched.EnrollmentTarget != 150 {
			t.Errorf("expected enrollment_target=150, got %v", fetched.EnrollmentTarget)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*research.ResearchStudy
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 study")
		}
		_ = results
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*research.ResearchStudy
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active-recruiting",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 active-recruiting study")
		}
		for _, r := range results {
			if r.Status != "active-recruiting" {
				t.Errorf("expected status=active-recruiting, got %s", r.Status)
			}
		}
	})

	t.Run("Search_ByTitle", func(t *testing.T) {
		var results []*research.ResearchStudy
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"title": "Drug XYZ",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by title: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Title == "" {
				t.Error("expected non-empty title")
			}
		}
	})

	t.Run("Arms", func(t *testing.T) {
		var studyID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			s := &research.ResearchStudy{
				Title:          "Arm Test Study",
				ProtocolNumber: "PROTO-ARM-001",
				Status:         "active-recruiting",
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			studyID = s.ID

			// Add arms
			a1 := &research.ResearchArm{
				StudyID:          studyID,
				Name:             "Treatment Arm A",
				ArmType:          ptrStr("experimental"),
				Description:      ptrStr("Drug XYZ 100mg daily"),
				TargetEnrollment: ptrInt(100),
			}
			if err := repo.AddArm(ctx, a1); err != nil {
				return err
			}

			a2 := &research.ResearchArm{
				StudyID:          studyID,
				Name:             "Placebo Arm",
				ArmType:          ptrStr("placebo-comparator"),
				Description:      ptrStr("Matching placebo daily"),
				TargetEnrollment: ptrInt(100),
			}
			return repo.AddArm(ctx, a2)
		})
		if err != nil {
			t.Fatalf("Create study with arms: %v", err)
		}

		var arms []*research.ResearchArm
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			var err error
			arms, err = repo.GetArms(ctx, studyID)
			return err
		})
		if err != nil {
			t.Fatalf("GetArms: %v", err)
		}
		if len(arms) != 2 {
			t.Fatalf("expected 2 arms, got %d", len(arms))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var studyID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			s := &research.ResearchStudy{
				Title:          "Delete Test Study",
				ProtocolNumber: "PROTO-DEL-001",
				Status:         "in-review",
			}
			if err := repo.Create(ctx, s); err != nil {
				return err
			}
			studyID = s.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			return repo.Delete(ctx, studyID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewStudyRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, studyID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted study")
		}
	})
}

func TestEnrollmentCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("enroll")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "EnrollPatient", "Test", "MRN-ENROLL-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "EnrollDoc", "Smith")

	// Create prerequisite study
	var studyID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := research.NewStudyRepoPG(globalDB.Pool)
		s := &research.ResearchStudy{
			Title:          "Enrollment Test Study",
			ProtocolNumber: "PROTO-ENR-001",
			Status:         "active-recruiting",
		}
		if err := repo.Create(ctx, s); err != nil {
			return err
		}
		studyID = s.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create prerequisite study: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *research.ResearchEnrollment
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			e := &research.ResearchEnrollment{
				StudyID:             studyID,
				PatientID:           patient.ID,
				Status:              "enrolled",
				EnrolledDate:        &now,
				ScreeningDate:       &now,
				RandomizationNumber: ptrStr("RND-001"),
				SubjectNumber:       ptrStr("SUBJ-001"),
				EnrolledByID:        &practitioner.ID,
				Note:                ptrStr("Patient enrolled after screening"),
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			created = e
			return nil
		})
		if err != nil {
			t.Fatalf("Create enrollment: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			e := &research.ResearchEnrollment{
				StudyID:   uuid.New(),
				PatientID: patient.ID,
				Status:    "enrolled",
			}
			return repo.Create(ctx, e)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent study")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		patient2 := createTestPatient(t, ctx, globalDB.Pool, tenantID, "EnrGetP", "Test", "MRN-ENR-GET-001")
		var enrollID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			e := &research.ResearchEnrollment{
				StudyID:   studyID,
				PatientID: patient2.ID,
				Status:    "screening",
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			enrollID = e.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *research.ResearchEnrollment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, enrollID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "screening" {
			t.Errorf("expected status=screening, got %s", fetched.Status)
		}
		if fetched.StudyID != studyID {
			t.Errorf("expected study_id=%s, got %s", studyID, fetched.StudyID)
		}
	})

	t.Run("Update", func(t *testing.T) {
		patient3 := createTestPatient(t, ctx, globalDB.Pool, tenantID, "EnrUpdP", "Test", "MRN-ENR-UPD-001")
		var enroll *research.ResearchEnrollment
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			e := &research.ResearchEnrollment{
				StudyID:   studyID,
				PatientID: patient3.ID,
				Status:    "enrolled",
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			enroll = e
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		now := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			enroll.Status = "completed"
			enroll.CompletionDate = &now
			enroll.SubjectNumber = ptrStr("SUBJ-UPD-001")
			enroll.Note = ptrStr("Study completed successfully")
			return repo.Update(ctx, enroll)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *research.ResearchEnrollment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, enroll.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.CompletionDate == nil {
			t.Error("expected non-nil CompletionDate")
		}
	})

	t.Run("ListByStudy", func(t *testing.T) {
		var results []*research.ResearchEnrollment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByStudy(ctx, studyID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByStudy: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 enrollment for study")
		}
		for _, r := range results {
			if r.StudyID != studyID {
				t.Errorf("expected study_id=%s, got %s", studyID, r.StudyID)
			}
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*research.ResearchEnrollment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 enrollment for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		patient4 := createTestPatient(t, ctx, globalDB.Pool, tenantID, "EnrDelP", "Test", "MRN-ENR-DEL-001")
		var enrollID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			e := &research.ResearchEnrollment{
				StudyID:   studyID,
				PatientID: patient4.ID,
				Status:    "screening",
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			enrollID = e.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			return repo.Delete(ctx, enrollID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewEnrollmentRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, enrollID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted enrollment")
		}
	})
}

func TestAdverseEventCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("ae")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "AEPatient", "Test", "MRN-AE-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "AEDoc", "Smith")

	// Create prerequisite study and enrollment
	var enrollmentID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		studyRepo := research.NewStudyRepoPG(globalDB.Pool)
		s := &research.ResearchStudy{
			Title:          "AE Test Study",
			ProtocolNumber: "PROTO-AE-001",
			Status:         "active-recruiting",
		}
		if err := studyRepo.Create(ctx, s); err != nil {
			return err
		}

		enrollRepo := research.NewEnrollmentRepoPG(globalDB.Pool)
		e := &research.ResearchEnrollment{
			StudyID:   s.ID,
			PatientID: patient.ID,
			Status:    "enrolled",
		}
		if err := enrollRepo.Create(ctx, e); err != nil {
			return err
		}
		enrollmentID = e.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create prerequisite study/enrollment: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *research.ResearchAdverseEvent
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			ae := &research.ResearchAdverseEvent{
				EnrollmentID: enrollmentID,
				EventDate:    now,
				ReportedDate: now,
				ReportedByID: &practitioner.ID,
				Description:  "Patient developed mild rash on arms",
				Severity:     ptrStr("mild"),
				Seriousness:  ptrStr("non-serious"),
				Causality:    ptrStr("possible"),
				Expectedness: ptrStr("expected"),
				Outcome:      ptrStr("recovering"),
				ActionTaken:  ptrStr("dose-not-changed"),
				Note:         ptrStr("Monitoring rash progression"),
			}
			if err := repo.Create(ctx, ae); err != nil {
				return err
			}
			created = ae
			return nil
		})
		if err != nil {
			t.Fatalf("Create adverse event: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			ae := &research.ResearchAdverseEvent{
				EnrollmentID: uuid.New(),
				EventDate:    now,
				ReportedDate: now,
				Description:  "test",
			}
			return repo.Create(ctx, ae)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent enrollment")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var aeID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			ae := &research.ResearchAdverseEvent{
				EnrollmentID: enrollmentID,
				EventDate:    now,
				ReportedDate: now,
				Description:  "Headache reported after dose",
				Severity:     ptrStr("moderate"),
			}
			if err := repo.Create(ctx, ae); err != nil {
				return err
			}
			aeID = ae.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *research.ResearchAdverseEvent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, aeID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Description != "Headache reported after dose" {
			t.Errorf("expected description='Headache reported after dose', got %s", fetched.Description)
		}
		if fetched.Severity == nil || *fetched.Severity != "moderate" {
			t.Errorf("expected severity=moderate, got %v", fetched.Severity)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var ae *research.ResearchAdverseEvent
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			a := &research.ResearchAdverseEvent{
				EnrollmentID: enrollmentID,
				EventDate:    now,
				ReportedDate: now,
				Description:  "Nausea after treatment",
				Severity:     ptrStr("mild"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			ae = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		resolution := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			ae.Outcome = ptrStr("resolved")
			ae.ResolutionDate = &resolution
			ae.ReportedToIRB = ptrBool(true)
			ae.IRBReportDate = &resolution
			ae.Note = ptrStr("Resolved without intervention")
			return repo.Update(ctx, ae)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *research.ResearchAdverseEvent
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ae.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Outcome == nil || *fetched.Outcome != "resolved" {
			t.Errorf("expected outcome=resolved, got %v", fetched.Outcome)
		}
		if fetched.ReportedToIRB == nil || !*fetched.ReportedToIRB {
			t.Error("expected reported_to_irb=true")
		}
		if fetched.ResolutionDate == nil {
			t.Error("expected non-nil ResolutionDate")
		}
	})

	t.Run("ListByEnrollment", func(t *testing.T) {
		var results []*research.ResearchAdverseEvent
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByEnrollment(ctx, enrollmentID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByEnrollment: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 adverse event for enrollment")
		}
		for _, r := range results {
			if r.EnrollmentID != enrollmentID {
				t.Errorf("expected enrollment_id=%s, got %s", enrollmentID, r.EnrollmentID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var aeID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			ae := &research.ResearchAdverseEvent{
				EnrollmentID: enrollmentID,
				EventDate:    now,
				ReportedDate: now,
				Description:  "Delete test AE",
			}
			if err := repo.Create(ctx, ae); err != nil {
				return err
			}
			aeID = ae.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			return repo.Delete(ctx, aeID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewAdverseEventRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, aeID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted adverse event")
		}
	})
}

func TestDeviationCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("dev")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "DevPatient", "Test", "MRN-DEV-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "DevDoc", "Smith")

	// Create prerequisite study and enrollment
	var enrollmentID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		studyRepo := research.NewStudyRepoPG(globalDB.Pool)
		s := &research.ResearchStudy{
			Title:          "Deviation Test Study",
			ProtocolNumber: "PROTO-DEV-001",
			Status:         "active-recruiting",
		}
		if err := studyRepo.Create(ctx, s); err != nil {
			return err
		}

		enrollRepo := research.NewEnrollmentRepoPG(globalDB.Pool)
		e := &research.ResearchEnrollment{
			StudyID:   s.ID,
			PatientID: patient.ID,
			Status:    "enrolled",
		}
		if err := enrollRepo.Create(ctx, e); err != nil {
			return err
		}
		enrollmentID = e.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create prerequisite study/enrollment: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *research.ResearchProtocolDeviation
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			d := &research.ResearchProtocolDeviation{
				EnrollmentID:     enrollmentID,
				DeviationDate:    now,
				ReportedDate:     now,
				ReportedByID:     &practitioner.ID,
				Category:         ptrStr("procedural"),
				Description:      "Blood draw performed outside protocol window",
				Severity:         ptrStr("minor"),
				CorrectiveAction: ptrStr("Sample collected within 24 hours"),
				PreventiveAction: ptrStr("Calendar reminders set for future visits"),
				ImpactOnSubject:  ptrStr("none"),
				ImpactOnStudy:    ptrStr("minimal"),
				Note:             ptrStr("Staff notified of correct schedule"),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			created = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create deviation: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			d := &research.ResearchProtocolDeviation{
				EnrollmentID:  uuid.New(),
				DeviationDate: now,
				ReportedDate:  now,
				Description:   "test",
			}
			return repo.Create(ctx, d)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent enrollment")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var devID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			d := &research.ResearchProtocolDeviation{
				EnrollmentID:  enrollmentID,
				DeviationDate: now,
				ReportedDate:  now,
				Category:      ptrStr("documentation"),
				Description:   "Informed consent not dated correctly",
				Severity:      ptrStr("minor"),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			devID = d.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *research.ResearchProtocolDeviation
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, devID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Description != "Informed consent not dated correctly" {
			t.Errorf("unexpected description: %s", fetched.Description)
		}
		if fetched.Category == nil || *fetched.Category != "documentation" {
			t.Errorf("expected category=documentation, got %v", fetched.Category)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var dev *research.ResearchProtocolDeviation
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			d := &research.ResearchProtocolDeviation{
				EnrollmentID:  enrollmentID,
				DeviationDate: now,
				ReportedDate:  now,
				Description:   "Missed study visit",
				Severity:      ptrStr("minor"),
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			dev = d
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		irbDate := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			dev.Severity = ptrStr("major")
			dev.CorrectiveAction = ptrStr("Visit rescheduled within protocol window")
			dev.PreventiveAction = ptrStr("Enhanced reminder system")
			dev.ReportedToIRB = ptrBool(true)
			dev.IRBReportDate = &irbDate
			dev.Note = ptrStr("Escalated to major due to missed visit impact")
			return repo.Update(ctx, dev)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *research.ResearchProtocolDeviation
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, dev.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Severity == nil || *fetched.Severity != "major" {
			t.Errorf("expected severity=major, got %v", fetched.Severity)
		}
		if fetched.ReportedToIRB == nil || !*fetched.ReportedToIRB {
			t.Error("expected reported_to_irb=true")
		}
		if fetched.CorrectiveAction == nil || *fetched.CorrectiveAction != "Visit rescheduled within protocol window" {
			t.Errorf("expected corrective action update, got %v", fetched.CorrectiveAction)
		}
	})

	t.Run("ListByEnrollment", func(t *testing.T) {
		var results []*research.ResearchProtocolDeviation
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByEnrollment(ctx, enrollmentID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByEnrollment: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 deviation for enrollment")
		}
		for _, r := range results {
			if r.EnrollmentID != enrollmentID {
				t.Errorf("expected enrollment_id=%s, got %s", enrollmentID, r.EnrollmentID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var devID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			d := &research.ResearchProtocolDeviation{
				EnrollmentID:  enrollmentID,
				DeviationDate: now,
				ReportedDate:  now,
				Description:   "Delete test deviation",
			}
			if err := repo.Create(ctx, d); err != nil {
				return err
			}
			devID = d.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			return repo.Delete(ctx, devID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := research.NewDeviationRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, devID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted deviation")
		}
	})
}
