package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/cds"
	"github.com/google/uuid"
)

func TestCDSRuleCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("cdsrule")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *cds.CDSRule
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSRuleRepoPG(globalDB.Pool)
			rule := &cds.CDSRule{
				RuleName:       "High Potassium Alert",
				RuleType:       "lab-alert",
				Description:    ptrStr("Alert when potassium > 5.5 mEq/L"),
				Severity:       ptrStr("high"),
				Category:       ptrStr("laboratory"),
				TriggerEvent:   ptrStr("lab-result"),
				ConditionExpr:  ptrStr("potassium > 5.5"),
				ActionType:     ptrStr("alert"),
				ActionDetail:   ptrStr("Notify ordering provider immediately"),
				EvidenceSource: ptrStr("Clinical guidelines"),
				Active:         true,
				Version:        ptrStr("1.0"),
			}
			if err := repo.Create(ctx, rule); err != nil {
				return err
			}
			created = rule
			return nil
		})
		if err != nil {
			t.Fatalf("Create CDS rule: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		rule := createTestCDSRule(t, ctx, tenantID)

		var fetched *cds.CDSRule
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSRuleRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, rule.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.RuleName != "Drug Allergy Check" {
			t.Errorf("expected rule_name=Drug Allergy Check, got %s", fetched.RuleName)
		}
		if !fetched.Active {
			t.Error("expected rule to be active")
		}
	})

	t.Run("Update", func(t *testing.T) {
		rule := createTestCDSRule(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSRuleRepoPG(globalDB.Pool)
			rule.Description = ptrStr("Updated description")
			rule.Severity = ptrStr("critical")
			rule.Active = false
			rule.Version = ptrStr("2.0")
			return repo.Update(ctx, rule)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *cds.CDSRule
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSRuleRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, rule.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Severity == nil || *fetched.Severity != "critical" {
			t.Errorf("expected severity=critical, got %v", fetched.Severity)
		}
		if fetched.Active {
			t.Error("expected rule to be inactive")
		}
		if fetched.Version == nil || *fetched.Version != "2.0" {
			t.Errorf("expected version=2.0, got %v", fetched.Version)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		rule := createTestCDSRule(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSRuleRepoPG(globalDB.Pool)
			return repo.Delete(ctx, rule.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSRuleRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, rule.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted CDS rule")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*cds.CDSRule
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSRuleRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 CDS rule")
		}
		_ = results
	})
}

func TestCDSAlertCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("cdsalert")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "AlertPatient", "Test", "MRN-ALERT-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "AlertDoc", "Smith")
	rule := createTestCDSRule(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *cds.CDSAlert
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
			a := &cds.CDSAlert{
				RuleID:          rule.ID,
				PatientID:       patient.ID,
				PractitionerID:  &practitioner.ID,
				Status:          "active",
				Severity:        ptrStr("high"),
				Summary:         "Drug allergy detected: Penicillin",
				Detail:          ptrStr("Patient has documented allergy to Penicillin"),
				SuggestedAction: ptrStr("Consider alternative antibiotic"),
				Source:          ptrStr("CDS Engine v2"),
				FiredAt:         now,
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create CDS alert: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		alert := createTestCDSAlert(t, ctx, tenantID, rule.ID, patient.ID)

		var fetched *cds.CDSAlert
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, alert.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.RuleID != rule.ID {
			t.Errorf("expected rule_id=%s, got %s", rule.ID, fetched.RuleID)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		alert := createTestCDSAlert(t, ctx, tenantID, rule.ID, patient.ID)

		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
			alert.Status = "resolved"
			alert.ResolvedAt = &now
			return repo.Update(ctx, alert)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *cds.CDSAlert
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, alert.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "resolved" {
			t.Errorf("expected status=resolved, got %s", fetched.Status)
		}
		if fetched.ResolvedAt == nil {
			t.Error("expected non-nil ResolvedAt")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		alert := createTestCDSAlert(t, ctx, tenantID, rule.ID, patient.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
			return repo.Delete(ctx, alert.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, alert.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted CDS alert")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*cds.CDSAlert
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 CDS alert")
		}
		_ = results
	})

	t.Run("ListByPatient", func(t *testing.T) {
		createTestCDSAlert(t, ctx, tenantID, rule.ID, patient.ID)

		var results []*cds.CDSAlert
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 alert for patient")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("AddResponse_and_GetResponses", func(t *testing.T) {
		alert := createTestCDSAlert(t, ctx, tenantID, rule.ID, patient.ID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
			resp := &cds.CDSAlertResponse{
				AlertID:        alert.ID,
				PractitionerID: practitioner.ID,
				Action:         "acknowledged",
				Reason:         ptrStr("Aware of allergy, using alternative"),
				Comment:        ptrStr("Switched to azithromycin"),
			}
			return repo.AddResponse(ctx, resp)
		})
		if err != nil {
			t.Fatalf("AddResponse: %v", err)
		}

		var responses []*cds.CDSAlertResponse
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
			var err error
			responses, err = repo.GetResponses(ctx, alert.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetResponses: %v", err)
		}
		if len(responses) != 1 {
			t.Fatalf("expected 1 response, got %d", len(responses))
		}
		if responses[0].Action != "acknowledged" {
			t.Errorf("expected action=acknowledged, got %s", responses[0].Action)
		}
		if responses[0].PractitionerID != practitioner.ID {
			t.Errorf("expected practitioner_id=%s, got %s", practitioner.ID, responses[0].PractitionerID)
		}
	})
}

func TestDrugInteractionCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("drugint")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *cds.DrugInteraction
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewDrugInteractionRepoPG(globalDB.Pool)
			di := &cds.DrugInteraction{
				MedicationAName: "Warfarin",
				MedicationBName: "Aspirin",
				Severity:        "major",
				Description:     ptrStr("Increased risk of bleeding"),
				ClinicalEffect:  ptrStr("Additive anticoagulant effect"),
				Management:      ptrStr("Monitor INR closely"),
				EvidenceLevel:   ptrStr("high"),
				Source:          ptrStr("DrugBank"),
				Active:          true,
			}
			if err := repo.Create(ctx, di); err != nil {
				return err
			}
			created = di
			return nil
		})
		if err != nil {
			t.Fatalf("Create drug interaction: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		di := createTestDrugInteraction(t, ctx, tenantID)

		var fetched *cds.DrugInteraction
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewDrugInteractionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, di.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Severity != "moderate" {
			t.Errorf("expected severity=moderate, got %s", fetched.Severity)
		}
	})

	t.Run("Update", func(t *testing.T) {
		di := createTestDrugInteraction(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewDrugInteractionRepoPG(globalDB.Pool)
			di.Severity = "major"
			di.Management = ptrStr("Avoid combination")
			di.Active = false
			return repo.Update(ctx, di)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *cds.DrugInteraction
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewDrugInteractionRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, di.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Severity != "major" {
			t.Errorf("expected severity=major, got %s", fetched.Severity)
		}
		if fetched.Active {
			t.Error("expected interaction to be inactive")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		di := createTestDrugInteraction(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewDrugInteractionRepoPG(globalDB.Pool)
			return repo.Delete(ctx, di.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewDrugInteractionRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, di.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted drug interaction")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*cds.DrugInteraction
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewDrugInteractionRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 drug interaction")
		}
		_ = results
	})
}

func TestOrderSetCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("orderset")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *cds.OrderSet
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			os := &cds.OrderSet{
				Name:        "Sepsis Bundle",
				Description: ptrStr("Standard sepsis order set per CMS SEP-1"),
				Category:    ptrStr("emergency"),
				Status:      "active",
				Version:     ptrStr("1.0"),
				Active:      true,
			}
			if err := repo.Create(ctx, os); err != nil {
				return err
			}
			created = os
			return nil
		})
		if err != nil {
			t.Fatalf("Create order set: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		os := createTestOrderSet(t, ctx, tenantID)

		var fetched *cds.OrderSet
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, os.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "Admission Order Set" {
			t.Errorf("expected name=Admission Order Set, got %s", fetched.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {
		os := createTestOrderSet(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			os.Name = "Updated Admission Order Set"
			os.Version = ptrStr("2.0")
			os.Active = false
			return repo.Update(ctx, os)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *cds.OrderSet
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, os.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Name != "Updated Admission Order Set" {
			t.Errorf("expected updated name, got %s", fetched.Name)
		}
		if fetched.Active {
			t.Error("expected order set to be inactive")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		os := createTestOrderSet(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			return repo.Delete(ctx, os.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, os.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted order set")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*cds.OrderSet
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 order set")
		}
		_ = results
	})

	t.Run("AddSection_GetSections_AddItem_GetItems", func(t *testing.T) {
		os := createTestOrderSet(t, ctx, tenantID)

		var sectionID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			section := &cds.OrderSetSection{
				OrderSetID:  os.ID,
				Name:        "Medications",
				Description: ptrStr("Standard admission medications"),
				SortOrder:   1,
			}
			if err := repo.AddSection(ctx, section); err != nil {
				return err
			}
			sectionID = section.ID
			return nil
		})
		if err != nil {
			t.Fatalf("AddSection: %v", err)
		}

		var sections []*cds.OrderSetSection
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			var err error
			sections, err = repo.GetSections(ctx, os.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetSections: %v", err)
		}
		if len(sections) != 1 {
			t.Fatalf("expected 1 section, got %d", len(sections))
		}
		if sections[0].Name != "Medications" {
			t.Errorf("expected section name=Medications, got %s", sections[0].Name)
		}

		// Add item to section
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			item := &cds.OrderSetItem{
				SectionID:        sectionID,
				ItemType:         "medication",
				ItemName:         "Normal Saline 1000mL",
				ItemCode:         ptrStr("313002"),
				DefaultDose:      ptrStr("1000 mL"),
				DefaultFrequency: ptrStr("continuous"),
				Instructions:     ptrStr("Infuse at 125 mL/hr"),
				IsRequired:       true,
				SortOrder:        1,
			}
			return repo.AddItem(ctx, item)
		})
		if err != nil {
			t.Fatalf("AddItem: %v", err)
		}

		var items []*cds.OrderSetItem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewOrderSetRepoPG(globalDB.Pool)
			var err error
			items, err = repo.GetItems(ctx, sectionID)
			return err
		})
		if err != nil {
			t.Fatalf("GetItems: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if items[0].ItemName != "Normal Saline 1000mL" {
			t.Errorf("expected item=Normal Saline 1000mL, got %s", items[0].ItemName)
		}
		if !items[0].IsRequired {
			t.Error("expected item to be required")
		}
	})
}

func TestClinicalPathwayCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("pathway")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *cds.ClinicalPathway
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
			p := &cds.ClinicalPathway{
				Name:             "CHF Management Pathway",
				Description:      ptrStr("Clinical pathway for congestive heart failure"),
				Condition:        ptrStr("heart-failure"),
				Category:         ptrStr("cardiology"),
				Version:          ptrStr("1.0"),
				Active:           true,
				ExpectedDuration: ptrStr("14 days"),
			}
			if err := repo.Create(ctx, p); err != nil {
				return err
			}
			created = p
			return nil
		})
		if err != nil {
			t.Fatalf("Create clinical pathway: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		pw := createTestClinicalPathway(t, ctx, tenantID)

		var fetched *cds.ClinicalPathway
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pw.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "Pneumonia Care Pathway" {
			t.Errorf("expected name=Pneumonia Care Pathway, got %s", fetched.Name)
		}
	})

	t.Run("Update", func(t *testing.T) {
		pw := createTestClinicalPathway(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
			pw.Name = "Updated Pneumonia Pathway"
			pw.Version = ptrStr("2.0")
			pw.ExpectedDuration = ptrStr("10 days")
			return repo.Update(ctx, pw)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *cds.ClinicalPathway
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, pw.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Name != "Updated Pneumonia Pathway" {
			t.Errorf("expected updated name, got %s", fetched.Name)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		pw := createTestClinicalPathway(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
			return repo.Delete(ctx, pw.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, pw.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted pathway")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*cds.ClinicalPathway
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 pathway")
		}
		_ = results
	})

	t.Run("AddPhase_GetPhases", func(t *testing.T) {
		pw := createTestClinicalPathway(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
			phase := &cds.ClinicalPathwayPhase{
				PathwayID:     pw.ID,
				Name:          "Day 1-2: Acute Phase",
				Description:   ptrStr("Initial stabilization and treatment"),
				Duration:      ptrStr("48 hours"),
				Goals:         ptrStr("Stabilize vitals, start antibiotics"),
				Interventions: ptrStr("IV antibiotics, respiratory support"),
				SortOrder:     1,
			}
			return repo.AddPhase(ctx, phase)
		})
		if err != nil {
			t.Fatalf("AddPhase: %v", err)
		}

		// Add second phase
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
			phase := &cds.ClinicalPathwayPhase{
				PathwayID:     pw.ID,
				Name:          "Day 3-5: Transition Phase",
				Description:   ptrStr("Transition to oral medications"),
				Duration:      ptrStr("72 hours"),
				Goals:         ptrStr("Switch to PO antibiotics"),
				Interventions: ptrStr("PO antibiotics, ambulation"),
				SortOrder:     2,
			}
			return repo.AddPhase(ctx, phase)
		})
		if err != nil {
			t.Fatalf("AddPhase (second): %v", err)
		}

		var phases []*cds.ClinicalPathwayPhase
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
			var err error
			phases, err = repo.GetPhases(ctx, pw.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetPhases: %v", err)
		}
		if len(phases) != 2 {
			t.Fatalf("expected 2 phases, got %d", len(phases))
		}
		if phases[0].SortOrder != 1 {
			t.Errorf("expected first phase sort_order=1, got %d", phases[0].SortOrder)
		}
	})
}

func TestFormularyCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("formulary")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	t.Run("Create", func(t *testing.T) {
		var created *cds.Formulary
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewFormularyRepoPG(globalDB.Pool)
			f := &cds.Formulary{
				Name:        "Hospital Formulary 2024",
				Description: ptrStr("Standard hospital formulary"),
				Version:     ptrStr("2024.1"),
				Active:      true,
			}
			if err := repo.Create(ctx, f); err != nil {
				return err
			}
			created = f
			return nil
		})
		if err != nil {
			t.Fatalf("Create formulary: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID_Update_Delete_List", func(t *testing.T) {
		form := createTestFormulary(t, ctx, tenantID)

		// GetByID
		var fetched *cds.Formulary
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewFormularyRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, form.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "Test Formulary" {
			t.Errorf("expected name=Test Formulary, got %s", fetched.Name)
		}

		// Update
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewFormularyRepoPG(globalDB.Pool)
			fetched.Name = "Updated Formulary"
			fetched.Active = false
			return repo.Update(ctx, fetched)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		// List
		var results []*cds.Formulary
		var total int
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewFormularyRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 formulary")
		}
		_ = results

		// Delete
		form2 := createTestFormulary(t, ctx, tenantID)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewFormularyRepoPG(globalDB.Pool)
			return repo.Delete(ctx, form2.ID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewFormularyRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, form2.ID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted formulary")
		}
	})

	t.Run("AddItem_GetItems", func(t *testing.T) {
		form := createTestFormulary(t, ctx, tenantID)

		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewFormularyRepoPG(globalDB.Pool)
			item := &cds.FormularyItem{
				FormularyID:       form.ID,
				MedicationName:    "Lisinopril 10mg",
				TierLevel:         ptrInt(1),
				RequiresPriorAuth: false,
				StepTherapyReq:    false,
				PreferredStatus:   ptrStr("preferred"),
				Note:              ptrStr("First-line ACE inhibitor"),
			}
			return repo.AddItem(ctx, item)
		})
		if err != nil {
			t.Fatalf("AddItem: %v", err)
		}

		var items []*cds.FormularyItem
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := cds.NewFormularyRepoPG(globalDB.Pool)
			var err error
			items, err = repo.GetItems(ctx, form.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetItems: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}
		if items[0].MedicationName != "Lisinopril 10mg" {
			t.Errorf("expected medication=Lisinopril 10mg, got %s", items[0].MedicationName)
		}
	})
}

