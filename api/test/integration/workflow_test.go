package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/workflow"
	"github.com/google/uuid"
)

// =========== ActivityDefinition ===========

func TestActivityDefinitionCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("wf")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *workflow.ActivityDefinition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			ad := &workflow.ActivityDefinition{
				Status:      "active",
				Name:        ptrStr("OrderLabPanel"),
				Title:       ptrStr("Order Lab Panel"),
				Description: ptrStr("Order a standard lab panel for new patients"),
				Kind:        ptrStr("ServiceRequest"),
				CodeCode:    ptrStr("26604007"),
				CodeDisplay: ptrStr("Complete blood count"),
				Intent:      ptrStr("proposal"),
				Priority:    ptrStr("routine"),
				Publisher:   ptrStr("Acme Health System"),
				URL:         ptrStr("http://example.org/fhir/ActivityDefinition/order-lab-panel"),
			}
			if err := repo.Create(ctx, ad); err != nil {
				return err
			}
			created = ad
			return nil
		})
		if err != nil {
			t.Fatalf("Create activity definition: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var adID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			ad := &workflow.ActivityDefinition{
				Status:      "active",
				Name:        ptrStr("CheckBP"),
				Title:       ptrStr("Check Blood Pressure"),
				Description: ptrStr("Routine blood pressure measurement"),
				Kind:        ptrStr("ServiceRequest"),
				CodeCode:    ptrStr("75367002"),
				CodeDisplay: ptrStr("Blood pressure"),
				Intent:      ptrStr("order"),
				Priority:    ptrStr("routine"),
				Publisher:   ptrStr("Test Publisher"),
				URL:         ptrStr("http://example.org/fhir/ActivityDefinition/check-bp"),
			}
			if err := repo.Create(ctx, ad); err != nil {
				return err
			}
			adID = ad.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *workflow.ActivityDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, adID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.Name == nil || *fetched.Name != "CheckBP" {
			t.Errorf("expected name=CheckBP, got %v", fetched.Name)
		}
		if fetched.CodeCode == nil || *fetched.CodeCode != "75367002" {
			t.Errorf("expected code_code=75367002, got %v", fetched.CodeCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			ad := &workflow.ActivityDefinition{
				Status:   "draft",
				Name:     ptrStr("FHIRLookup"),
				Kind:     ptrStr("ServiceRequest"),
				CodeCode: ptrStr("12345"),
			}
			if err := repo.Create(ctx, ad); err != nil {
				return err
			}
			fhirID = ad.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *workflow.ActivityDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
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
		if fetched.Name == nil || *fetched.Name != "FHIRLookup" {
			t.Errorf("expected name=FHIRLookup, got %v", fetched.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var ad *workflow.ActivityDefinition
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			a := &workflow.ActivityDefinition{
				Status:      "draft",
				Name:        ptrStr("UpdateTest"),
				Title:       ptrStr("Update Test Definition"),
				Description: ptrStr("Will be updated"),
				Kind:        ptrStr("ServiceRequest"),
				Publisher:   ptrStr("Draft Publisher"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			ad = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			ad.Status = "active"
			ad.Name = ptrStr("UpdatedName")
			ad.Publisher = ptrStr("Active Publisher")
			return repo.Update(ctx, ad)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *workflow.ActivityDefinition
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, ad.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.Name == nil || *fetched.Name != "UpdatedName" {
			t.Errorf("expected name=UpdatedName, got %v", fetched.Name)
		}
		if fetched.Publisher == nil || *fetched.Publisher != "Active Publisher" {
			t.Errorf("expected publisher=Active Publisher, got %v", fetched.Publisher)
		}
	})

	t.Run("Search_ByStatus", func(t *testing.T) {
		var results []*workflow.ActivityDefinition
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 active activity definition")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Search_ByName", func(t *testing.T) {
		var results []*workflow.ActivityDefinition
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"name": "CheckBP",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by name: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 result for name=CheckBP")
		}
		for _, r := range results {
			if r.Name == nil || *r.Name != "CheckBP" {
				t.Errorf("expected name containing CheckBP, got %v", r.Name)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var adID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			ad := &workflow.ActivityDefinition{
				Status:   "active",
				Name:     ptrStr("ToDelete"),
				Kind:     ptrStr("ServiceRequest"),
				CodeCode: ptrStr("99999"),
			}
			if err := repo.Create(ctx, ad); err != nil {
				return err
			}
			adID = ad.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			return repo.Delete(ctx, adID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewActivityDefinitionRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, adID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted activity definition")
		}
	})
}

// =========== RequestGroup ===========

func TestRequestGroupCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("wf")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "WFPatient", "Test", "MRN-WF-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "WFDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *workflow.RequestGroup
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			rg := &workflow.RequestGroup{
				Status:           "active",
				Intent:           "proposal",
				Priority:         ptrStr("routine"),
				SubjectPatientID: &patient.ID,
				AuthoredOn:       &now,
				AuthorID:         &practitioner.ID,
				ReasonCode:       ptrStr("chronic-disease-management"),
			}
			if err := repo.Create(ctx, rg); err != nil {
				return err
			}
			created = rg
			return nil
		})
		if err != nil {
			t.Fatalf("Create request group: %v", err)
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
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			rg := &workflow.RequestGroup{
				Status:           "active",
				Intent:           "proposal",
				SubjectPatientID: &fakePatient,
			}
			return repo.Create(ctx, rg)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var rgID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			rg := &workflow.RequestGroup{
				Status:           "active",
				Intent:           "order",
				Priority:         ptrStr("urgent"),
				SubjectPatientID: &patient.ID,
				AuthoredOn:       &now,
				AuthorID:         &practitioner.ID,
				ReasonCode:       ptrStr("diabetes-management"),
			}
			if err := repo.Create(ctx, rg); err != nil {
				return err
			}
			rgID = rg.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *workflow.RequestGroup
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, rgID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.Intent != "order" {
			t.Errorf("expected intent=order, got %s", fetched.Intent)
		}
		if fetched.Priority == nil || *fetched.Priority != "urgent" {
			t.Errorf("expected priority=urgent, got %v", fetched.Priority)
		}
		if fetched.SubjectPatientID == nil || *fetched.SubjectPatientID != patient.ID {
			t.Errorf("expected subject_patient_id=%s, got %v", patient.ID, fetched.SubjectPatientID)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			rg := &workflow.RequestGroup{
				Status:           "active",
				Intent:           "proposal",
				SubjectPatientID: &patient.ID,
			}
			if err := repo.Create(ctx, rg); err != nil {
				return err
			}
			fhirID = rg.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *workflow.RequestGroup
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
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
		var rg *workflow.RequestGroup
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			r := &workflow.RequestGroup{
				Status:           "draft",
				Intent:           "proposal",
				Priority:         ptrStr("routine"),
				SubjectPatientID: &patient.ID,
				AuthoredOn:       &now,
				AuthorID:         &practitioner.ID,
				ReasonCode:       ptrStr("initial-assessment"),
			}
			if err := repo.Create(ctx, r); err != nil {
				return err
			}
			rg = r
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			rg.Status = "active"
			rg.Priority = ptrStr("urgent")
			rg.ReasonCode = ptrStr("follow-up-assessment")
			rg.Note = ptrStr("Escalated to urgent priority")
			return repo.Update(ctx, rg)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *workflow.RequestGroup
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, rg.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.Priority == nil || *fetched.Priority != "urgent" {
			t.Errorf("expected priority=urgent, got %v", fetched.Priority)
		}
		if fetched.ReasonCode == nil || *fetched.ReasonCode != "follow-up-assessment" {
			t.Errorf("expected reason_code=follow-up-assessment, got %v", fetched.ReasonCode)
		}
		if fetched.Note == nil || *fetched.Note != "Escalated to urgent priority" {
			t.Errorf("expected note set, got %v", fetched.Note)
		}
	})

	t.Run("Search_ByPatientAndStatus", func(t *testing.T) {
		var results []*workflow.RequestGroup
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by patient and status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 active request group for patient")
		}
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
			if r.SubjectPatientID == nil || *r.SubjectPatientID != patient.ID {
				t.Errorf("expected subject_patient_id=%s, got %v", patient.ID, r.SubjectPatientID)
			}
		}
	})

	t.Run("Actions", func(t *testing.T) {
		var rgID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			rg := &workflow.RequestGroup{
				Status:           "active",
				Intent:           "proposal",
				SubjectPatientID: &patient.ID,
			}
			if err := repo.Create(ctx, rg); err != nil {
				return err
			}
			rgID = rg.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create request group for actions: %v", err)
		}

		// Add action
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			action := &workflow.RequestGroupAction{
				RequestGroupID:    rgID,
				Prefix:            ptrStr("1"),
				Title:             ptrStr("Order CBC"),
				Description:       ptrStr("Order a complete blood count"),
				Priority:          ptrStr("routine"),
				ResourceReference: ptrStr("ActivityDefinition/order-cbc"),
				SelectionBehavior: ptrStr("any"),
			}
			return repo.AddAction(ctx, action)
		})
		if err != nil {
			t.Fatalf("AddAction: %v", err)
		}

		// Add second action
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			action := &workflow.RequestGroupAction{
				RequestGroupID:    rgID,
				Prefix:            ptrStr("2"),
				Title:             ptrStr("Order BMP"),
				Description:       ptrStr("Order a basic metabolic panel"),
				Priority:          ptrStr("routine"),
				ResourceReference: ptrStr("ActivityDefinition/order-bmp"),
			}
			return repo.AddAction(ctx, action)
		})
		if err != nil {
			t.Fatalf("AddAction (second): %v", err)
		}

		// Get actions
		var actions []*workflow.RequestGroupAction
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			var err error
			actions, err = repo.GetActions(ctx, rgID)
			return err
		})
		if err != nil {
			t.Fatalf("GetActions: %v", err)
		}
		if len(actions) != 2 {
			t.Fatalf("expected 2 actions, got %d", len(actions))
		}
		if actions[0].Title == nil || *actions[0].Title != "Order CBC" {
			t.Errorf("expected first action title=Order CBC, got %v", actions[0].Title)
		}
		if actions[1].Title == nil || *actions[1].Title != "Order BMP" {
			t.Errorf("expected second action title=Order BMP, got %v", actions[1].Title)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var rgID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			rg := &workflow.RequestGroup{
				Status:           "active",
				Intent:           "proposal",
				SubjectPatientID: &patient.ID,
			}
			if err := repo.Create(ctx, rg); err != nil {
				return err
			}
			rgID = rg.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			return repo.Delete(ctx, rgID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewRequestGroupRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, rgID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted request group")
		}
	})
}

