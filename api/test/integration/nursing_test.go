package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ehr/ehr/internal/domain/nursing"
	"github.com/google/uuid"
)

func TestFlowsheetTemplateCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fstpl")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "TemplateNurse", "Adams")

	t.Run("Create", func(t *testing.T) {
		var created *nursing.FlowsheetTemplate
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			tpl := &nursing.FlowsheetTemplate{
				Name:        "Vital Signs Flowsheet",
				Description: ptrStr("Standard vital signs monitoring template"),
				Category:    ptrStr("vital-signs"),
				IsActive:    true,
				CreatedBy:   &practitioner.ID,
			}
			if err := repo.Create(ctx, tpl); err != nil {
				return err
			}
			created = tpl
			return nil
		})
		if err != nil {
			t.Fatalf("Create flowsheet template: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var tplID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			tpl := &nursing.FlowsheetTemplate{
				Name:     "Neurological Checks",
				Category: ptrStr("neuro"),
				IsActive: true,
			}
			if err := repo.Create(ctx, tpl); err != nil {
				return err
			}
			tplID = tpl.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *nursing.FlowsheetTemplate
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, tplID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Name != "Neurological Checks" {
			t.Errorf("expected name=Neurological Checks, got %s", fetched.Name)
		}
		if !fetched.IsActive {
			t.Error("expected is_active=true")
		}
	})

	t.Run("Update", func(t *testing.T) {
		var tpl *nursing.FlowsheetTemplate
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			t := &nursing.FlowsheetTemplate{
				Name:     "Draft Template",
				IsActive: false,
			}
			if err := repo.Create(ctx, t); err != nil {
				return err
			}
			tpl = t
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			tpl.Name = "Finalized Template"
			tpl.Description = ptrStr("Now active and finalized")
			tpl.IsActive = true
			return repo.Update(ctx, tpl)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *nursing.FlowsheetTemplate
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, tpl.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Name != "Finalized Template" {
			t.Errorf("expected name=Finalized Template, got %s", fetched.Name)
		}
		if !fetched.IsActive {
			t.Error("expected is_active=true after update")
		}
	})

	t.Run("List", func(t *testing.T) {
		var results []*nursing.FlowsheetTemplate
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.List(ctx, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 template")
		}
		_ = results
	})

	t.Run("AddRow_GetRows", func(t *testing.T) {
		var tplID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			tpl := &nursing.FlowsheetTemplate{
				Name:     "Row Test Template",
				IsActive: true,
			}
			if err := repo.Create(ctx, tpl); err != nil {
				return err
			}
			tplID = tpl.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create template: %v", err)
		}

		// Add rows
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			row1 := &nursing.FlowsheetRow{
				TemplateID: tplID,
				Label:      "Heart Rate",
				DataType:   "numeric",
				Unit:       ptrStr("bpm"),
				SortOrder:  1,
				IsRequired: true,
			}
			if err := repo.AddRow(ctx, row1); err != nil {
				return err
			}

			row2 := &nursing.FlowsheetRow{
				TemplateID: tplID,
				Label:      "Blood Pressure",
				DataType:   "text",
				SortOrder:  2,
				IsRequired: true,
			}
			return repo.AddRow(ctx, row2)
		})
		if err != nil {
			t.Fatalf("AddRow: %v", err)
		}

		var rows []*nursing.FlowsheetRow
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			var err error
			rows, err = repo.GetRows(ctx, tplID)
			return err
		})
		if err != nil {
			t.Fatalf("GetRows: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("expected 2 rows, got %d", len(rows))
		}
		if rows[0].Label != "Heart Rate" {
			t.Errorf("expected first row label=Heart Rate, got %s", rows[0].Label)
		}
		if rows[1].Label != "Blood Pressure" {
			t.Errorf("expected second row label=Blood Pressure, got %s", rows[1].Label)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var tplID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			tpl := &nursing.FlowsheetTemplate{
				Name:     "Delete Test Template",
				IsActive: false,
			}
			if err := repo.Create(ctx, tpl); err != nil {
				return err
			}
			tplID = tpl.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			return repo.Delete(ctx, tplID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, tplID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted template")
		}
	})
}

func TestFlowsheetEntryCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fsentry")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "EntryPatient", "Test", "MRN-ENTRY-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "EntryNurse", "Brown")
	enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

	// Create a template and row first
	var tplID, rowID uuid.UUID
	err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
		repo := nursing.NewFlowsheetTemplateRepoPG(globalDB.Pool)
		tpl := &nursing.FlowsheetTemplate{
			Name:     "Entry Test Template",
			IsActive: true,
		}
		if err := repo.Create(ctx, tpl); err != nil {
			return err
		}
		tplID = tpl.ID

		row := &nursing.FlowsheetRow{
			TemplateID: tplID,
			Label:      "Temperature",
			DataType:   "numeric",
			Unit:       ptrStr("F"),
			SortOrder:  1,
			IsRequired: true,
		}
		if err := repo.AddRow(ctx, row); err != nil {
			return err
		}
		rowID = row.ID
		return nil
	})
	if err != nil {
		t.Fatalf("setup template and row: %v", err)
	}

	t.Run("Create", func(t *testing.T) {
		var created *nursing.FlowsheetEntry
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetEntryRepoPG(globalDB.Pool)
			e := &nursing.FlowsheetEntry{
				TemplateID:   tplID,
				RowID:        rowID,
				PatientID:    patient.ID,
				EncounterID:  enc.ID,
				ValueNumeric: ptrFloat(98.6),
				RecordedAt:   now,
				RecordedByID: practitioner.ID,
				Note:         ptrStr("Oral temperature"),
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			created = e
			return nil
		})
		if err != nil {
			t.Fatalf("Create flowsheet entry: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var entryID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetEntryRepoPG(globalDB.Pool)
			e := &nursing.FlowsheetEntry{
				TemplateID:   tplID,
				RowID:        rowID,
				PatientID:    patient.ID,
				EncounterID:  enc.ID,
				ValueNumeric: ptrFloat(99.1),
				RecordedAt:   now,
				RecordedByID: practitioner.ID,
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			entryID = e.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *nursing.FlowsheetEntry
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetEntryRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, entryID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.ValueNumeric == nil || *fetched.ValueNumeric != 99.1 {
			t.Errorf("expected value_numeric=99.1, got %v", fetched.ValueNumeric)
		}
		if fetched.PatientID != patient.ID {
			t.Errorf("expected patient_id=%s, got %s", patient.ID, fetched.PatientID)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*nursing.FlowsheetEntry
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetEntryRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 entry")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("ListByEncounter", func(t *testing.T) {
		var results []*nursing.FlowsheetEntry
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetEntryRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByEncounter(ctx, enc.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByEncounter: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 entry")
		}
		for _, r := range results {
			if r.EncounterID != enc.ID {
				t.Errorf("expected encounter_id=%s, got %s", enc.ID, r.EncounterID)
			}
		}
	})

	t.Run("Search", func(t *testing.T) {
		var results []*nursing.FlowsheetEntry
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetEntryRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.Search(ctx, map[string]string{
				"patient_id":  patient.ID.String(),
				"template_id": tplID.String(),
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
			if r.TemplateID != tplID {
				t.Errorf("expected template_id=%s, got %s", tplID, r.TemplateID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var entryID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetEntryRepoPG(globalDB.Pool)
			e := &nursing.FlowsheetEntry{
				TemplateID:   tplID,
				RowID:        rowID,
				PatientID:    patient.ID,
				EncounterID:  enc.ID,
				ValueText:    ptrStr("delete-test"),
				RecordedAt:   now,
				RecordedByID: practitioner.ID,
			}
			if err := repo.Create(ctx, e); err != nil {
				return err
			}
			entryID = e.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetEntryRepoPG(globalDB.Pool)
			return repo.Delete(ctx, entryID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFlowsheetEntryRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, entryID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted entry")
		}
	})
}

func TestNursingAssessmentCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("nassess")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "AssessPatient", "Test", "MRN-NASSESS-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "AssessNurse", "Clark")
	enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

	t.Run("Create", func(t *testing.T) {
		var created *nursing.NursingAssessment
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
			a := &nursing.NursingAssessment{
				PatientID:      patient.ID,
				EncounterID:    enc.ID,
				NurseID:        practitioner.ID,
				AssessmentType: "admission",
				AssessmentData: ptrStr(`{"general": "alert, oriented"}`),
				Status:         "in-progress",
				Note:           ptrStr("Admission nursing assessment"),
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create nursing assessment: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var assessID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
			a := &nursing.NursingAssessment{
				PatientID:      patient.ID,
				EncounterID:    enc.ID,
				NurseID:        practitioner.ID,
				AssessmentType: "shift",
				Status:         "completed",
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

		var fetched *nursing.NursingAssessment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, assessID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.AssessmentType != "shift" {
			t.Errorf("expected assessment_type=shift, got %s", fetched.AssessmentType)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var assess *nursing.NursingAssessment
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
			a := &nursing.NursingAssessment{
				PatientID:      patient.ID,
				EncounterID:    enc.ID,
				NurseID:        practitioner.ID,
				AssessmentType: "focused",
				Status:         "in-progress",
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

		completedAt := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
			assess.Status = "completed"
			assess.CompletedAt = &completedAt
			assess.AssessmentData = ptrStr(`{"respiratory": "clear bilateral"}`)
			assess.Note = ptrStr("Assessment completed")
			return repo.Update(ctx, assess)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *nursing.NursingAssessment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, assess.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "completed" {
			t.Errorf("expected status=completed, got %s", fetched.Status)
		}
		if fetched.CompletedAt == nil {
			t.Error("expected non-nil CompletedAt")
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*nursing.NursingAssessment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
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

	t.Run("ListByEncounter", func(t *testing.T) {
		var results []*nursing.NursingAssessment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByEncounter(ctx, enc.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByEncounter: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 assessment")
		}
		for _, r := range results {
			if r.EncounterID != enc.ID {
				t.Errorf("expected encounter_id=%s, got %s", enc.ID, r.EncounterID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var assessID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
			a := &nursing.NursingAssessment{
				PatientID:      patient.ID,
				EncounterID:    enc.ID,
				NurseID:        practitioner.ID,
				AssessmentType: "delete-test",
				Status:         "draft",
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
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
			return repo.Delete(ctx, assessID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewNursingAssessmentRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, assessID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted assessment")
		}
	})
}

func TestFallRiskAssessmentCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("fallrisk")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "FallPatient", "Test", "MRN-FALL-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "FallNurse", "Davis")

	t.Run("Create", func(t *testing.T) {
		var created *nursing.FallRiskAssessment
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFallRiskRepoPG(globalDB.Pool)
			a := &nursing.FallRiskAssessment{
				PatientID:      patient.ID,
				AssessedByID:   practitioner.ID,
				ToolUsed:       ptrStr("Morse Fall Scale"),
				TotalScore:     ptrInt(55),
				RiskLevel:      ptrStr("high"),
				HistoryOfFalls: ptrBool(true),
				Medications:    ptrBool(true),
				GaitBalance:    ptrStr("unsteady"),
				MentalStatus:   ptrStr("forgets limitations"),
				Interventions:  ptrStr("Bed alarm, non-slip socks, call light within reach"),
				Note:           ptrStr("High fall risk - implement fall prevention bundle"),
				AssessedAt:     now,
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create fall risk assessment: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var assessID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFallRiskRepoPG(globalDB.Pool)
			a := &nursing.FallRiskAssessment{
				PatientID:    patient.ID,
				AssessedByID: practitioner.ID,
				TotalScore:   ptrInt(25),
				RiskLevel:    ptrStr("low"),
				AssessedAt:   now,
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

		var fetched *nursing.FallRiskAssessment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFallRiskRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, assessID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.RiskLevel == nil || *fetched.RiskLevel != "low" {
			t.Errorf("expected risk_level=low, got %v", fetched.RiskLevel)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*nursing.FallRiskAssessment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewFallRiskRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 fall risk assessment")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})
}

func TestSkinAssessmentCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("skin")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "SkinPatient", "Test", "MRN-SKIN-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "SkinNurse", "Evans")

	t.Run("Create", func(t *testing.T) {
		var created *nursing.SkinAssessment
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewSkinAssessmentRepoPG(globalDB.Pool)
			a := &nursing.SkinAssessment{
				PatientID:     patient.ID,
				AssessedByID:  practitioner.ID,
				ToolUsed:      ptrStr("Braden Scale"),
				TotalScore:    ptrInt(14),
				RiskLevel:     ptrStr("moderate"),
				SkinIntegrity: ptrStr("intact"),
				MoistureLevel: ptrStr("occasionally moist"),
				Mobility:      ptrStr("slightly limited"),
				Nutrition:     ptrStr("adequate"),
				WoundPresent:  ptrBool(false),
				Interventions: ptrStr("Reposition q2h, moisture barrier cream"),
				Note:          ptrStr("Moderate pressure injury risk"),
				AssessedAt:    now,
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create skin assessment: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var assessID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewSkinAssessmentRepoPG(globalDB.Pool)
			a := &nursing.SkinAssessment{
				PatientID:    patient.ID,
				AssessedByID: practitioner.ID,
				WoundPresent: ptrBool(true),
				WoundLocation: ptrStr("sacrum"),
				WoundStage:   ptrStr("stage 2"),
				AssessedAt:   now,
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

		var fetched *nursing.SkinAssessment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewSkinAssessmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, assessID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.WoundPresent == nil || !*fetched.WoundPresent {
			t.Error("expected wound_present=true")
		}
		if fetched.WoundLocation == nil || *fetched.WoundLocation != "sacrum" {
			t.Errorf("expected wound_location=sacrum, got %v", fetched.WoundLocation)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*nursing.SkinAssessment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewSkinAssessmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 skin assessment")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})
}

func TestPainAssessmentCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("pain")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "PainPatient", "Test", "MRN-PAIN-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "PainNurse", "Foster")

	t.Run("Create", func(t *testing.T) {
		var created *nursing.PainAssessment
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewPainAssessmentRepoPG(globalDB.Pool)
			a := &nursing.PainAssessment{
				PatientID:     patient.ID,
				AssessedByID:  practitioner.ID,
				ToolUsed:      ptrStr("Numeric Rating Scale"),
				PainScore:     ptrInt(7),
				PainLocation:  ptrStr("lower back"),
				PainCharacter: ptrStr("sharp, stabbing"),
				PainDuration:  ptrStr("constant"),
				PainRadiation: ptrStr("down left leg"),
				Aggravating:   ptrStr("bending, lifting"),
				Alleviating:   ptrStr("lying flat, ice"),
				Interventions: ptrStr("PRN medication administered, repositioned"),
				ReassessScore: ptrInt(4),
				Note:          ptrStr("Pain reassessed 30 min post medication"),
				AssessedAt:    now,
			}
			if err := repo.Create(ctx, a); err != nil {
				return err
			}
			created = a
			return nil
		})
		if err != nil {
			t.Fatalf("Create pain assessment: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var assessID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewPainAssessmentRepoPG(globalDB.Pool)
			a := &nursing.PainAssessment{
				PatientID:    patient.ID,
				AssessedByID: practitioner.ID,
				PainScore:    ptrInt(3),
				PainLocation: ptrStr("right knee"),
				AssessedAt:   now,
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

		var fetched *nursing.PainAssessment
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewPainAssessmentRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, assessID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.PainScore == nil || *fetched.PainScore != 3 {
			t.Errorf("expected pain_score=3, got %v", fetched.PainScore)
		}
		if fetched.PainLocation == nil || *fetched.PainLocation != "right knee" {
			t.Errorf("expected pain_location=right knee, got %v", fetched.PainLocation)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*nursing.PainAssessment
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewPainAssessmentRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 pain assessment")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})
}

func TestLinesDrainsCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("lines")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "LinesPatient", "Test", "MRN-LINES-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "LinesNurse", "Garcia")
	enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

	t.Run("Create", func(t *testing.T) {
		var created *nursing.LinesDrainsAirways
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			l := &nursing.LinesDrainsAirways{
				PatientID:    patient.ID,
				EncounterID:  enc.ID,
				Type:         "IV",
				Description:  ptrStr("Peripheral IV"),
				Site:         ptrStr("Right antecubital"),
				Size:         ptrStr("20 gauge"),
				InsertedAt:   &now,
				InsertedByID: &practitioner.ID,
				Status:       "active",
				Note:         ptrStr("Patent, no redness or swelling"),
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			created = l
			return nil
		})
		if err != nil {
			t.Fatalf("Create lines/drains: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		var lineID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			l := &nursing.LinesDrainsAirways{
				PatientID:   patient.ID,
				EncounterID: enc.ID,
				Type:        "Foley",
				Description: ptrStr("Indwelling urinary catheter"),
				Status:      "active",
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			lineID = l.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *nursing.LinesDrainsAirways
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, lineID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Type != "Foley" {
			t.Errorf("expected type=Foley, got %s", fetched.Type)
		}
		if fetched.Status != "active" {
			t.Errorf("expected status=active, got %s", fetched.Status)
		}
	})

	t.Run("Update", func(t *testing.T) {
		var line *nursing.LinesDrainsAirways
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			l := &nursing.LinesDrainsAirways{
				PatientID:   patient.ID,
				EncounterID: enc.ID,
				Type:        "NG tube",
				Description: ptrStr("Nasogastric tube"),
				Status:      "active",
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			line = l
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		removedAt := time.Now()
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			line.Status = "removed"
			line.RemovedAt = &removedAt
			line.RemovedByID = &practitioner.ID
			line.Note = ptrStr("Removed per physician order")
			return repo.Update(ctx, line)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *nursing.LinesDrainsAirways
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, line.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.Status != "removed" {
			t.Errorf("expected status=removed, got %s", fetched.Status)
		}
		if fetched.RemovedAt == nil {
			t.Error("expected non-nil RemovedAt")
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*nursing.LinesDrainsAirways
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 line/drain")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("ListByEncounter", func(t *testing.T) {
		var results []*nursing.LinesDrainsAirways
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByEncounter(ctx, enc.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByEncounter: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 line/drain")
		}
		for _, r := range results {
			if r.EncounterID != enc.ID {
				t.Errorf("expected encounter_id=%s, got %s", enc.ID, r.EncounterID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		var lineID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			l := &nursing.LinesDrainsAirways{
				PatientID:   patient.ID,
				EncounterID: enc.ID,
				Type:        "JP drain",
				Status:      "active",
			}
			if err := repo.Create(ctx, l); err != nil {
				return err
			}
			lineID = l.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			return repo.Delete(ctx, lineID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewLinesDrainsRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, lineID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted line/drain")
		}
	})
}

func TestRestraintRecordCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("restraint")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "RestraintPatient", "Test", "MRN-REST-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "RestraintNurse", "Hall")

	t.Run("Create", func(t *testing.T) {
		var created *nursing.RestraintRecord
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewRestraintRepoPG(globalDB.Pool)
			r := &nursing.RestraintRecord{
				PatientID:     patient.ID,
				RestraintType: "soft wrist",
				Reason:        ptrStr("Pulling at IV and ETT"),
				BodySite:      ptrStr("bilateral wrists"),
				AppliedAt:     now,
				AppliedByID:   practitioner.ID,
				Note:          ptrStr("Applied per physician order"),
			}
			if err := repo.Create(ctx, r); err != nil {
				return err
			}
			created = r
			return nil
		})
		if err != nil {
			t.Fatalf("Create restraint record: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var recID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewRestraintRepoPG(globalDB.Pool)
			r := &nursing.RestraintRecord{
				PatientID:     patient.ID,
				RestraintType: "mitt",
				BodySite:      ptrStr("bilateral hands"),
				AppliedAt:     now,
				AppliedByID:   practitioner.ID,
			}
			if err := repo.Create(ctx, r); err != nil {
				return err
			}
			recID = r.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *nursing.RestraintRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewRestraintRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, recID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.RestraintType != "mitt" {
			t.Errorf("expected restraint_type=mitt, got %s", fetched.RestraintType)
		}
	})

	t.Run("Update", func(t *testing.T) {
		now := time.Now()
		var rec *nursing.RestraintRecord
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewRestraintRepoPG(globalDB.Pool)
			r := &nursing.RestraintRecord{
				PatientID:     patient.ID,
				RestraintType: "soft wrist",
				AppliedAt:     now,
				AppliedByID:   practitioner.ID,
			}
			if err := repo.Create(ctx, r); err != nil {
				return err
			}
			rec = r
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		removedAt := now.Add(4 * time.Hour)
		assessedAt := now.Add(2 * time.Hour)
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewRestraintRepoPG(globalDB.Pool)
			rec.RemovedAt = &removedAt
			rec.RemovedByID = &practitioner.ID
			rec.LastAssessedAt = &assessedAt
			rec.LastAssessedByID = &practitioner.ID
			rec.SkinCondition = ptrStr("intact, no erythema")
			rec.Circulation = ptrStr("CMS intact bilaterally")
			rec.Note = ptrStr("Removed per order, skin intact")
			return repo.Update(ctx, rec)
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}

		var fetched *nursing.RestraintRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewRestraintRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, rec.ID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID after update: %v", err)
		}
		if fetched.RemovedAt == nil {
			t.Error("expected non-nil RemovedAt")
		}
		if fetched.SkinCondition == nil || *fetched.SkinCondition != "intact, no erythema" {
			t.Errorf("expected skin_condition updated, got %v", fetched.SkinCondition)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*nursing.RestraintRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewRestraintRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 restraint record")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})
}

func TestIntakeOutputCRUD(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueTenantID("io")
	createTenantSchema(t, ctx, tenantID)
	defer dropTenantSchema(t, ctx, tenantID)

	patient := createTestPatient(t, ctx, globalDB.Pool, tenantID, "IOPatient", "Test", "MRN-IO-001")
	practitioner := createTestPractitioner(t, ctx, globalDB.Pool, tenantID, "IONurse", "Irving")
	enc := createTestEncounter(t, ctx, globalDB.Pool, tenantID, patient.ID, &practitioner.ID)

	t.Run("Create", func(t *testing.T) {
		var created *nursing.IntakeOutputRecord
		now := time.Now()
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewIntakeOutputRepoPG(globalDB.Pool)
			r := &nursing.IntakeOutputRecord{
				PatientID:    patient.ID,
				EncounterID:  enc.ID,
				Category:     "intake",
				Type:         ptrStr("IV fluid"),
				Volume:       ptrFloat(1000),
				Unit:         ptrStr("mL"),
				Route:        ptrStr("IV"),
				RecordedAt:   now,
				RecordedByID: practitioner.ID,
				Note:         ptrStr("NS 1000mL infused over 8 hours"),
			}
			if err := repo.Create(ctx, r); err != nil {
				return err
			}
			created = r
			return nil
		})
		if err != nil {
			t.Fatalf("Create intake/output: %v", err)
		}
		if created.ID == uuid.Nil {
			t.Fatal("expected non-nil ID")
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		now := time.Now()
		var recID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewIntakeOutputRepoPG(globalDB.Pool)
			r := &nursing.IntakeOutputRecord{
				PatientID:    patient.ID,
				EncounterID:  enc.ID,
				Category:     "output",
				Type:         ptrStr("urine"),
				Volume:       ptrFloat(400),
				Unit:         ptrStr("mL"),
				RecordedAt:   now,
				RecordedByID: practitioner.ID,
			}
			if err := repo.Create(ctx, r); err != nil {
				return err
			}
			recID = r.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		var fetched *nursing.IntakeOutputRecord
		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewIntakeOutputRepoPG(globalDB.Pool)
			var err error
			fetched, err = repo.GetByID(ctx, recID)
			return err
		})
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if fetched.Category != "output" {
			t.Errorf("expected category=output, got %s", fetched.Category)
		}
		if fetched.Volume == nil || *fetched.Volume != 400 {
			t.Errorf("expected volume=400, got %v", fetched.Volume)
		}
	})

	t.Run("ListByPatient", func(t *testing.T) {
		var results []*nursing.IntakeOutputRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewIntakeOutputRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByPatient(ctx, patient.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByPatient: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 I&O record")
		}
		for _, r := range results {
			if r.PatientID != patient.ID {
				t.Errorf("expected patient_id=%s, got %s", patient.ID, r.PatientID)
			}
		}
	})

	t.Run("ListByEncounter", func(t *testing.T) {
		var results []*nursing.IntakeOutputRecord
		var total int
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewIntakeOutputRepoPG(globalDB.Pool)
			var err error
			results, total, err = repo.ListByEncounter(ctx, enc.ID, 100, 0)
			return err
		})
		if err != nil {
			t.Fatalf("ListByEncounter: %v", err)
		}
		if total == 0 {
			t.Error("expected at least 1 I&O record")
		}
		for _, r := range results {
			if r.EncounterID != enc.ID {
				t.Errorf("expected encounter_id=%s, got %s", enc.ID, r.EncounterID)
			}
		}
	})

	t.Run("Delete", func(t *testing.T) {
		now := time.Now()
		var recID uuid.UUID
		err := withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewIntakeOutputRepoPG(globalDB.Pool)
			r := &nursing.IntakeOutputRecord{
				PatientID:    patient.ID,
				EncounterID:  enc.ID,
				Category:     "intake",
				Type:         ptrStr("PO fluid"),
				Volume:       ptrFloat(240),
				Unit:         ptrStr("mL"),
				RecordedAt:   now,
				RecordedByID: practitioner.ID,
			}
			if err := repo.Create(ctx, r); err != nil {
				return err
			}
			recID = r.ID
			return nil
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewIntakeOutputRepoPG(globalDB.Pool)
			return repo.Delete(ctx, recID)
		})
		if err != nil {
			t.Fatalf("Delete: %v", err)
		}

		err = withTenantConn(ctx, globalDB.Pool, tenantID, func(ctx context.Context) error {
			repo := nursing.NewIntakeOutputRepoPG(globalDB.Pool)
			_, err := repo.GetByID(ctx, recID)
			return err
		})
		if err == nil {
			t.Fatal("expected error getting deleted I&O record")
		}
	})
}
