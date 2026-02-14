package cds

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// ── Mock Repositories ──

type mockRuleRepo struct {
	data map[uuid.UUID]*CDSRule
}

func (m *mockRuleRepo) Create(_ context.Context, r *CDSRule) error {
	r.ID = uuid.New()
	m.data[r.ID] = r
	return nil
}
func (m *mockRuleRepo) GetByID(_ context.Context, id uuid.UUID) (*CDSRule, error) {
	if r, ok := m.data[id]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockRuleRepo) Update(_ context.Context, r *CDSRule) error {
	if _, ok := m.data[r.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[r.ID] = r
	return nil
}
func (m *mockRuleRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockRuleRepo) List(_ context.Context, limit, offset int) ([]*CDSRule, int, error) {
	var out []*CDSRule
	for _, r := range m.data {
		out = append(out, r)
	}
	return out, len(out), nil
}

type mockAlertRepo struct {
	data      map[uuid.UUID]*CDSAlert
	responses map[uuid.UUID][]*CDSAlertResponse
}

func (m *mockAlertRepo) Create(_ context.Context, a *CDSAlert) error {
	a.ID = uuid.New()
	m.data[a.ID] = a
	return nil
}
func (m *mockAlertRepo) GetByID(_ context.Context, id uuid.UUID) (*CDSAlert, error) {
	if a, ok := m.data[id]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockAlertRepo) Update(_ context.Context, a *CDSAlert) error {
	if _, ok := m.data[a.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[a.ID] = a
	return nil
}
func (m *mockAlertRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockAlertRepo) List(_ context.Context, limit, offset int) ([]*CDSAlert, int, error) {
	var out []*CDSAlert
	for _, a := range m.data {
		out = append(out, a)
	}
	return out, len(out), nil
}
func (m *mockAlertRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*CDSAlert, int, error) {
	var out []*CDSAlert
	for _, a := range m.data {
		if a.PatientID == patientID {
			out = append(out, a)
		}
	}
	return out, len(out), nil
}
func (m *mockAlertRepo) AddResponse(_ context.Context, resp *CDSAlertResponse) error {
	resp.ID = uuid.New()
	m.responses[resp.AlertID] = append(m.responses[resp.AlertID], resp)
	return nil
}
func (m *mockAlertRepo) GetResponses(_ context.Context, alertID uuid.UUID) ([]*CDSAlertResponse, error) {
	return m.responses[alertID], nil
}

type mockInteractionRepo struct {
	data map[uuid.UUID]*DrugInteraction
}

func (m *mockInteractionRepo) Create(_ context.Context, d *DrugInteraction) error {
	d.ID = uuid.New()
	m.data[d.ID] = d
	return nil
}
func (m *mockInteractionRepo) GetByID(_ context.Context, id uuid.UUID) (*DrugInteraction, error) {
	if d, ok := m.data[id]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockInteractionRepo) Update(_ context.Context, d *DrugInteraction) error {
	if _, ok := m.data[d.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[d.ID] = d
	return nil
}
func (m *mockInteractionRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockInteractionRepo) List(_ context.Context, limit, offset int) ([]*DrugInteraction, int, error) {
	var out []*DrugInteraction
	for _, d := range m.data {
		out = append(out, d)
	}
	return out, len(out), nil
}

type mockOrderSetRepo struct {
	data     map[uuid.UUID]*OrderSet
	sections map[uuid.UUID][]*OrderSetSection
	items    map[uuid.UUID][]*OrderSetItem
}

func (m *mockOrderSetRepo) Create(_ context.Context, o *OrderSet) error {
	o.ID = uuid.New()
	m.data[o.ID] = o
	return nil
}
func (m *mockOrderSetRepo) GetByID(_ context.Context, id uuid.UUID) (*OrderSet, error) {
	if o, ok := m.data[id]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockOrderSetRepo) Update(_ context.Context, o *OrderSet) error {
	if _, ok := m.data[o.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[o.ID] = o
	return nil
}
func (m *mockOrderSetRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockOrderSetRepo) List(_ context.Context, limit, offset int) ([]*OrderSet, int, error) {
	var out []*OrderSet
	for _, o := range m.data {
		out = append(out, o)
	}
	return out, len(out), nil
}
func (m *mockOrderSetRepo) AddSection(_ context.Context, s *OrderSetSection) error {
	s.ID = uuid.New()
	m.sections[s.OrderSetID] = append(m.sections[s.OrderSetID], s)
	return nil
}
func (m *mockOrderSetRepo) GetSections(_ context.Context, orderSetID uuid.UUID) ([]*OrderSetSection, error) {
	return m.sections[orderSetID], nil
}
func (m *mockOrderSetRepo) AddItem(_ context.Context, item *OrderSetItem) error {
	item.ID = uuid.New()
	m.items[item.SectionID] = append(m.items[item.SectionID], item)
	return nil
}
func (m *mockOrderSetRepo) GetItems(_ context.Context, sectionID uuid.UUID) ([]*OrderSetItem, error) {
	return m.items[sectionID], nil
}

type mockPathwayRepo struct {
	data   map[uuid.UUID]*ClinicalPathway
	phases map[uuid.UUID][]*ClinicalPathwayPhase
}

func (m *mockPathwayRepo) Create(_ context.Context, p *ClinicalPathway) error {
	p.ID = uuid.New()
	m.data[p.ID] = p
	return nil
}
func (m *mockPathwayRepo) GetByID(_ context.Context, id uuid.UUID) (*ClinicalPathway, error) {
	if p, ok := m.data[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockPathwayRepo) Update(_ context.Context, p *ClinicalPathway) error {
	if _, ok := m.data[p.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[p.ID] = p
	return nil
}
func (m *mockPathwayRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockPathwayRepo) List(_ context.Context, limit, offset int) ([]*ClinicalPathway, int, error) {
	var out []*ClinicalPathway
	for _, p := range m.data {
		out = append(out, p)
	}
	return out, len(out), nil
}
func (m *mockPathwayRepo) AddPhase(_ context.Context, phase *ClinicalPathwayPhase) error {
	phase.ID = uuid.New()
	m.phases[phase.PathwayID] = append(m.phases[phase.PathwayID], phase)
	return nil
}
func (m *mockPathwayRepo) GetPhases(_ context.Context, pathwayID uuid.UUID) ([]*ClinicalPathwayPhase, error) {
	return m.phases[pathwayID], nil
}

type mockEnrollmentRepo struct {
	data map[uuid.UUID]*PatientPathwayEnrollment
}

func (m *mockEnrollmentRepo) Create(_ context.Context, e *PatientPathwayEnrollment) error {
	e.ID = uuid.New()
	m.data[e.ID] = e
	return nil
}
func (m *mockEnrollmentRepo) GetByID(_ context.Context, id uuid.UUID) (*PatientPathwayEnrollment, error) {
	if e, ok := m.data[id]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockEnrollmentRepo) Update(_ context.Context, e *PatientPathwayEnrollment) error {
	if _, ok := m.data[e.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[e.ID] = e
	return nil
}
func (m *mockEnrollmentRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockEnrollmentRepo) List(_ context.Context, limit, offset int) ([]*PatientPathwayEnrollment, int, error) {
	var out []*PatientPathwayEnrollment
	for _, e := range m.data {
		out = append(out, e)
	}
	return out, len(out), nil
}
func (m *mockEnrollmentRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*PatientPathwayEnrollment, int, error) {
	var out []*PatientPathwayEnrollment
	for _, e := range m.data {
		if e.PatientID == patientID {
			out = append(out, e)
		}
	}
	return out, len(out), nil
}

type mockFormularyRepo struct {
	data  map[uuid.UUID]*Formulary
	items map[uuid.UUID][]*FormularyItem
}

func (m *mockFormularyRepo) Create(_ context.Context, f *Formulary) error {
	f.ID = uuid.New()
	m.data[f.ID] = f
	return nil
}
func (m *mockFormularyRepo) GetByID(_ context.Context, id uuid.UUID) (*Formulary, error) {
	if f, ok := m.data[id]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockFormularyRepo) Update(_ context.Context, f *Formulary) error {
	if _, ok := m.data[f.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[f.ID] = f
	return nil
}
func (m *mockFormularyRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockFormularyRepo) List(_ context.Context, limit, offset int) ([]*Formulary, int, error) {
	var out []*Formulary
	for _, f := range m.data {
		out = append(out, f)
	}
	return out, len(out), nil
}
func (m *mockFormularyRepo) AddItem(_ context.Context, item *FormularyItem) error {
	item.ID = uuid.New()
	m.items[item.FormularyID] = append(m.items[item.FormularyID], item)
	return nil
}
func (m *mockFormularyRepo) GetItems(_ context.Context, formularyID uuid.UUID) ([]*FormularyItem, error) {
	return m.items[formularyID], nil
}

type mockMedReconcRepo struct {
	data  map[uuid.UUID]*MedicationReconciliation
	items map[uuid.UUID][]*MedicationReconciliationItem
}

func (m *mockMedReconcRepo) Create(_ context.Context, mr *MedicationReconciliation) error {
	mr.ID = uuid.New()
	m.data[mr.ID] = mr
	return nil
}
func (m *mockMedReconcRepo) GetByID(_ context.Context, id uuid.UUID) (*MedicationReconciliation, error) {
	if mr, ok := m.data[id]; ok {
		return mr, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockMedReconcRepo) Update(_ context.Context, mr *MedicationReconciliation) error {
	if _, ok := m.data[mr.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[mr.ID] = mr
	return nil
}
func (m *mockMedReconcRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockMedReconcRepo) List(_ context.Context, limit, offset int) ([]*MedicationReconciliation, int, error) {
	var out []*MedicationReconciliation
	for _, mr := range m.data {
		out = append(out, mr)
	}
	return out, len(out), nil
}
func (m *mockMedReconcRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*MedicationReconciliation, int, error) {
	var out []*MedicationReconciliation
	for _, mr := range m.data {
		if mr.PatientID == patientID {
			out = append(out, mr)
		}
	}
	return out, len(out), nil
}
func (m *mockMedReconcRepo) AddItem(_ context.Context, item *MedicationReconciliationItem) error {
	item.ID = uuid.New()
	m.items[item.ReconciliationID] = append(m.items[item.ReconciliationID], item)
	return nil
}
func (m *mockMedReconcRepo) GetItems(_ context.Context, reconciliationID uuid.UUID) ([]*MedicationReconciliationItem, error) {
	return m.items[reconciliationID], nil
}

// ── Helper ──

func newTestService() *Service {
	return NewService(
		&mockRuleRepo{data: make(map[uuid.UUID]*CDSRule)},
		&mockAlertRepo{data: make(map[uuid.UUID]*CDSAlert), responses: make(map[uuid.UUID][]*CDSAlertResponse)},
		&mockInteractionRepo{data: make(map[uuid.UUID]*DrugInteraction)},
		&mockOrderSetRepo{data: make(map[uuid.UUID]*OrderSet), sections: make(map[uuid.UUID][]*OrderSetSection), items: make(map[uuid.UUID][]*OrderSetItem)},
		&mockPathwayRepo{data: make(map[uuid.UUID]*ClinicalPathway), phases: make(map[uuid.UUID][]*ClinicalPathwayPhase)},
		&mockEnrollmentRepo{data: make(map[uuid.UUID]*PatientPathwayEnrollment)},
		&mockFormularyRepo{data: make(map[uuid.UUID]*Formulary), items: make(map[uuid.UUID][]*FormularyItem)},
		&mockMedReconcRepo{data: make(map[uuid.UUID]*MedicationReconciliation), items: make(map[uuid.UUID][]*MedicationReconciliationItem)},
	)
}

// ── CDS Rule Tests ──

func TestService_CreateCDSRule(t *testing.T) {
	svc := newTestService()
	r := &CDSRule{RuleName: "Drug Allergy", RuleType: "allergy-check"}
	if err := svc.CreateCDSRule(nil, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestService_CreateCDSRule_MissingRuleName(t *testing.T) {
	svc := newTestService()
	r := &CDSRule{RuleType: "allergy-check"}
	if err := svc.CreateCDSRule(nil, r); err == nil {
		t.Error("expected error for missing rule_name")
	}
}

func TestService_CreateCDSRule_MissingRuleType(t *testing.T) {
	svc := newTestService()
	r := &CDSRule{RuleName: "Drug Allergy"}
	if err := svc.CreateCDSRule(nil, r); err == nil {
		t.Error("expected error for missing rule_type")
	}
}

func TestService_GetCDSRule(t *testing.T) {
	svc := newTestService()
	r := &CDSRule{RuleName: "Drug Allergy", RuleType: "allergy-check"}
	svc.CreateCDSRule(nil, r)
	got, err := svc.GetCDSRule(nil, r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.RuleName != "Drug Allergy" {
		t.Errorf("expected 'Drug Allergy', got %s", got.RuleName)
	}
}

func TestService_GetCDSRule_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetCDSRule(nil, uuid.New()); err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_DeleteCDSRule(t *testing.T) {
	svc := newTestService()
	r := &CDSRule{RuleName: "Drug Allergy", RuleType: "allergy-check"}
	svc.CreateCDSRule(nil, r)
	if err := svc.DeleteCDSRule(nil, r.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListCDSRules(t *testing.T) {
	svc := newTestService()
	svc.CreateCDSRule(nil, &CDSRule{RuleName: "R1", RuleType: "t1"})
	svc.CreateCDSRule(nil, &CDSRule{RuleName: "R2", RuleType: "t2"})
	items, total, err := svc.ListCDSRules(nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}

// ── CDS Alert Tests ──

func TestService_CreateCDSAlert(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Drug allergy detected"}
	if err := svc.CreateCDSAlert(nil, a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Status != "fired" {
		t.Errorf("expected default status 'fired', got %s", a.Status)
	}
}

func TestService_CreateCDSAlert_MissingRuleID(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{PatientID: uuid.New(), Summary: "Alert"}
	if err := svc.CreateCDSAlert(nil, a); err == nil {
		t.Error("expected error for missing rule_id")
	}
}

func TestService_CreateCDSAlert_MissingPatientID(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{RuleID: uuid.New(), Summary: "Alert"}
	if err := svc.CreateCDSAlert(nil, a); err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestService_CreateCDSAlert_MissingSummary(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New()}
	if err := svc.CreateCDSAlert(nil, a); err == nil {
		t.Error("expected error for missing summary")
	}
}

func TestService_CreateCDSAlert_InvalidStatus(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Alert", Status: "bogus"}
	if err := svc.CreateCDSAlert(nil, a); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_CreateCDSAlert_ValidStatuses(t *testing.T) {
	statuses := []string{"fired", "accepted", "overridden", "auto-resolved", "expired", "suppressed"}
	for _, status := range statuses {
		svc := newTestService()
		a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Alert", Status: status}
		if err := svc.CreateCDSAlert(nil, a); err != nil {
			t.Errorf("status %q should be valid, got error: %v", status, err)
		}
	}
}

func TestService_GetCDSAlert(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Alert"}
	svc.CreateCDSAlert(nil, a)
	got, err := svc.GetCDSAlert(nil, a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Summary != "Alert" {
		t.Errorf("expected 'Alert', got %s", got.Summary)
	}
}

func TestService_UpdateCDSAlert(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Alert"}
	svc.CreateCDSAlert(nil, a)
	a.Status = "accepted"
	if err := svc.UpdateCDSAlert(nil, a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdateCDSAlert_InvalidStatus(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Alert"}
	svc.CreateCDSAlert(nil, a)
	a.Status = "bad"
	if err := svc.UpdateCDSAlert(nil, a); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_DeleteCDSAlert(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Alert"}
	svc.CreateCDSAlert(nil, a)
	if err := svc.DeleteCDSAlert(nil, a.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListCDSAlertsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateCDSAlert(nil, &CDSAlert{RuleID: uuid.New(), PatientID: pid, Summary: "A1"})
	svc.CreateCDSAlert(nil, &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "A2"})
	items, total, err := svc.ListCDSAlertsByPatient(nil, pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

func TestService_AddAlertResponse(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Alert"}
	svc.CreateCDSAlert(nil, a)
	resp := &CDSAlertResponse{AlertID: a.ID, PractitionerID: uuid.New(), Action: "accept"}
	if err := svc.AddAlertResponse(nil, resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestService_AddAlertResponse_MissingAlertID(t *testing.T) {
	svc := newTestService()
	resp := &CDSAlertResponse{PractitionerID: uuid.New(), Action: "accept"}
	if err := svc.AddAlertResponse(nil, resp); err == nil {
		t.Error("expected error for missing alert_id")
	}
}

func TestService_AddAlertResponse_MissingPractitionerID(t *testing.T) {
	svc := newTestService()
	resp := &CDSAlertResponse{AlertID: uuid.New(), Action: "accept"}
	if err := svc.AddAlertResponse(nil, resp); err == nil {
		t.Error("expected error for missing practitioner_id")
	}
}

func TestService_AddAlertResponse_MissingAction(t *testing.T) {
	svc := newTestService()
	resp := &CDSAlertResponse{AlertID: uuid.New(), PractitionerID: uuid.New()}
	if err := svc.AddAlertResponse(nil, resp); err == nil {
		t.Error("expected error for missing action")
	}
}

func TestService_GetAlertResponses(t *testing.T) {
	svc := newTestService()
	a := &CDSAlert{RuleID: uuid.New(), PatientID: uuid.New(), Summary: "Alert"}
	svc.CreateCDSAlert(nil, a)
	svc.AddAlertResponse(nil, &CDSAlertResponse{AlertID: a.ID, PractitionerID: uuid.New(), Action: "accept"})
	svc.AddAlertResponse(nil, &CDSAlertResponse{AlertID: a.ID, PractitionerID: uuid.New(), Action: "override"})
	responses, err := svc.GetAlertResponses(nil, a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(responses) != 2 {
		t.Errorf("expected 2, got %d", len(responses))
	}
}

// ── Drug Interaction Tests ──

func TestService_CreateDrugInteraction(t *testing.T) {
	svc := newTestService()
	d := &DrugInteraction{MedicationAName: "Warfarin", MedicationBName: "Aspirin", Severity: "high"}
	if err := svc.CreateDrugInteraction(nil, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestService_CreateDrugInteraction_MissingMedA(t *testing.T) {
	svc := newTestService()
	d := &DrugInteraction{MedicationBName: "Aspirin", Severity: "high"}
	if err := svc.CreateDrugInteraction(nil, d); err == nil {
		t.Error("expected error for missing medication_a_name")
	}
}

func TestService_CreateDrugInteraction_MissingMedB(t *testing.T) {
	svc := newTestService()
	d := &DrugInteraction{MedicationAName: "Warfarin", Severity: "high"}
	if err := svc.CreateDrugInteraction(nil, d); err == nil {
		t.Error("expected error for missing medication_b_name")
	}
}

func TestService_CreateDrugInteraction_MissingSeverity(t *testing.T) {
	svc := newTestService()
	d := &DrugInteraction{MedicationAName: "Warfarin", MedicationBName: "Aspirin"}
	if err := svc.CreateDrugInteraction(nil, d); err == nil {
		t.Error("expected error for missing severity")
	}
}

func TestService_GetDrugInteraction(t *testing.T) {
	svc := newTestService()
	d := &DrugInteraction{MedicationAName: "Warfarin", MedicationBName: "Aspirin", Severity: "high"}
	svc.CreateDrugInteraction(nil, d)
	got, err := svc.GetDrugInteraction(nil, d.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MedicationAName != "Warfarin" {
		t.Errorf("expected 'Warfarin', got %s", got.MedicationAName)
	}
}

func TestService_DeleteDrugInteraction(t *testing.T) {
	svc := newTestService()
	d := &DrugInteraction{MedicationAName: "Warfarin", MedicationBName: "Aspirin", Severity: "high"}
	svc.CreateDrugInteraction(nil, d)
	if err := svc.DeleteDrugInteraction(nil, d.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── Order Set Tests ──

func TestService_CreateOrderSet(t *testing.T) {
	svc := newTestService()
	o := &OrderSet{Name: "Sepsis Bundle"}
	if err := svc.CreateOrderSet(nil, o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Status != "draft" {
		t.Errorf("expected default status 'draft', got %s", o.Status)
	}
}

func TestService_CreateOrderSet_MissingName(t *testing.T) {
	svc := newTestService()
	o := &OrderSet{}
	if err := svc.CreateOrderSet(nil, o); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestService_GetOrderSet(t *testing.T) {
	svc := newTestService()
	o := &OrderSet{Name: "Sepsis Bundle"}
	svc.CreateOrderSet(nil, o)
	got, err := svc.GetOrderSet(nil, o.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Sepsis Bundle" {
		t.Errorf("expected 'Sepsis Bundle', got %s", got.Name)
	}
}

func TestService_DeleteOrderSet(t *testing.T) {
	svc := newTestService()
	o := &OrderSet{Name: "Sepsis Bundle"}
	svc.CreateOrderSet(nil, o)
	if err := svc.DeleteOrderSet(nil, o.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_AddOrderSetSection(t *testing.T) {
	svc := newTestService()
	o := &OrderSet{Name: "Sepsis Bundle"}
	svc.CreateOrderSet(nil, o)
	sec := &OrderSetSection{OrderSetID: o.ID, Name: "Antibiotics"}
	if err := svc.AddOrderSetSection(nil, sec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sec.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestService_AddOrderSetSection_MissingOrderSetID(t *testing.T) {
	svc := newTestService()
	sec := &OrderSetSection{Name: "Antibiotics"}
	if err := svc.AddOrderSetSection(nil, sec); err == nil {
		t.Error("expected error for missing order_set_id")
	}
}

func TestService_AddOrderSetSection_MissingName(t *testing.T) {
	svc := newTestService()
	sec := &OrderSetSection{OrderSetID: uuid.New()}
	if err := svc.AddOrderSetSection(nil, sec); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestService_GetOrderSetSections(t *testing.T) {
	svc := newTestService()
	o := &OrderSet{Name: "Sepsis"}
	svc.CreateOrderSet(nil, o)
	svc.AddOrderSetSection(nil, &OrderSetSection{OrderSetID: o.ID, Name: "S1"})
	svc.AddOrderSetSection(nil, &OrderSetSection{OrderSetID: o.ID, Name: "S2"})
	sections, err := svc.GetOrderSetSections(nil, o.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sections) != 2 {
		t.Errorf("expected 2, got %d", len(sections))
	}
}

func TestService_AddOrderSetItem(t *testing.T) {
	svc := newTestService()
	secID := uuid.New()
	item := &OrderSetItem{SectionID: secID, ItemName: "Ceftriaxone"}
	if err := svc.AddOrderSetItem(nil, item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_AddOrderSetItem_MissingSectionID(t *testing.T) {
	svc := newTestService()
	item := &OrderSetItem{ItemName: "Ceftriaxone"}
	if err := svc.AddOrderSetItem(nil, item); err == nil {
		t.Error("expected error for missing section_id")
	}
}

func TestService_AddOrderSetItem_MissingItemName(t *testing.T) {
	svc := newTestService()
	item := &OrderSetItem{SectionID: uuid.New()}
	if err := svc.AddOrderSetItem(nil, item); err == nil {
		t.Error("expected error for missing item_name")
	}
}

func TestService_GetOrderSetItems(t *testing.T) {
	svc := newTestService()
	secID := uuid.New()
	svc.AddOrderSetItem(nil, &OrderSetItem{SectionID: secID, ItemName: "I1"})
	items, err := svc.GetOrderSetItems(nil, secID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

// ── Clinical Pathway Tests ──

func TestService_CreateClinicalPathway(t *testing.T) {
	svc := newTestService()
	p := &ClinicalPathway{Name: "Heart Failure"}
	if err := svc.CreateClinicalPathway(nil, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestService_CreateClinicalPathway_MissingName(t *testing.T) {
	svc := newTestService()
	p := &ClinicalPathway{}
	if err := svc.CreateClinicalPathway(nil, p); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestService_GetClinicalPathway(t *testing.T) {
	svc := newTestService()
	p := &ClinicalPathway{Name: "Heart Failure"}
	svc.CreateClinicalPathway(nil, p)
	got, err := svc.GetClinicalPathway(nil, p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Heart Failure" {
		t.Errorf("expected 'Heart Failure', got %s", got.Name)
	}
}

func TestService_DeleteClinicalPathway(t *testing.T) {
	svc := newTestService()
	p := &ClinicalPathway{Name: "Heart Failure"}
	svc.CreateClinicalPathway(nil, p)
	if err := svc.DeleteClinicalPathway(nil, p.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_AddPathwayPhase(t *testing.T) {
	svc := newTestService()
	p := &ClinicalPathway{Name: "HF"}
	svc.CreateClinicalPathway(nil, p)
	phase := &ClinicalPathwayPhase{PathwayID: p.ID, Name: "Acute Phase"}
	if err := svc.AddPathwayPhase(nil, phase); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_AddPathwayPhase_MissingPathwayID(t *testing.T) {
	svc := newTestService()
	phase := &ClinicalPathwayPhase{Name: "Phase"}
	if err := svc.AddPathwayPhase(nil, phase); err == nil {
		t.Error("expected error for missing pathway_id")
	}
}

func TestService_AddPathwayPhase_MissingName(t *testing.T) {
	svc := newTestService()
	phase := &ClinicalPathwayPhase{PathwayID: uuid.New()}
	if err := svc.AddPathwayPhase(nil, phase); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestService_GetPathwayPhases(t *testing.T) {
	svc := newTestService()
	p := &ClinicalPathway{Name: "HF"}
	svc.CreateClinicalPathway(nil, p)
	svc.AddPathwayPhase(nil, &ClinicalPathwayPhase{PathwayID: p.ID, Name: "P1"})
	svc.AddPathwayPhase(nil, &ClinicalPathwayPhase{PathwayID: p.ID, Name: "P2"})
	phases, err := svc.GetPathwayPhases(nil, p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(phases) != 2 {
		t.Errorf("expected 2, got %d", len(phases))
	}
}

// ── Pathway Enrollment Tests ──

func TestService_CreatePathwayEnrollment(t *testing.T) {
	svc := newTestService()
	e := &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: uuid.New()}
	if err := svc.CreatePathwayEnrollment(nil, e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Status != "active" {
		t.Errorf("expected default status 'active', got %s", e.Status)
	}
}

func TestService_CreatePathwayEnrollment_MissingPathwayID(t *testing.T) {
	svc := newTestService()
	e := &PatientPathwayEnrollment{PatientID: uuid.New()}
	if err := svc.CreatePathwayEnrollment(nil, e); err == nil {
		t.Error("expected error for missing pathway_id")
	}
}

func TestService_CreatePathwayEnrollment_MissingPatientID(t *testing.T) {
	svc := newTestService()
	e := &PatientPathwayEnrollment{PathwayID: uuid.New()}
	if err := svc.CreatePathwayEnrollment(nil, e); err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestService_CreatePathwayEnrollment_InvalidStatus(t *testing.T) {
	svc := newTestService()
	e := &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: uuid.New(), Status: "bogus"}
	if err := svc.CreatePathwayEnrollment(nil, e); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_CreatePathwayEnrollment_ValidStatuses(t *testing.T) {
	statuses := []string{"active", "completed", "withdrawn", "deviated"}
	for _, status := range statuses {
		svc := newTestService()
		e := &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: uuid.New(), Status: status}
		if err := svc.CreatePathwayEnrollment(nil, e); err != nil {
			t.Errorf("status %q should be valid, got error: %v", status, err)
		}
	}
}

func TestService_GetPathwayEnrollment(t *testing.T) {
	svc := newTestService()
	e := &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: uuid.New()}
	svc.CreatePathwayEnrollment(nil, e)
	got, err := svc.GetPathwayEnrollment(nil, e.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != e.ID {
		t.Error("ID mismatch")
	}
}

func TestService_UpdatePathwayEnrollment(t *testing.T) {
	svc := newTestService()
	e := &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: uuid.New()}
	svc.CreatePathwayEnrollment(nil, e)
	e.Status = "completed"
	if err := svc.UpdatePathwayEnrollment(nil, e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdatePathwayEnrollment_InvalidStatus(t *testing.T) {
	svc := newTestService()
	e := &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: uuid.New()}
	svc.CreatePathwayEnrollment(nil, e)
	e.Status = "bad"
	if err := svc.UpdatePathwayEnrollment(nil, e); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_DeletePathwayEnrollment(t *testing.T) {
	svc := newTestService()
	e := &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: uuid.New()}
	svc.CreatePathwayEnrollment(nil, e)
	if err := svc.DeletePathwayEnrollment(nil, e.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListPathwayEnrollmentsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreatePathwayEnrollment(nil, &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: pid})
	svc.CreatePathwayEnrollment(nil, &PatientPathwayEnrollment{PathwayID: uuid.New(), PatientID: uuid.New()})
	items, total, err := svc.ListPathwayEnrollmentsByPatient(nil, pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

// ── Formulary Tests ──

func TestService_CreateFormulary(t *testing.T) {
	svc := newTestService()
	f := &Formulary{Name: "2025 Formulary"}
	if err := svc.CreateFormulary(nil, f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestService_CreateFormulary_MissingName(t *testing.T) {
	svc := newTestService()
	f := &Formulary{}
	if err := svc.CreateFormulary(nil, f); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestService_GetFormulary(t *testing.T) {
	svc := newTestService()
	f := &Formulary{Name: "2025 Formulary"}
	svc.CreateFormulary(nil, f)
	got, err := svc.GetFormulary(nil, f.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "2025 Formulary" {
		t.Errorf("expected '2025 Formulary', got %s", got.Name)
	}
}

func TestService_DeleteFormulary(t *testing.T) {
	svc := newTestService()
	f := &Formulary{Name: "2025 Formulary"}
	svc.CreateFormulary(nil, f)
	if err := svc.DeleteFormulary(nil, f.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_AddFormularyItem(t *testing.T) {
	svc := newTestService()
	f := &Formulary{Name: "F"}
	svc.CreateFormulary(nil, f)
	item := &FormularyItem{FormularyID: f.ID, MedicationName: "Metformin"}
	if err := svc.AddFormularyItem(nil, item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_AddFormularyItem_MissingFormularyID(t *testing.T) {
	svc := newTestService()
	item := &FormularyItem{MedicationName: "Metformin"}
	if err := svc.AddFormularyItem(nil, item); err == nil {
		t.Error("expected error for missing formulary_id")
	}
}

func TestService_AddFormularyItem_MissingMedicationName(t *testing.T) {
	svc := newTestService()
	item := &FormularyItem{FormularyID: uuid.New()}
	if err := svc.AddFormularyItem(nil, item); err == nil {
		t.Error("expected error for missing medication_name")
	}
}

func TestService_GetFormularyItems(t *testing.T) {
	svc := newTestService()
	f := &Formulary{Name: "F"}
	svc.CreateFormulary(nil, f)
	svc.AddFormularyItem(nil, &FormularyItem{FormularyID: f.ID, MedicationName: "M1"})
	svc.AddFormularyItem(nil, &FormularyItem{FormularyID: f.ID, MedicationName: "M2"})
	items, err := svc.GetFormularyItems(nil, f.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}

// ── Medication Reconciliation Tests ──

func TestService_CreateMedReconciliation(t *testing.T) {
	svc := newTestService()
	mr := &MedicationReconciliation{PatientID: uuid.New()}
	if err := svc.CreateMedReconciliation(nil, mr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mr.Status != "in-progress" {
		t.Errorf("expected default status 'in-progress', got %s", mr.Status)
	}
}

func TestService_CreateMedReconciliation_MissingPatientID(t *testing.T) {
	svc := newTestService()
	mr := &MedicationReconciliation{}
	if err := svc.CreateMedReconciliation(nil, mr); err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestService_CreateMedReconciliation_InvalidStatus(t *testing.T) {
	svc := newTestService()
	mr := &MedicationReconciliation{PatientID: uuid.New(), Status: "bogus"}
	if err := svc.CreateMedReconciliation(nil, mr); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_CreateMedReconciliation_ValidStatuses(t *testing.T) {
	statuses := []string{"in-progress", "completed", "pending-verification"}
	for _, status := range statuses {
		svc := newTestService()
		mr := &MedicationReconciliation{PatientID: uuid.New(), Status: status}
		if err := svc.CreateMedReconciliation(nil, mr); err != nil {
			t.Errorf("status %q should be valid, got error: %v", status, err)
		}
	}
}

func TestService_GetMedReconciliation(t *testing.T) {
	svc := newTestService()
	mr := &MedicationReconciliation{PatientID: uuid.New()}
	svc.CreateMedReconciliation(nil, mr)
	got, err := svc.GetMedReconciliation(nil, mr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != mr.ID {
		t.Error("ID mismatch")
	}
}

func TestService_UpdateMedReconciliation(t *testing.T) {
	svc := newTestService()
	mr := &MedicationReconciliation{PatientID: uuid.New()}
	svc.CreateMedReconciliation(nil, mr)
	mr.Status = "completed"
	if err := svc.UpdateMedReconciliation(nil, mr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdateMedReconciliation_InvalidStatus(t *testing.T) {
	svc := newTestService()
	mr := &MedicationReconciliation{PatientID: uuid.New()}
	svc.CreateMedReconciliation(nil, mr)
	mr.Status = "bad"
	if err := svc.UpdateMedReconciliation(nil, mr); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_DeleteMedReconciliation(t *testing.T) {
	svc := newTestService()
	mr := &MedicationReconciliation{PatientID: uuid.New()}
	svc.CreateMedReconciliation(nil, mr)
	if err := svc.DeleteMedReconciliation(nil, mr.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListMedReconciliationsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateMedReconciliation(nil, &MedicationReconciliation{PatientID: pid})
	svc.CreateMedReconciliation(nil, &MedicationReconciliation{PatientID: uuid.New()})
	items, total, err := svc.ListMedReconciliationsByPatient(nil, pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1, got %d", total)
	}
	if len(items) != 1 {
		t.Errorf("expected 1, got %d", len(items))
	}
}

func TestService_AddMedReconciliationItem(t *testing.T) {
	svc := newTestService()
	mr := &MedicationReconciliation{PatientID: uuid.New()}
	svc.CreateMedReconciliation(nil, mr)
	item := &MedicationReconciliationItem{ReconciliationID: mr.ID, MedicationName: "Lisinopril"}
	if err := svc.AddMedReconciliationItem(nil, item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_AddMedReconciliationItem_MissingReconciliationID(t *testing.T) {
	svc := newTestService()
	item := &MedicationReconciliationItem{MedicationName: "Lisinopril"}
	if err := svc.AddMedReconciliationItem(nil, item); err == nil {
		t.Error("expected error for missing reconciliation_id")
	}
}

func TestService_AddMedReconciliationItem_MissingMedicationName(t *testing.T) {
	svc := newTestService()
	item := &MedicationReconciliationItem{ReconciliationID: uuid.New()}
	if err := svc.AddMedReconciliationItem(nil, item); err == nil {
		t.Error("expected error for missing medication_name")
	}
}

func TestService_GetMedReconciliationItems(t *testing.T) {
	svc := newTestService()
	mr := &MedicationReconciliation{PatientID: uuid.New()}
	svc.CreateMedReconciliation(nil, mr)
	svc.AddMedReconciliationItem(nil, &MedicationReconciliationItem{ReconciliationID: mr.ID, MedicationName: "M1"})
	svc.AddMedReconciliationItem(nil, &MedicationReconciliationItem{ReconciliationID: mr.ID, MedicationName: "M2"})
	items, err := svc.GetMedReconciliationItems(nil, mr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2, got %d", len(items))
	}
}
