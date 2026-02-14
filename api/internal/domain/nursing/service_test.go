package nursing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockFlowsheetTemplateRepo struct {
	records map[uuid.UUID]*FlowsheetTemplate
	rows    map[uuid.UUID]*FlowsheetRow
}

func newMockFlowsheetTemplateRepo() *mockFlowsheetTemplateRepo {
	return &mockFlowsheetTemplateRepo{
		records: make(map[uuid.UUID]*FlowsheetTemplate),
		rows:    make(map[uuid.UUID]*FlowsheetRow),
	}
}

func (m *mockFlowsheetTemplateRepo) Create(_ context.Context, t *FlowsheetTemplate) error {
	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	m.records[t.ID] = t
	return nil
}

func (m *mockFlowsheetTemplateRepo) GetByID(_ context.Context, id uuid.UUID) (*FlowsheetTemplate, error) {
	t, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

func (m *mockFlowsheetTemplateRepo) Update(_ context.Context, t *FlowsheetTemplate) error {
	m.records[t.ID] = t
	return nil
}

func (m *mockFlowsheetTemplateRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockFlowsheetTemplateRepo) List(_ context.Context, limit, offset int) ([]*FlowsheetTemplate, int, error) {
	var result []*FlowsheetTemplate
	for _, t := range m.records {
		result = append(result, t)
	}
	return result, len(result), nil
}

func (m *mockFlowsheetTemplateRepo) AddRow(_ context.Context, r *FlowsheetRow) error {
	r.ID = uuid.New()
	m.rows[r.ID] = r
	return nil
}

func (m *mockFlowsheetTemplateRepo) GetRows(_ context.Context, templateID uuid.UUID) ([]*FlowsheetRow, error) {
	var result []*FlowsheetRow
	for _, r := range m.rows {
		if r.TemplateID == templateID {
			result = append(result, r)
		}
	}
	return result, nil
}

type mockFlowsheetEntryRepo struct {
	records map[uuid.UUID]*FlowsheetEntry
}

func newMockFlowsheetEntryRepo() *mockFlowsheetEntryRepo {
	return &mockFlowsheetEntryRepo{records: make(map[uuid.UUID]*FlowsheetEntry)}
}

func (m *mockFlowsheetEntryRepo) Create(_ context.Context, e *FlowsheetEntry) error {
	e.ID = uuid.New()
	e.CreatedAt = time.Now()
	m.records[e.ID] = e
	return nil
}

func (m *mockFlowsheetEntryRepo) GetByID(_ context.Context, id uuid.UUID) (*FlowsheetEntry, error) {
	e, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return e, nil
}

func (m *mockFlowsheetEntryRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockFlowsheetEntryRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*FlowsheetEntry, int, error) {
	var result []*FlowsheetEntry
	for _, e := range m.records {
		if e.PatientID == patientID {
			result = append(result, e)
		}
	}
	return result, len(result), nil
}

func (m *mockFlowsheetEntryRepo) ListByEncounter(_ context.Context, encounterID uuid.UUID, limit, offset int) ([]*FlowsheetEntry, int, error) {
	var result []*FlowsheetEntry
	for _, e := range m.records {
		if e.EncounterID == encounterID {
			result = append(result, e)
		}
	}
	return result, len(result), nil
}

func (m *mockFlowsheetEntryRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*FlowsheetEntry, int, error) {
	var result []*FlowsheetEntry
	for _, e := range m.records {
		result = append(result, e)
	}
	return result, len(result), nil
}

type mockNursingAssessmentRepo struct {
	records map[uuid.UUID]*NursingAssessment
}

func newMockNursingAssessmentRepo() *mockNursingAssessmentRepo {
	return &mockNursingAssessmentRepo{records: make(map[uuid.UUID]*NursingAssessment)}
}

func (m *mockNursingAssessmentRepo) Create(_ context.Context, a *NursingAssessment) error {
	a.ID = uuid.New()
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	m.records[a.ID] = a
	return nil
}

func (m *mockNursingAssessmentRepo) GetByID(_ context.Context, id uuid.UUID) (*NursingAssessment, error) {
	a, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return a, nil
}

func (m *mockNursingAssessmentRepo) Update(_ context.Context, a *NursingAssessment) error {
	m.records[a.ID] = a
	return nil
}

func (m *mockNursingAssessmentRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockNursingAssessmentRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*NursingAssessment, int, error) {
	var result []*NursingAssessment
	for _, a := range m.records {
		if a.PatientID == patientID {
			result = append(result, a)
		}
	}
	return result, len(result), nil
}

func (m *mockNursingAssessmentRepo) ListByEncounter(_ context.Context, encounterID uuid.UUID, limit, offset int) ([]*NursingAssessment, int, error) {
	var result []*NursingAssessment
	for _, a := range m.records {
		if a.EncounterID == encounterID {
			result = append(result, a)
		}
	}
	return result, len(result), nil
}

type mockFallRiskRepo struct {
	records map[uuid.UUID]*FallRiskAssessment
}

func newMockFallRiskRepo() *mockFallRiskRepo {
	return &mockFallRiskRepo{records: make(map[uuid.UUID]*FallRiskAssessment)}
}

func (m *mockFallRiskRepo) Create(_ context.Context, a *FallRiskAssessment) error {
	a.ID = uuid.New()
	a.CreatedAt = time.Now()
	m.records[a.ID] = a
	return nil
}

func (m *mockFallRiskRepo) GetByID(_ context.Context, id uuid.UUID) (*FallRiskAssessment, error) {
	a, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return a, nil
}

func (m *mockFallRiskRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*FallRiskAssessment, int, error) {
	var result []*FallRiskAssessment
	for _, a := range m.records {
		if a.PatientID == patientID {
			result = append(result, a)
		}
	}
	return result, len(result), nil
}

type mockSkinAssessmentRepo struct {
	records map[uuid.UUID]*SkinAssessment
}

func newMockSkinAssessmentRepo() *mockSkinAssessmentRepo {
	return &mockSkinAssessmentRepo{records: make(map[uuid.UUID]*SkinAssessment)}
}

func (m *mockSkinAssessmentRepo) Create(_ context.Context, a *SkinAssessment) error {
	a.ID = uuid.New()
	a.CreatedAt = time.Now()
	m.records[a.ID] = a
	return nil
}

func (m *mockSkinAssessmentRepo) GetByID(_ context.Context, id uuid.UUID) (*SkinAssessment, error) {
	a, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return a, nil
}

func (m *mockSkinAssessmentRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*SkinAssessment, int, error) {
	var result []*SkinAssessment
	for _, a := range m.records {
		if a.PatientID == patientID {
			result = append(result, a)
		}
	}
	return result, len(result), nil
}

type mockPainAssessmentRepo struct {
	records map[uuid.UUID]*PainAssessment
}

func newMockPainAssessmentRepo() *mockPainAssessmentRepo {
	return &mockPainAssessmentRepo{records: make(map[uuid.UUID]*PainAssessment)}
}

func (m *mockPainAssessmentRepo) Create(_ context.Context, a *PainAssessment) error {
	a.ID = uuid.New()
	a.CreatedAt = time.Now()
	m.records[a.ID] = a
	return nil
}

func (m *mockPainAssessmentRepo) GetByID(_ context.Context, id uuid.UUID) (*PainAssessment, error) {
	a, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return a, nil
}

func (m *mockPainAssessmentRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*PainAssessment, int, error) {
	var result []*PainAssessment
	for _, a := range m.records {
		if a.PatientID == patientID {
			result = append(result, a)
		}
	}
	return result, len(result), nil
}

type mockLinesDrainsRepo struct {
	records map[uuid.UUID]*LinesDrainsAirways
}

func newMockLinesDrainsRepo() *mockLinesDrainsRepo {
	return &mockLinesDrainsRepo{records: make(map[uuid.UUID]*LinesDrainsAirways)}
}

func (m *mockLinesDrainsRepo) Create(_ context.Context, l *LinesDrainsAirways) error {
	l.ID = uuid.New()
	l.CreatedAt = time.Now()
	l.UpdatedAt = time.Now()
	m.records[l.ID] = l
	return nil
}

func (m *mockLinesDrainsRepo) GetByID(_ context.Context, id uuid.UUID) (*LinesDrainsAirways, error) {
	l, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return l, nil
}

func (m *mockLinesDrainsRepo) Update(_ context.Context, l *LinesDrainsAirways) error {
	m.records[l.ID] = l
	return nil
}

func (m *mockLinesDrainsRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockLinesDrainsRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*LinesDrainsAirways, int, error) {
	var result []*LinesDrainsAirways
	for _, l := range m.records {
		if l.PatientID == patientID {
			result = append(result, l)
		}
	}
	return result, len(result), nil
}

func (m *mockLinesDrainsRepo) ListByEncounter(_ context.Context, encounterID uuid.UUID, limit, offset int) ([]*LinesDrainsAirways, int, error) {
	var result []*LinesDrainsAirways
	for _, l := range m.records {
		if l.EncounterID == encounterID {
			result = append(result, l)
		}
	}
	return result, len(result), nil
}

type mockRestraintRepo struct {
	records map[uuid.UUID]*RestraintRecord
}

func newMockRestraintRepo() *mockRestraintRepo {
	return &mockRestraintRepo{records: make(map[uuid.UUID]*RestraintRecord)}
}

func (m *mockRestraintRepo) Create(_ context.Context, r *RestraintRecord) error {
	r.ID = uuid.New()
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	m.records[r.ID] = r
	return nil
}

func (m *mockRestraintRepo) GetByID(_ context.Context, id uuid.UUID) (*RestraintRecord, error) {
	r, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (m *mockRestraintRepo) Update(_ context.Context, r *RestraintRecord) error {
	m.records[r.ID] = r
	return nil
}

func (m *mockRestraintRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*RestraintRecord, int, error) {
	var result []*RestraintRecord
	for _, r := range m.records {
		if r.PatientID == patientID {
			result = append(result, r)
		}
	}
	return result, len(result), nil
}

type mockIntakeOutputRepo struct {
	records map[uuid.UUID]*IntakeOutputRecord
}

func newMockIntakeOutputRepo() *mockIntakeOutputRepo {
	return &mockIntakeOutputRepo{records: make(map[uuid.UUID]*IntakeOutputRecord)}
}

func (m *mockIntakeOutputRepo) Create(_ context.Context, r *IntakeOutputRecord) error {
	r.ID = uuid.New()
	r.CreatedAt = time.Now()
	m.records[r.ID] = r
	return nil
}

func (m *mockIntakeOutputRepo) GetByID(_ context.Context, id uuid.UUID) (*IntakeOutputRecord, error) {
	r, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (m *mockIntakeOutputRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockIntakeOutputRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*IntakeOutputRecord, int, error) {
	var result []*IntakeOutputRecord
	for _, r := range m.records {
		if r.PatientID == patientID {
			result = append(result, r)
		}
	}
	return result, len(result), nil
}

func (m *mockIntakeOutputRepo) ListByEncounter(_ context.Context, encounterID uuid.UUID, limit, offset int) ([]*IntakeOutputRecord, int, error) {
	var result []*IntakeOutputRecord
	for _, r := range m.records {
		if r.EncounterID == encounterID {
			result = append(result, r)
		}
	}
	return result, len(result), nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(
		newMockFlowsheetTemplateRepo(),
		newMockFlowsheetEntryRepo(),
		newMockNursingAssessmentRepo(),
		newMockFallRiskRepo(),
		newMockSkinAssessmentRepo(),
		newMockPainAssessmentRepo(),
		newMockLinesDrainsRepo(),
		newMockRestraintRepo(),
		newMockIntakeOutputRepo(),
	)
}

// -- Flowsheet Template Tests --

func TestCreateTemplate(t *testing.T) {
	svc := newTestService()
	tmpl := &FlowsheetTemplate{Name: "Vitals"}
	err := svc.CreateTemplate(context.Background(), tmpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestCreateTemplate_NameRequired(t *testing.T) {
	svc := newTestService()
	tmpl := &FlowsheetTemplate{}
	err := svc.CreateTemplate(context.Background(), tmpl)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestGetTemplate(t *testing.T) {
	svc := newTestService()
	tmpl := &FlowsheetTemplate{Name: "Vitals"}
	svc.CreateTemplate(context.Background(), tmpl)
	fetched, err := svc.GetTemplate(context.Background(), tmpl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.Name != "Vitals" {
		t.Errorf("expected name 'Vitals', got %s", fetched.Name)
	}
}

func TestDeleteTemplate(t *testing.T) {
	svc := newTestService()
	tmpl := &FlowsheetTemplate{Name: "Vitals"}
	svc.CreateTemplate(context.Background(), tmpl)
	err := svc.DeleteTemplate(context.Background(), tmpl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetTemplate(context.Background(), tmpl.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestAddTemplateRow(t *testing.T) {
	svc := newTestService()
	r := &FlowsheetRow{TemplateID: uuid.New(), Label: "Heart Rate"}
	err := svc.AddTemplateRow(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestAddTemplateRow_TemplateRequired(t *testing.T) {
	svc := newTestService()
	r := &FlowsheetRow{Label: "Heart Rate"}
	err := svc.AddTemplateRow(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing template_id")
	}
}

func TestAddTemplateRow_LabelRequired(t *testing.T) {
	svc := newTestService()
	r := &FlowsheetRow{TemplateID: uuid.New()}
	err := svc.AddTemplateRow(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing label")
	}
}

// -- Flowsheet Entry Tests --

func TestCreateEntry(t *testing.T) {
	svc := newTestService()
	e := &FlowsheetEntry{
		TemplateID:   uuid.New(),
		RowID:        uuid.New(),
		PatientID:    uuid.New(),
		EncounterID:  uuid.New(),
		RecordedByID: uuid.New(),
	}
	err := svc.CreateEntry(context.Background(), e)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestCreateEntry_TemplateRequired(t *testing.T) {
	svc := newTestService()
	e := &FlowsheetEntry{RowID: uuid.New(), PatientID: uuid.New(), EncounterID: uuid.New(), RecordedByID: uuid.New()}
	err := svc.CreateEntry(context.Background(), e)
	if err == nil {
		t.Error("expected error for missing template_id")
	}
}

func TestCreateEntry_PatientRequired(t *testing.T) {
	svc := newTestService()
	e := &FlowsheetEntry{TemplateID: uuid.New(), RowID: uuid.New(), EncounterID: uuid.New(), RecordedByID: uuid.New()}
	err := svc.CreateEntry(context.Background(), e)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

// -- Nursing Assessment Tests --

func TestCreateAssessment(t *testing.T) {
	svc := newTestService()
	a := &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "admission"}
	err := svc.CreateAssessment(context.Background(), a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Status != "in-progress" {
		t.Errorf("expected default status 'in-progress', got %s", a.Status)
	}
}

func TestCreateAssessment_PatientRequired(t *testing.T) {
	svc := newTestService()
	a := &NursingAssessment{EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "admission"}
	err := svc.CreateAssessment(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateAssessment_TypeRequired(t *testing.T) {
	svc := newTestService()
	a := &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), NurseID: uuid.New()}
	err := svc.CreateAssessment(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing assessment_type")
	}
}

func TestDeleteAssessment(t *testing.T) {
	svc := newTestService()
	a := &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "admission"}
	svc.CreateAssessment(context.Background(), a)
	err := svc.DeleteAssessment(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetAssessment(context.Background(), a.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- Fall Risk Tests --

func TestCreateFallRisk(t *testing.T) {
	svc := newTestService()
	a := &FallRiskAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()}
	err := svc.CreateFallRisk(context.Background(), a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestCreateFallRisk_PatientRequired(t *testing.T) {
	svc := newTestService()
	a := &FallRiskAssessment{AssessedByID: uuid.New()}
	err := svc.CreateFallRisk(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateFallRisk_AssessedByRequired(t *testing.T) {
	svc := newTestService()
	a := &FallRiskAssessment{PatientID: uuid.New()}
	err := svc.CreateFallRisk(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing assessed_by_id")
	}
}

// -- Skin Assessment Tests --

func TestCreateSkinAssessment(t *testing.T) {
	svc := newTestService()
	a := &SkinAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()}
	err := svc.CreateSkinAssessment(context.Background(), a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateSkinAssessment_PatientRequired(t *testing.T) {
	svc := newTestService()
	a := &SkinAssessment{AssessedByID: uuid.New()}
	err := svc.CreateSkinAssessment(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

// -- Pain Assessment Tests --

func TestCreatePainAssessment(t *testing.T) {
	svc := newTestService()
	a := &PainAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()}
	err := svc.CreatePainAssessment(context.Background(), a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePainAssessment_PatientRequired(t *testing.T) {
	svc := newTestService()
	a := &PainAssessment{AssessedByID: uuid.New()}
	err := svc.CreatePainAssessment(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

// -- Lines/Drains Tests --

func TestCreateLinesDrains(t *testing.T) {
	svc := newTestService()
	l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "IV"}
	err := svc.CreateLinesDrains(context.Background(), l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Status != "active" {
		t.Errorf("expected default status 'active', got %s", l.Status)
	}
}

func TestCreateLinesDrains_PatientRequired(t *testing.T) {
	svc := newTestService()
	l := &LinesDrainsAirways{EncounterID: uuid.New(), Type: "IV"}
	err := svc.CreateLinesDrains(context.Background(), l)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateLinesDrains_TypeRequired(t *testing.T) {
	svc := newTestService()
	l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New()}
	err := svc.CreateLinesDrains(context.Background(), l)
	if err == nil {
		t.Error("expected error for missing type")
	}
}

func TestCreateLinesDrains_InvalidStatus(t *testing.T) {
	svc := newTestService()
	l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "IV", Status: "bogus"}
	err := svc.CreateLinesDrains(context.Background(), l)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestDeleteLinesDrains(t *testing.T) {
	svc := newTestService()
	l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "IV"}
	svc.CreateLinesDrains(context.Background(), l)
	err := svc.DeleteLinesDrains(context.Background(), l.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetLinesDrains(context.Background(), l.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// -- Restraint Tests --

func TestCreateRestraint(t *testing.T) {
	svc := newTestService()
	r := &RestraintRecord{PatientID: uuid.New(), RestraintType: "wrist", AppliedByID: uuid.New()}
	err := svc.CreateRestraint(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestCreateRestraint_PatientRequired(t *testing.T) {
	svc := newTestService()
	r := &RestraintRecord{RestraintType: "wrist", AppliedByID: uuid.New()}
	err := svc.CreateRestraint(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateRestraint_TypeRequired(t *testing.T) {
	svc := newTestService()
	r := &RestraintRecord{PatientID: uuid.New(), AppliedByID: uuid.New()}
	err := svc.CreateRestraint(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing restraint_type")
	}
}

func TestCreateRestraint_AppliedByRequired(t *testing.T) {
	svc := newTestService()
	r := &RestraintRecord{PatientID: uuid.New(), RestraintType: "wrist"}
	err := svc.CreateRestraint(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing applied_by_id")
	}
}

// -- Intake/Output Tests --

func TestCreateIntakeOutput(t *testing.T) {
	svc := newTestService()
	r := &IntakeOutputRecord{PatientID: uuid.New(), EncounterID: uuid.New(), Category: "intake", RecordedByID: uuid.New()}
	err := svc.CreateIntakeOutput(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestCreateIntakeOutput_PatientRequired(t *testing.T) {
	svc := newTestService()
	r := &IntakeOutputRecord{EncounterID: uuid.New(), Category: "intake", RecordedByID: uuid.New()}
	err := svc.CreateIntakeOutput(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateIntakeOutput_CategoryRequired(t *testing.T) {
	svc := newTestService()
	r := &IntakeOutputRecord{PatientID: uuid.New(), EncounterID: uuid.New(), RecordedByID: uuid.New()}
	err := svc.CreateIntakeOutput(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing category")
	}
}

func TestDeleteIntakeOutput(t *testing.T) {
	svc := newTestService()
	r := &IntakeOutputRecord{PatientID: uuid.New(), EncounterID: uuid.New(), Category: "intake", RecordedByID: uuid.New()}
	svc.CreateIntakeOutput(context.Background(), r)
	err := svc.DeleteIntakeOutput(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetIntakeOutput(context.Background(), r.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

// =========== Additional Template Tests ===========

func TestGetTemplate_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetTemplate(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateTemplate(t *testing.T) {
	svc := newTestService()
	tmpl := &FlowsheetTemplate{Name: "Vitals"}
	svc.CreateTemplate(context.Background(), tmpl)
	tmpl.Name = "Vitals v2"
	err := svc.UpdateTemplate(context.Background(), tmpl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTemplates(t *testing.T) {
	svc := newTestService()
	svc.CreateTemplate(context.Background(), &FlowsheetTemplate{Name: "Vitals"})
	svc.CreateTemplate(context.Background(), &FlowsheetTemplate{Name: "Neuro"})
	items, total, err := svc.ListTemplates(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 templates, got %d", total)
	}
}

func TestGetTemplateRows(t *testing.T) {
	svc := newTestService()
	tmpl := &FlowsheetTemplate{Name: "Vitals"}
	svc.CreateTemplate(context.Background(), tmpl)
	svc.AddTemplateRow(context.Background(), &FlowsheetRow{TemplateID: tmpl.ID, Label: "HR"})
	svc.AddTemplateRow(context.Background(), &FlowsheetRow{TemplateID: tmpl.ID, Label: "BP"})

	rows, err := svc.GetTemplateRows(context.Background(), tmpl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(rows))
	}
}

// =========== Additional Entry Tests ===========

func TestCreateEntry_RowRequired(t *testing.T) {
	svc := newTestService()
	e := &FlowsheetEntry{TemplateID: uuid.New(), PatientID: uuid.New(), EncounterID: uuid.New(), RecordedByID: uuid.New()}
	err := svc.CreateEntry(context.Background(), e)
	if err == nil {
		t.Error("expected error for missing row_id")
	}
}

func TestCreateEntry_EncounterRequired(t *testing.T) {
	svc := newTestService()
	e := &FlowsheetEntry{TemplateID: uuid.New(), RowID: uuid.New(), PatientID: uuid.New(), RecordedByID: uuid.New()}
	err := svc.CreateEntry(context.Background(), e)
	if err == nil {
		t.Error("expected error for missing encounter_id")
	}
}

func TestCreateEntry_RecordedByRequired(t *testing.T) {
	svc := newTestService()
	e := &FlowsheetEntry{TemplateID: uuid.New(), RowID: uuid.New(), PatientID: uuid.New(), EncounterID: uuid.New()}
	err := svc.CreateEntry(context.Background(), e)
	if err == nil {
		t.Error("expected error for missing recorded_by_id")
	}
}

func TestGetEntry(t *testing.T) {
	svc := newTestService()
	e := &FlowsheetEntry{TemplateID: uuid.New(), RowID: uuid.New(), PatientID: uuid.New(), EncounterID: uuid.New(), RecordedByID: uuid.New()}
	svc.CreateEntry(context.Background(), e)
	got, err := svc.GetEntry(context.Background(), e.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != e.ID {
		t.Errorf("expected ID %v, got %v", e.ID, got.ID)
	}
}

func TestGetEntry_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetEntry(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestDeleteEntry(t *testing.T) {
	svc := newTestService()
	e := &FlowsheetEntry{TemplateID: uuid.New(), RowID: uuid.New(), PatientID: uuid.New(), EncounterID: uuid.New(), RecordedByID: uuid.New()}
	svc.CreateEntry(context.Background(), e)
	err := svc.DeleteEntry(context.Background(), e.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetEntry(context.Background(), e.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListEntriesByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateEntry(context.Background(), &FlowsheetEntry{TemplateID: uuid.New(), RowID: uuid.New(), PatientID: pid, EncounterID: uuid.New(), RecordedByID: uuid.New()})
	svc.CreateEntry(context.Background(), &FlowsheetEntry{TemplateID: uuid.New(), RowID: uuid.New(), PatientID: uuid.New(), EncounterID: uuid.New(), RecordedByID: uuid.New()})
	items, total, err := svc.ListEntriesByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

func TestListEntriesByEncounter(t *testing.T) {
	svc := newTestService()
	eid := uuid.New()
	svc.CreateEntry(context.Background(), &FlowsheetEntry{TemplateID: uuid.New(), RowID: uuid.New(), PatientID: uuid.New(), EncounterID: eid, RecordedByID: uuid.New()})
	svc.CreateEntry(context.Background(), &FlowsheetEntry{TemplateID: uuid.New(), RowID: uuid.New(), PatientID: uuid.New(), EncounterID: uuid.New(), RecordedByID: uuid.New()})
	items, total, err := svc.ListEntriesByEncounter(context.Background(), eid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

func TestSearchEntries(t *testing.T) {
	svc := newTestService()
	svc.CreateEntry(context.Background(), &FlowsheetEntry{TemplateID: uuid.New(), RowID: uuid.New(), PatientID: uuid.New(), EncounterID: uuid.New(), RecordedByID: uuid.New()})
	items, total, err := svc.SearchEntries(context.Background(), map[string]string{}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 || len(items) < 1 {
		t.Error("expected items")
	}
}

// =========== Additional Assessment Tests ===========

func TestCreateAssessment_EncounterRequired(t *testing.T) {
	svc := newTestService()
	a := &NursingAssessment{PatientID: uuid.New(), NurseID: uuid.New(), AssessmentType: "admission"}
	err := svc.CreateAssessment(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing encounter_id")
	}
}

func TestCreateAssessment_NurseRequired(t *testing.T) {
	svc := newTestService()
	a := &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), AssessmentType: "admission"}
	err := svc.CreateAssessment(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing nurse_id")
	}
}

func TestGetAssessment(t *testing.T) {
	svc := newTestService()
	a := &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "admission"}
	svc.CreateAssessment(context.Background(), a)
	got, err := svc.GetAssessment(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != a.ID {
		t.Errorf("expected ID %v, got %v", a.ID, got.ID)
	}
}

func TestGetAssessment_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetAssessment(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateAssessment(t *testing.T) {
	svc := newTestService()
	a := &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "admission"}
	svc.CreateAssessment(context.Background(), a)
	a.Status = "completed"
	err := svc.UpdateAssessment(context.Background(), a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListAssessmentsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateAssessment(context.Background(), &NursingAssessment{PatientID: pid, EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "admission"})
	svc.CreateAssessment(context.Background(), &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "shift"})
	items, total, err := svc.ListAssessmentsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

func TestListAssessmentsByEncounter(t *testing.T) {
	svc := newTestService()
	eid := uuid.New()
	svc.CreateAssessment(context.Background(), &NursingAssessment{PatientID: uuid.New(), EncounterID: eid, NurseID: uuid.New(), AssessmentType: "admission"})
	svc.CreateAssessment(context.Background(), &NursingAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), NurseID: uuid.New(), AssessmentType: "shift"})
	items, total, err := svc.ListAssessmentsByEncounter(context.Background(), eid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

// =========== Additional Fall Risk Tests ===========

func TestGetFallRisk(t *testing.T) {
	svc := newTestService()
	a := &FallRiskAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()}
	svc.CreateFallRisk(context.Background(), a)
	got, err := svc.GetFallRisk(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != a.ID {
		t.Errorf("expected ID %v, got %v", a.ID, got.ID)
	}
}

func TestGetFallRisk_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetFallRisk(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestListFallRiskByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateFallRisk(context.Background(), &FallRiskAssessment{PatientID: pid, AssessedByID: uuid.New()})
	svc.CreateFallRisk(context.Background(), &FallRiskAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()})
	items, total, err := svc.ListFallRiskByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

// =========== Additional Skin Assessment Tests ===========

func TestCreateSkinAssessment_AssessedByRequired(t *testing.T) {
	svc := newTestService()
	a := &SkinAssessment{PatientID: uuid.New()}
	err := svc.CreateSkinAssessment(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing assessed_by_id")
	}
}

func TestGetSkinAssessment(t *testing.T) {
	svc := newTestService()
	a := &SkinAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()}
	svc.CreateSkinAssessment(context.Background(), a)
	got, err := svc.GetSkinAssessment(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != a.ID {
		t.Errorf("expected ID %v, got %v", a.ID, got.ID)
	}
}

func TestGetSkinAssessment_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetSkinAssessment(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestListSkinAssessmentsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateSkinAssessment(context.Background(), &SkinAssessment{PatientID: pid, AssessedByID: uuid.New()})
	svc.CreateSkinAssessment(context.Background(), &SkinAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()})
	items, total, err := svc.ListSkinAssessmentsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

// =========== Additional Pain Assessment Tests ===========

func TestCreatePainAssessment_AssessedByRequired(t *testing.T) {
	svc := newTestService()
	a := &PainAssessment{PatientID: uuid.New()}
	err := svc.CreatePainAssessment(context.Background(), a)
	if err == nil {
		t.Error("expected error for missing assessed_by_id")
	}
}

func TestGetPainAssessment(t *testing.T) {
	svc := newTestService()
	a := &PainAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()}
	svc.CreatePainAssessment(context.Background(), a)
	got, err := svc.GetPainAssessment(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != a.ID {
		t.Errorf("expected ID %v, got %v", a.ID, got.ID)
	}
}

func TestGetPainAssessment_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetPainAssessment(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestListPainAssessmentsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreatePainAssessment(context.Background(), &PainAssessment{PatientID: pid, AssessedByID: uuid.New()})
	svc.CreatePainAssessment(context.Background(), &PainAssessment{PatientID: uuid.New(), AssessedByID: uuid.New()})
	items, total, err := svc.ListPainAssessmentsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

// =========== Additional Lines/Drains Tests ===========

func TestCreateLinesDrains_EncounterRequired(t *testing.T) {
	svc := newTestService()
	l := &LinesDrainsAirways{PatientID: uuid.New(), Type: "IV"}
	err := svc.CreateLinesDrains(context.Background(), l)
	if err == nil {
		t.Error("expected error for missing encounter_id")
	}
}

func TestCreateLinesDrains_ValidStatuses(t *testing.T) {
	for _, s := range []string{"active", "removed", "replaced", "capped"} {
		svc := newTestService()
		l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "IV", Status: s}
		if err := svc.CreateLinesDrains(context.Background(), l); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

func TestGetLinesDrains(t *testing.T) {
	svc := newTestService()
	l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "IV"}
	svc.CreateLinesDrains(context.Background(), l)
	got, err := svc.GetLinesDrains(context.Background(), l.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != l.ID {
		t.Errorf("expected ID %v, got %v", l.ID, got.ID)
	}
}

func TestGetLinesDrains_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetLinesDrains(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateLinesDrains(t *testing.T) {
	svc := newTestService()
	l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "IV"}
	svc.CreateLinesDrains(context.Background(), l)
	l.Status = "removed"
	err := svc.UpdateLinesDrains(context.Background(), l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateLinesDrains_InvalidStatus(t *testing.T) {
	svc := newTestService()
	l := &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "IV"}
	svc.CreateLinesDrains(context.Background(), l)
	l.Status = "bogus"
	err := svc.UpdateLinesDrains(context.Background(), l)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestListLinesDrainsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateLinesDrains(context.Background(), &LinesDrainsAirways{PatientID: pid, EncounterID: uuid.New(), Type: "IV"})
	svc.CreateLinesDrains(context.Background(), &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "Drain"})
	items, total, err := svc.ListLinesDrainsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

func TestListLinesDrainsByEncounter(t *testing.T) {
	svc := newTestService()
	eid := uuid.New()
	svc.CreateLinesDrains(context.Background(), &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: eid, Type: "IV"})
	svc.CreateLinesDrains(context.Background(), &LinesDrainsAirways{PatientID: uuid.New(), EncounterID: uuid.New(), Type: "Drain"})
	items, total, err := svc.ListLinesDrainsByEncounter(context.Background(), eid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

// =========== Additional Restraint Tests ===========

func TestGetRestraint(t *testing.T) {
	svc := newTestService()
	r := &RestraintRecord{PatientID: uuid.New(), RestraintType: "wrist", AppliedByID: uuid.New()}
	svc.CreateRestraint(context.Background(), r)
	got, err := svc.GetRestraint(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != r.ID {
		t.Errorf("expected ID %v, got %v", r.ID, got.ID)
	}
}

func TestGetRestraint_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetRestraint(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestUpdateRestraint(t *testing.T) {
	svc := newTestService()
	r := &RestraintRecord{PatientID: uuid.New(), RestraintType: "wrist", AppliedByID: uuid.New()}
	svc.CreateRestraint(context.Background(), r)
	r.RestraintType = "vest"
	err := svc.UpdateRestraint(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListRestraintsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateRestraint(context.Background(), &RestraintRecord{PatientID: pid, RestraintType: "wrist", AppliedByID: uuid.New()})
	svc.CreateRestraint(context.Background(), &RestraintRecord{PatientID: uuid.New(), RestraintType: "vest", AppliedByID: uuid.New()})
	items, total, err := svc.ListRestraintsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

// =========== Additional Intake/Output Tests ===========

func TestCreateIntakeOutput_EncounterRequired(t *testing.T) {
	svc := newTestService()
	r := &IntakeOutputRecord{PatientID: uuid.New(), Category: "intake", RecordedByID: uuid.New()}
	err := svc.CreateIntakeOutput(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing encounter_id")
	}
}

func TestCreateIntakeOutput_RecordedByRequired(t *testing.T) {
	svc := newTestService()
	r := &IntakeOutputRecord{PatientID: uuid.New(), EncounterID: uuid.New(), Category: "intake"}
	err := svc.CreateIntakeOutput(context.Background(), r)
	if err == nil {
		t.Error("expected error for missing recorded_by_id")
	}
}

func TestGetIntakeOutput(t *testing.T) {
	svc := newTestService()
	r := &IntakeOutputRecord{PatientID: uuid.New(), EncounterID: uuid.New(), Category: "intake", RecordedByID: uuid.New()}
	svc.CreateIntakeOutput(context.Background(), r)
	got, err := svc.GetIntakeOutput(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != r.ID {
		t.Errorf("expected ID %v, got %v", r.ID, got.ID)
	}
}

func TestGetIntakeOutput_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetIntakeOutput(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error for not found")
	}
}

func TestListIntakeOutputByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateIntakeOutput(context.Background(), &IntakeOutputRecord{PatientID: pid, EncounterID: uuid.New(), Category: "intake", RecordedByID: uuid.New()})
	svc.CreateIntakeOutput(context.Background(), &IntakeOutputRecord{PatientID: uuid.New(), EncounterID: uuid.New(), Category: "output", RecordedByID: uuid.New()})
	items, total, err := svc.ListIntakeOutputByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}

func TestListIntakeOutputByEncounter(t *testing.T) {
	svc := newTestService()
	eid := uuid.New()
	svc.CreateIntakeOutput(context.Background(), &IntakeOutputRecord{PatientID: uuid.New(), EncounterID: eid, Category: "intake", RecordedByID: uuid.New()})
	svc.CreateIntakeOutput(context.Background(), &IntakeOutputRecord{PatientID: uuid.New(), EncounterID: uuid.New(), Category: "output", RecordedByID: uuid.New()})
	items, total, err := svc.ListIntakeOutputByEncounter(context.Background(), eid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1, got %d", total)
	}
}