// =========== Test Helpers ===========

func createTestCDSRule(t *testing.T, ctx context.Context, tenantID string) *cds.CDSRule {
	t.Helper()
	var result *cds.CDSRule
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := cds.NewCDSRuleRepoPG(globalDB.Pool)
		rule := &cds.CDSRule{
			RuleName:    "Drug Allergy Check",
			RuleType:    "drug-allergy",
			Description: ptrStr("Check for drug allergies before prescribing"),
			Severity:    ptrStr("high"),
			Category:    ptrStr("medication-safety"),
			Active:      true,
			Version:     ptrStr("1.0"),
		}
		if err := repo.Create(ctx, rule); err != nil {
			return err
		}
		result = rule
		return nil
	})
	if err != nil {
		t.Fatalf("create test CDS rule: %v", err)
	}
	return result
}

func createTestCDSAlert(t *testing.T, ctx context.Context, tenantID string, ruleID, patientID uuid.UUID) *cds.CDSAlert {
	t.Helper()
	var result *cds.CDSAlert
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := cds.NewCDSAlertRepoPG(globalDB.Pool)
		a := &cds.CDSAlert{
			RuleID:    ruleID,
			PatientID: patientID,
			Status:    "active",
			Severity:  ptrStr("warning"),
			Summary:   "Test alert for patient",
			FiredAt:   time.Now(),
		}
		if err := repo.Create(ctx, a); err != nil {
			return err
		}
		result = a
		return nil
	})
	if err != nil {
		t.Fatalf("create test CDS alert: %v", err)
	}
	return result
}