// =========== GuidanceResponse ===========

func TestGuidanceResponseCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("wf")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "GRPatient", "Test", "MRN-GR-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "GRDoc", "Jones")

	t.Run("Create", func(t *testing.T) {
		var created *workflow.GuidanceResponse
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			gr := &workflow.GuidanceResponse{
				Status:           "success",
				ModuleURI:        "http://example.org/fhir/Library/diabetes-screening",
				SubjectPatientID: &patient.ID,
				OccurrenceDate:   &now,
				PerformerID:      &practitioner.ID,
				ReasonCode:       ptrStr("diabetes-risk-assessment"),
			}
			if err := repo.Create(ctx, gr); err != nil {
				return err
			}
			created = gr
			return nil
		})
		if err != nil {
			t.Fatalf("Create guidance response: %v", err)
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
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			gr := &workflow.GuidanceResponse{
				Status:           "success",
				ModuleURI:        "http://example.org/fhir/Library/test",
				SubjectPatientID: &fakePatient,
			}
			return repo.Create(ctx, gr)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var grID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			gr := &workflow.GuidanceResponse{
				Status:           "success",
				ModuleURI:        "http://example.org/fhir/Library/cardiac-risk",
				SubjectPatientID: &patient.ID,
				OccurrenceDate:   &now,
				PerformerID:      &practitioner.ID,
				ReasonCode:       ptrStr("cardiac-risk-score"),
				ReasonDisplay:    ptrStr("Cardiac Risk Score Calculation"),
			}
			if err := repo.Create(ctx, gr); err != nil {
				return err
			}
			grID = gr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *workflow.GuidanceResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, grID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "success" {
			t.Errorf("expected status=success, got %s", fetched.Status)
		}
		if fetched.ModuleURI != "http://example.org/fhir/Library/cardiac-risk" {
			t.Errorf("expected module_uri=http://example.org/fhir/Library/cardiac-risk, got %s", fetched.ModuleURI)
		}
		if fetched.SubjectPatientID == nil || *fetched.SubjectPatientID != patient.ID {
			t.Errorf("expected subject_patient_id=%s, got %v", patient.ID, fetched.SubjectPatientID)
		}
		if fetched.ReasonCode == nil || *fetched.ReasonCode != "cardiac-risk-score" {
			t.Errorf("expected reason_code=cardiac-risk-score, got %v", fetched.ReasonCode)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			gr := &workflow.GuidanceResponse{
				Status:           "success",
				ModuleURI:        "http://example.org/fhir/Library/fhir-lookup-test",
				SubjectPatientID: &patient.ID,
			}
			if err := repo.Create(ctx, gr); err != nil {
				return err
			}
			fhirID = gr.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *workflow.GuidanceResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
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
		if fetched.ModuleURI != "http://example.org/fhir/Library/fhir-lookup-test" {
			t.Errorf("expected module_uri for fhir lookup test, got %s", fetched.ModuleURI)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var gr *workflow.GuidanceResponse
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			g := &workflow.GuidanceResponse{
				Status:           "in-progress",
				ModuleURI:        "http://example.org/fhir/Library/update-test",
				SubjectPatientID: &patient.ID,
				OccurrenceDate:   &now,
				PerformerID:      &practitioner.ID,
				ReasonCode:       ptrStr("initial-reason"),
			}
			if err := repo.Create(ctx, g); err != nil {
				return err
			}
			gr = g
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			gr.Status = "success"
			gr.ReasonCode = ptrStr("completed-reason")
			gr.ReasonDisplay = ptrStr("Assessment completed")
			gr.Note = ptrStr("Guidance evaluation completed successfully")
			gr.ResultReference = ptrStr("CarePlan/cp-123")
			return repo.Update(ctx, gr)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *workflow.GuidanceResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, gr.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "success" {
			t.Errorf("expected status=success, got %s", fetched.Status)
		}
		if fetched.ReasonCode == nil || *fetched.ReasonCode != "completed-reason" {
			t.Errorf("expected reason_code=completed-reason, got %v", fetched.ReasonCode)
		}
		if fetched.Note == nil || *fetched.Note != "Guidance evaluation completed successfully" {
			t.Errorf("expected note set, got %v", fetched.Note)
		}
		if fetched.ResultReference == nil || *fetched.ResultReference != "CarePlan/cp-123" {
			t.Errorf("expected result_reference=CarePlan/cp-123, got %v", fetched.ResultReference)
		}
	})

	t.Run("Search_ByPatientAndStatus", func(t *testing.T) {
		var results []*workflow.GuidanceResponse
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "success",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search by patient and status: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 successful guidance response for patient")
		}
		for _, r := range results {
			if r.Status != "success" {
				t.Errorf("expected status=success, got %s", r.Status)
			}
			if r.SubjectPatientID == nil || *r.SubjectPatientID != patient.ID {
				t.Errorf("expected subject_patient_id=%s, got %v", patient.ID, r.SubjectPatientID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var grID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			gr := &workflow.GuidanceResponse{
				Status:           "success",
				ModuleURI:        "http://example.org/fhir/Library/delete-test",
				SubjectPatientID: &patient.ID,
			}
			if err := repo.Create(ctx, gr); err != nil {
				return err
			}
			grID = gr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			return repo.Delete(ctx, grID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := workflow.NewGuidanceResponseRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, grID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted guidance response")
		}
	})
}
