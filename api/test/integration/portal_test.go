package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/portal"
	"github.com/google/uuid"
)

func TestPortalAccountCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("pacct")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "PortalPatient", "Test", "MRN-PORTAL-001")

	t.Run("Create", func(t *testing.T) {
		var created *portal.PortalAccount
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			a := &portal.PortalAccount{
				PatientID:         patient.ID,
				Username:          "jdoe_portal",
				Email:             "jdoe@example.com",
				Phone:             ptrStr("555-0100"),
				Status:            "active",
				EmailVerified:     true,
				MFAEnabled:        false,
				PreferredLanguage: ptrStr("en"),
				Note:              ptrStr("Account created during registration"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create portal account: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			a := &portal.PortalAccount{
				PatientID: uuid.New(),
				Username:  "fake_user",
				Email:     "fake@example.com",
				Status:    "active",
			}
			return repo.Create(ctx, a)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var acctID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			a := &portal.PortalAccount{
				PatientID: patient.ID,
				Username:  "getbyid_user",
				Email:     "getbyid@example.com",
				Status:    "active",
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			acctID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *portal.PortalAccount
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, acctID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Username != "getbyid_user" {
			t.Errorf("expected username=getbyid_user, got %s", fetched.Username)
		}
		if fetched.Email != "getbyid@example.com" {
			t.Errorf("expected email=getbyid@example.com, got %s", fetched.Email)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var acct *portal.PortalAccount
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			a := &portal.PortalAccount{
				PatientID: patient.ID,
				Username:  "update_user",
				Email:     "update@example.com",
				Status:    "active",
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			acct = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			acct.Email = "updated@example.com"
			acct.MFAEnabled = true
			acct.Note = ptrStr("MFA enabled by user")
			return repo.Update(ctx, acct)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *portal.PortalAccount
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, acct.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Email != "updated@example.com" {
			t.Errorf("expected email=updated@example.com, got %s", fetched.Email)
		}
		if !fetched.MFAEnabled {
			t.Error("expected mfa_enabled=true")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*portal.PortalAccount
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 portal account")
		}
		_ = results
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*portal.PortalAccount
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 portal account for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var acctID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			a := &portal.PortalAccount{
				PatientID: patient.ID,
				Username:  "delete_user",
				Email:     "delete@example.com",
				Status:    "active",
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			acctID = a.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			return repo.Delete(ctx, acctID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalAccountRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, acctID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted portal account")
		}
	})
}

func TestPortalMessageCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("pmsg")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "MsgPatient", "Test", "MRN-MSG-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "MsgDoc", "Smith")

	t.Run("Create", func(t *testing.T) {
		var created *portal.PortalMessage
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			m := &portal.PortalMessage{
				PatientID:      patient.ID,
				PractitionerID: &practitioner.ID,
				Direction:      "inbound",
				Subject:        ptrStr("Medication question"),
				Body:           "Can I take ibuprofen with my current prescription?",
				Status:         "sent",
				Priority:       ptrStr("routine"),
				Category:       ptrStr("medication"),
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			created = m
			return nil
		})
		if err != nil {
			t.Fatalf("Create portal message: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			m := &portal.PortalMessage{
				PatientID: uuid.New(),
				Direction: "inbound",
				Body:      "test",
				Status:    "sent",
			}
			return repo.Create(ctx, m)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var msgID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			m := &portal.PortalMessage{
				PatientID: patient.ID,
				Direction: "outbound",
				Subject:   ptrStr("Lab results available"),
				Body:      "Your lab results are now available in the portal.",
				Status:    "sent",
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			msgID = m.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *portal.PortalMessage
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, msgID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Direction != "outbound" {
			t.Errorf("expected direction=outbound, got %s", fetched.Direction)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var msg *portal.PortalMessage
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			m := &portal.PortalMessage{
				PatientID: patient.ID,
				Direction: "outbound",
				Body:      "Test message body",
				Status:    "sent",
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			msg = m
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		now := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			msg.Status = "read"
			msg.ReadAt = &now
			return repo.Update(ctx, msg)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *portal.PortalMessage
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, msg.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "read" {
			t.Errorf("expected status=read, got %s", fetched.Status)
		}
		if fetched.ReadAt == nil {
			t.Error("expected non-nil ReadAt")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*portal.PortalMessage
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 message")
		}
		_ = results
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*portal.PortalMessage
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 message for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var msgID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			m := &portal.PortalMessage{
				PatientID: patient.ID,
				Direction: "inbound",
				Body:      "Delete me",
				Status:    "sent",
			}
			if err := repo.Create(ctx, m); err != nil {
				return err
			}
			msgID = m.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			return repo.Delete(ctx, msgID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPortalMessageRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, msgID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted message")
		}
	})
}

func TestQuestionnaireCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("quest")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *portal.Questionnaire
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			q := &portal.Questionnaire{
				Name:        "patient-intake-form",
				Title:       ptrStr("Patient Intake Form"),
				Status:      "active",
				Version:     ptrStr("1.0"),
				Description: ptrStr("Standard patient intake questionnaire"),
				Purpose:     ptrStr("Collect patient information at intake"),
				SubjectType: ptrStr("Patient"),
				Date:        &now,
				Publisher:   ptrStr("Example Health System"),
			}
			if err := repo.Create(ctx, q); err != nil {
				return err
			}
			created = q
			return nil
		})
		if err != nil {
			t.Fatalf("Create questionnaire: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
		if created.FHIRID == "" {
			t.Fatal("expected non-empty FHIR ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var questID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			q := &portal.Questionnaire{
				Name:   "getbyid-quest",
				Status: "draft",
			}
			if err := repo.Create(ctx, q); err != nil {
				return err
			}
			questID = q.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *portal.Questionnaire
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, questID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "getbyid-quest" {
			t.Errorf("expected name=getbyid-quest, got %s", fetched.Name)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			q := &portal.Questionnaire{
				Name:   "fhirid-quest",
				Status: "active",
			}
			if err := repo.Create(ctx, q); err != nil {
				return err
			}
			fhirID = q.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *portal.Questionnaire
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
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
		var quest *portal.Questionnaire
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			q := &portal.Questionnaire{
				Name:   "update-quest",
				Status: "draft",
				Title:  ptrStr("Draft Questionnaire"),
			}
			if err := repo.Create(ctx, q); err != nil {
				return err
			}
			quest = q
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			quest.Status = "active"
			quest.Title = ptrStr("Active Questionnaire")
			quest.Version = ptrStr("2.0")
			return repo.Update(ctx, quest)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *portal.Questionnaire
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, quest.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
		if fetched.Version == nil || *fetched.Version != "2.0" {
			t.Errorf("expected version=2.0, got %v", fetched.Version)
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*portal.Questionnaire
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 questionnaire")
		}
		_ = results
	})

	t.Run("Search", func(t *testing.T) {
		var results []*portal.Questionnaire
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"status": "active",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Status != "active" {
				t.Errorf("expected status=active, got %s", r.Status)
			}
		}
	})

	t.Run("Items", func(t *testing.T) {
		var questID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			q := &portal.Questionnaire{
				Name:   "items-quest",
				Status: "active",
			}
			if err := repo.Create(ctx, q); err != nil {
				return err
			}
			questID = q.ID

			// Add items
			item1 := &portal.QuestionnaireItem{
				QuestionnaireID: questID,
				LinkID:          "q1",
				Text:            "What is your name?",
				Type:            "string",
				Required:        true,
				SortOrder:       1,
			}
			if err := repo.AddItem(ctx, item1); err != nil {
				return err
			}

			opts, _ := json.Marshal([]map[string]string{
				{"value": "male", "display": "Male"},
				{"value": "female", "display": "Female"},
			})
			item2 := &portal.QuestionnaireItem{
				QuestionnaireID: questID,
				LinkID:          "q2",
				Text:            "What is your gender?",
				Type:            "choice",
				Required:        true,
				AnswerOptions:   opts,
				SortOrder:       2,
			}
			return repo.AddItem(ctx, item2)
		})
		if err != nil {
			t.Fatalf("Create questionnaire with items: %v", err)
		}

		var items []*portal.QuestionnaireItem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			var err error
			items, err = repo.GetItems(ctx, questID)
			return err
		})
		if err != nil {
			t.Fatalf("GetItems: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
		if items[0].LinkID != "q1" {
			t.Errorf("expected first item link_id=q1, got %s", items[0].LinkID)
		}
		if items[1].Type != "choice" {
			t.Errorf("expected second item type=choice, got %s", items[1].Type)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var questID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			q := &portal.Questionnaire{
				Name:   "delete-quest",
				Status: "draft",
			}
			if err := repo.Create(ctx, q); err != nil {
				return err
			}
			questID = q.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			return repo.Delete(ctx, questID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, questID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted questionnaire")
		}
	})
}

func TestQuestionnaireResponseCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("qresp")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "QRPatient", "Test", "MRN-QR-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "QRDoc", "Smith")

	// Create a questionnaire to reference
	var questID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := portal.NewQuestionnaireRepoPG(globalDB.Pool)
		q := &portal.Questionnaire{
			Name:   "qr-test-quest",
			Status: "active",
		}
		if err := repo.Create(ctx, q); err != nil {
			return err
		}
		questID = q.ID
		return nil
	})
	if err != nil {
		t.Fatalf("Create prerequisite questionnaire: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *portal.QuestionnaireResponse
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			qr := &portal.QuestionnaireResponse{
				QuestionnaireID: questID,
				PatientID:       patient.ID,
				AuthorID:        &practitioner.ID,
				Status:          "completed",
				Authored:        &now,
			}
			if err := repo.Create(ctx, qr); err != nil {
				return err
			}
			created = qr
			return nil
		})
		if err != nil {
			t.Fatalf("Create questionnaire response: %v", err)
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
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			qr := &portal.QuestionnaireResponse{
				QuestionnaireID: uuid.New(),
				PatientID:       patient.ID,
				Status:          "completed",
			}
			return repo.Create(ctx, qr)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent questionnaire")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var respID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			qr := &portal.QuestionnaireResponse{
				QuestionnaireID: questID,
				PatientID:       patient.ID,
				Status:          "in-progress",
			}
			if err := repo.Create(ctx, qr); err != nil {
				return err
			}
			respID = qr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *portal.QuestionnaireResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, respID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "in-progress" {
			t.Errorf("expected status=in-progress, got %s", fetched.Status)
		}
	})

	t.Run("GetByFHIRID", func(t *testing.T) {
		var fhirID string
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			qr := &portal.QuestionnaireResponse{
				QuestionnaireID: questID,
				PatientID:       patient.ID,
				Status:          "completed",
			}
			if err := repo.Create(ctx, qr); err != nil {
				return err
			}
			fhirID = qr.FHIRID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *portal.QuestionnaireResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
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
		var resp *portal.QuestionnaireResponse
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			qr := &portal.QuestionnaireResponse{
				QuestionnaireID: questID,
				PatientID:       patient.ID,
				Status:          "in-progress",
			}
			if err := repo.Create(ctx, qr); err != nil {
				return err
			}
			resp = qr
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		now := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			resp.Status = "completed"
			resp.Authored = &now
			return repo.Update(ctx, resp)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *portal.QuestionnaireResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, resp.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.Authored == nil {
			t.Error("expected non-nil Authored")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*portal.QuestionnaireResponse
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 response")
		}
		_ = results
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*portal.QuestionnaireResponse
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 response for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*portal.QuestionnaireResponse
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient": patient.ID.String(),
				"status":  "completed",
			}, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		_ = total
		for _, r := range results {
			if r.Status != "completed" {
				t.Errorf("expected status=completed, got %s", r.Status)
			}
		}
	})

	t.Run("ResponseItems", func(t *testing.T) {
		var respID uuid.UUID
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			qr := &portal.QuestionnaireResponse{
				QuestionnaireID: questID,
				PatientID:       patient.ID,
				Status:          "completed",
				Authored:        &now,
			}
			if err := repo.Create(ctx, qr); err != nil {
				return err
			}
			respID = qr.ID

			// Add response items
			item1 := &portal.QuestionnaireResponseItem{
				ResponseID: respID,
				LinkID:     "q1",
				Text:       ptrStr("What is your name?"),
				AnswerStr:  ptrStr("John Doe"),
			}
			if err := repo.AddResponseItem(ctx, item1); err != nil {
				return err
			}

			item2 := &portal.QuestionnaireResponseItem{
				ResponseID: respID,
				LinkID:     "q2",
				Text:       ptrStr("Age"),
				AnswerInt:  ptrInt(45),
			}
			if err := repo.AddResponseItem(ctx, item2); err != nil {
				return err
			}

			item3 := &portal.QuestionnaireResponseItem{
				ResponseID: respID,
				LinkID:     "q3",
				Text:       ptrStr("Active smoker?"),
				AnswerBool: ptrBool(false),
			}
			return repo.AddResponseItem(ctx, item3)
		})
		if err != nil {
			t.Fatalf("Create response with items: %v", err)
		}

		var items []*portal.QuestionnaireResponseItem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			var err error
			items, err = repo.GetResponseItems(ctx, respID)
			return err
		})
		if err != nil {
			t.Fatalf("GetResponseItems: %v", err)
		}
		if len(items) != 3 {
			t.Fatalf("expected 3 response items, got %d", len(items))
		}
		foundStr := false
		foundInt := false
		foundBool := false
		for _, item := range items {
			if item.AnswerStr != nil && *item.AnswerStr == "John Doe" {
				foundStr = true
			}
			if item.AnswerInt != nil && *item.AnswerInt == 45 {
				foundInt = true
			}
			if item.AnswerBool != nil && !*item.AnswerBool {
				foundBool = true
			}
		}
		if !foundStr {
			t.Error("expected to find string answer 'John Doe'")
		}
		if !foundInt {
			t.Error("expected to find integer answer 45")
		}
		if !foundBool {
			t.Error("expected to find boolean answer false")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var respID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			qr := &portal.QuestionnaireResponse{
				QuestionnaireID: questID,
				PatientID:       patient.ID,
				Status:          "completed",
			}
			if err := repo.Create(ctx, qr); err != nil {
				return err
			}
			respID = qr.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			return repo.Delete(ctx, respID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewQuestionnaireResponseRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, respID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted questionnaire response")
		}
	})
}

func TestPatientCheckinCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("checkin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "CheckinPatient", "Test", "MRN-CHECKIN-001")

	t.Run("Create", func(t *testing.T) {
		var created *portal.PatientCheckin
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			c := &portal.PatientCheckin{
				PatientID:         patient.ID,
				Status:            "completed",
				CheckinMethod:     ptrStr("kiosk"),
				CheckinTime:       &now,
				InsuranceVerified: ptrBool(true),
				CoPayCollected:    ptrBool(true),
				CoPayAmount:       ptrFloat(25.00),
				Note:              ptrStr("Checked in via kiosk"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			created = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create patient checkin: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("Create_FK_Violation", func(t *testing.T) {
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			c := &portal.PatientCheckin{
				PatientID: uuid.New(),
				Status:    "pending",
			}
			return repo.Create(ctx, c)
		})
		if err == nil {
			t.Fatal("expected FK violation for non-existent patient")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var checkinID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			c := &portal.PatientCheckin{
				PatientID:     patient.ID,
				Status:        "pending",
				CheckinMethod: ptrStr("online"),
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			checkinID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *portal.PatientCheckin
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, checkinID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Status != "pending" {
			t.Errorf("expected status=pending, got %s", fetched.Status)
		}
		if fetched.CheckinMethod == nil || *fetched.CheckinMethod != "online" {
			t.Errorf("expected checkin_method=online, got %v", fetched.CheckinMethod)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var checkin *portal.PatientCheckin
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			c := &portal.PatientCheckin{
				PatientID: patient.ID,
				Status:    "pending",
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			checkin = c
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		now := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			checkin.Status = "completed"
			checkin.CheckinTime = &now
			checkin.InsuranceVerified = ptrBool(true)
			checkin.CoPayCollected = ptrBool(false)
			checkin.Note = ptrStr("Insurance verified, co-pay waived")
			return repo.Update(ctx, checkin)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *portal.PatientCheckin
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, checkin.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.InsuranceVerified == nil || !*fetched.InsuranceVerified {
			t.Error("expected insurance_verified=true")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*portal.PatientCheckin
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 checkin")
		}
		_ = results
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*portal.PatientCheckin
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 checkin for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var checkinID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			c := &portal.PatientCheckin{
				PatientID: patient.ID,
				Status:    "pending",
			}
			if err := repo.Create(ctx, c); err != nil {
				return err
			}
			checkinID = c.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			return repo.Delete(ctx, checkinID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := portal.NewPatientCheckinRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, checkinID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted checkin")
		}
	})
}