func createTestDrugInteraction(t *testing.T, ctx context.Context, tenantID string) *cds.DrugInteraction {
	t.Helper()
	var result *cds.DrugInteraction
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := cds.NewDrugInteractionRepoPG(globalDB.Pool)
		di := &cds.DrugInteraction{
			MedicationAName: "Metformin",
			MedicationBName: "Contrast Dye",
			Severity:        "moderate",
			Description:     ptrStr("Risk of lactic acidosis"),
			ClinicalEffect:  ptrStr("Metformin accumulation"),
			Management:      ptrStr("Hold metformin 48hrs before contrast"),
			Active:          true,
		}
		if err := repo.Create(ctx, di); err != nil {
			return err
		}
		result = di
		return nil
	})
	if err != nil {
		t.Fatalf("create test drug interaction: %v", err)
	}
	return result
}

func createTestOrderSet(t *testing.T, ctx context.Context, tenantID string) *cds.OrderSet {
	t.Helper()
	var result *cds.OrderSet
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := cds.NewOrderSetRepoPG(globalDB.Pool)
		os := &cds.OrderSet{
			Name:        "Admission Order Set",
			Description: ptrStr("Standard admission orders"),
			Category:    ptrStr("general"),
			Status:      "active",
			Version:     ptrStr("1.0"),
			Active:      true,
		}
		if err := repo.Create(ctx, os); err != nil {
			return err
		}
		result = os
		return nil
	})
	if err != nil {
		t.Fatalf("create test order set: %v", err)
	}
	return result
}

