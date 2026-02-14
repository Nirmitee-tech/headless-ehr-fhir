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