func createTestClinicalPathway(t *testing.T, ctx context.Context, tenantID string) *cds.ClinicalPathway {
	t.Helper()
	var result *cds.ClinicalPathway
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := cds.NewClinicalPathwayRepoPG(globalDB.Pool)
		p := &cds.ClinicalPathway{
			Name:             "Pneumonia Care Pathway",
			Description:      ptrStr("Standard pneumonia treatment pathway"),
			Condition:        ptrStr("pneumonia"),
			Category:         ptrStr("pulmonary"),
			Version:          ptrStr("1.0"),
			Active:           true,
			ExpectedDuration: ptrStr("7 days"),
		}
		if err := repo.Create(ctx, p); err != nil {
			return err
		}
		result = p
		return nil
	})
	if err != nil {
		t.Fatalf("create test clinical pathway: %v", err)
	}
	return result
}

func createTestFormulary(t *testing.T, ctx context.Context, tenantID string) *cds.Formulary {
	t.Helper()
	var result *cds.Formulary
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := cds.NewFormularyRepoPG(globalDB.Pool)
		f := &cds.Formulary{
			Name:        "Test Formulary",
			Description: ptrStr("Test formulary for integration tests"),
			Version:     ptrStr("1.0"),
			Active:      true,
		}
		if err := repo.Create(ctx, f); err != nil {
			return err
		}
		result = f
		return nil
	})
	if err != nil {
		t.Fatalf("create test formulary: %v", err)
	}
	return result
}
