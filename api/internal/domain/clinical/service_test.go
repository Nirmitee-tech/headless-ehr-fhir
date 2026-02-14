package clinical

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// =========== Mock Repositories ===========

type mockConditionRepo struct {
	store map[uuid.UUID]*Condition
}

func newMockConditionRepo() *mockConditionRepo {
	return &mockConditionRepo{store: make(map[uuid.UUID]*Condition)}
}

func (m *mockConditionRepo) Create(_ context.Context, c *Condition) error {
	c.ID = uuid.New()
	if c.FHIRID == "" {
		c.FHIRID = c.ID.String()
	}
	m.store[c.ID] = c
	return nil
}

func (m *mockConditionRepo) GetByID(_ context.Context, id uuid.UUID) (*Condition, error) {
	c, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (m *mockConditionRepo) GetByFHIRID(_ context.Context, fhirID string) (*Condition, error) {
	for _, c := range m.store {
		if c.FHIRID == fhirID {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockConditionRepo) Update(_ context.Context, c *Condition) error {
	if _, ok := m.store[c.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[c.ID] = c
	return nil
}

func (m *mockConditionRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockConditionRepo) List(_ context.Context, limit, offset int) ([]*Condition, int, error) {
	var result []*Condition
	for _, c := range m.store {
		result = append(result, c)
	}
	return result, len(result), nil
}

func (m *mockConditionRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Condition, int, error) {
	var result []*Condition
	for _, c := range m.store {
		if c.PatientID == patientID {
			result = append(result, c)
		}
	}
	return result, len(result), nil
}

func (m *mockConditionRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Condition, int, error) {
	var result []*Condition
	for _, c := range m.store {
		result = append(result, c)
	}
	return result, len(result), nil
}

// -- Mock Observation Repo --

type mockObservationRepo struct {
	store      map[uuid.UUID]*Observation
	components map[uuid.UUID][]*ObservationComponent
}

func newMockObservationRepo() *mockObservationRepo {
	return &mockObservationRepo{
		store:      make(map[uuid.UUID]*Observation),
		components: make(map[uuid.UUID][]*ObservationComponent),
	}
}

func (m *mockObservationRepo) Create(_ context.Context, o *Observation) error {
	o.ID = uuid.New()
	if o.FHIRID == "" {
		o.FHIRID = o.ID.String()
	}
	m.store[o.ID] = o
	return nil
}

func (m *mockObservationRepo) GetByID(_ context.Context, id uuid.UUID) (*Observation, error) {
	o, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return o, nil
}

func (m *mockObservationRepo) GetByFHIRID(_ context.Context, fhirID string) (*Observation, error) {
	for _, o := range m.store {
		if o.FHIRID == fhirID {
			return o, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockObservationRepo) Update(_ context.Context, o *Observation) error {
	if _, ok := m.store[o.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[o.ID] = o
	return nil
}

func (m *mockObservationRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockObservationRepo) List(_ context.Context, limit, offset int) ([]*Observation, int, error) {
	var result []*Observation
	for _, o := range m.store {
		result = append(result, o)
	}
	return result, len(result), nil
}

func (m *mockObservationRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Observation, int, error) {
	var result []*Observation
	for _, o := range m.store {
		if o.PatientID == patientID {
			result = append(result, o)
		}
	}
	return result, len(result), nil
}

func (m *mockObservationRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Observation, int, error) {
	var result []*Observation
	for _, o := range m.store {
		result = append(result, o)
	}
	return result, len(result), nil
}

func (m *mockObservationRepo) AddComponent(_ context.Context, c *ObservationComponent) error {
	c.ID = uuid.New()
	m.components[c.ObservationID] = append(m.components[c.ObservationID], c)
	return nil
}

func (m *mockObservationRepo) GetComponents(_ context.Context, obsID uuid.UUID) ([]*ObservationComponent, error) {
	return m.components[obsID], nil
}

// -- Mock Allergy Repo --

type mockAllergyRepo struct {
	store     map[uuid.UUID]*AllergyIntolerance
	reactions map[uuid.UUID][]*AllergyReaction
}

func newMockAllergyRepo() *mockAllergyRepo {
	return &mockAllergyRepo{
		store:     make(map[uuid.UUID]*AllergyIntolerance),
		reactions: make(map[uuid.UUID][]*AllergyReaction),
	}
}

func (m *mockAllergyRepo) Create(_ context.Context, a *AllergyIntolerance) error {
	a.ID = uuid.New()
	if a.FHIRID == "" {
		a.FHIRID = a.ID.String()
	}
	m.store[a.ID] = a
	return nil
}

func (m *mockAllergyRepo) GetByID(_ context.Context, id uuid.UUID) (*AllergyIntolerance, error) {
	a, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return a, nil
}

func (m *mockAllergyRepo) GetByFHIRID(_ context.Context, fhirID string) (*AllergyIntolerance, error) {
	for _, a := range m.store {
		if a.FHIRID == fhirID {
			return a, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockAllergyRepo) Update(_ context.Context, a *AllergyIntolerance) error {
	if _, ok := m.store[a.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[a.ID] = a
	return nil
}

func (m *mockAllergyRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockAllergyRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*AllergyIntolerance, int, error) {
	var result []*AllergyIntolerance
	for _, a := range m.store {
		if a.PatientID == patientID {
			result = append(result, a)
		}
	}
	return result, len(result), nil
}

func (m *mockAllergyRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*AllergyIntolerance, int, error) {
	var result []*AllergyIntolerance
	for _, a := range m.store {
		result = append(result, a)
	}
	return result, len(result), nil
}

func (m *mockAllergyRepo) AddReaction(_ context.Context, r *AllergyReaction) error {
	r.ID = uuid.New()
	m.reactions[r.AllergyID] = append(m.reactions[r.AllergyID], r)
	return nil
}

func (m *mockAllergyRepo) GetReactions(_ context.Context, allergyID uuid.UUID) ([]*AllergyReaction, error) {
	return m.reactions[allergyID], nil
}

func (m *mockAllergyRepo) RemoveReaction(_ context.Context, id uuid.UUID) error {
	for aid, reactions := range m.reactions {
		for i, r := range reactions {
			if r.ID == id {
				m.reactions[aid] = append(reactions[:i], reactions[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

// -- Mock Procedure Repo --

type mockProcedureRepo struct {
	store      map[uuid.UUID]*ProcedureRecord
	performers map[uuid.UUID][]*ProcedurePerformer
}

func newMockProcedureRepo() *mockProcedureRepo {
	return &mockProcedureRepo{
		store:      make(map[uuid.UUID]*ProcedureRecord),
		performers: make(map[uuid.UUID][]*ProcedurePerformer),
	}
}

func (m *mockProcedureRepo) Create(_ context.Context, p *ProcedureRecord) error {
	p.ID = uuid.New()
	if p.FHIRID == "" {
		p.FHIRID = p.ID.String()
	}
	m.store[p.ID] = p
	return nil
}

func (m *mockProcedureRepo) GetByID(_ context.Context, id uuid.UUID) (*ProcedureRecord, error) {
	p, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return p, nil
}

func (m *mockProcedureRepo) GetByFHIRID(_ context.Context, fhirID string) (*ProcedureRecord, error) {
	for _, p := range m.store {
		if p.FHIRID == fhirID {
			return p, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockProcedureRepo) Update(_ context.Context, p *ProcedureRecord) error {
	if _, ok := m.store[p.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[p.ID] = p
	return nil
}

func (m *mockProcedureRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockProcedureRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*ProcedureRecord, int, error) {
	var result []*ProcedureRecord
	for _, p := range m.store {
		if p.PatientID == patientID {
			result = append(result, p)
		}
	}
	return result, len(result), nil
}

func (m *mockProcedureRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*ProcedureRecord, int, error) {
	var result []*ProcedureRecord
	for _, p := range m.store {
		result = append(result, p)
	}
	return result, len(result), nil
}

func (m *mockProcedureRepo) AddPerformer(_ context.Context, pf *ProcedurePerformer) error {
	pf.ID = uuid.New()
	m.performers[pf.ProcedureID] = append(m.performers[pf.ProcedureID], pf)
	return nil
}

func (m *mockProcedureRepo) GetPerformers(_ context.Context, procID uuid.UUID) ([]*ProcedurePerformer, error) {
	return m.performers[procID], nil
}

func (m *mockProcedureRepo) RemovePerformer(_ context.Context, id uuid.UUID) error {
	for pid, performers := range m.performers {
		for i, pf := range performers {
			if pf.ID == id {
				m.performers[pid] = append(performers[:i], performers[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

// =========== Helper ===========

func newTestService() *Service {
	return NewService(
		newMockConditionRepo(),
		newMockObservationRepo(),
		newMockAllergyRepo(),
		newMockProcedureRepo(),
	)
}

// =========== Condition Tests ===========

func TestCreateCondition_Success(t *testing.T) {
	svc := newTestService()
	c := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	if err := svc.CreateCondition(context.Background(), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ClinicalStatus != "active" {
		t.Errorf("expected default status 'active', got %q", c.ClinicalStatus)
	}
}

func TestCreateCondition_MissingPatient(t *testing.T) {
	svc := newTestService()
	c := &Condition{CodeValue: "J06.9", CodeDisplay: "URI"}
	if err := svc.CreateCondition(context.Background(), c); err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestCreateCondition_MissingCode(t *testing.T) {
	svc := newTestService()
	c := &Condition{PatientID: uuid.New(), CodeDisplay: "URI"}
	if err := svc.CreateCondition(context.Background(), c); err == nil {
		t.Fatal("expected error for missing code_value")
	}
}

func TestCreateCondition_MissingDisplay(t *testing.T) {
	svc := newTestService()
	c := &Condition{PatientID: uuid.New(), CodeValue: "J06.9"}
	if err := svc.CreateCondition(context.Background(), c); err == nil {
		t.Fatal("expected error for missing code_display")
	}
}

func TestCreateCondition_InvalidStatus(t *testing.T) {
	svc := newTestService()
	c := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI", ClinicalStatus: "bogus"}
	if err := svc.CreateCondition(context.Background(), c); err == nil {
		t.Fatal("expected error for invalid clinical_status")
	}
}

func TestCreateCondition_ExplicitStatus(t *testing.T) {
	svc := newTestService()
	c := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI", ClinicalStatus: "resolved"}
	if err := svc.CreateCondition(context.Background(), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ClinicalStatus != "resolved" {
		t.Errorf("expected 'resolved', got %q", c.ClinicalStatus)
	}
}

func TestGetCondition(t *testing.T) {
	svc := newTestService()
	c := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	svc.CreateCondition(context.Background(), c)

	got, err := svc.GetCondition(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != c.ID {
		t.Errorf("expected ID %v, got %v", c.ID, got.ID)
	}
}

func TestGetCondition_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetCondition(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateCondition_InvalidStatus(t *testing.T) {
	svc := newTestService()
	c := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	svc.CreateCondition(context.Background(), c)
	c.ClinicalStatus = "invalid"
	if err := svc.UpdateCondition(context.Background(), c); err == nil {
		t.Fatal("expected error for invalid clinical_status")
	}
}

func TestDeleteCondition(t *testing.T) {
	svc := newTestService()
	c := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	svc.CreateCondition(context.Background(), c)
	if err := svc.DeleteCondition(context.Background(), c.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetCondition(context.Background(), c.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListConditionsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateCondition(context.Background(), &Condition{PatientID: pid, CodeValue: "J06.9", CodeDisplay: "URI"})
	svc.CreateCondition(context.Background(), &Condition{PatientID: pid, CodeValue: "E11.9", CodeDisplay: "T2DM"})
	svc.CreateCondition(context.Background(), &Condition{PatientID: uuid.New(), CodeValue: "I10", CodeDisplay: "HTN"})

	items, total, err := svc.ListConditionsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 conditions, got %d", total)
	}
}

// =========== Observation Tests ===========

func TestCreateObservation_Success(t *testing.T) {
	svc := newTestService()
	o := &Observation{PatientID: uuid.New(), CodeValue: "8310-5"}
	if err := svc.CreateObservation(context.Background(), o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Status != "final" {
		t.Errorf("expected default status 'final', got %q", o.Status)
	}
}

func TestCreateObservation_MissingPatient(t *testing.T) {
	svc := newTestService()
	o := &Observation{CodeValue: "8310-5"}
	if err := svc.CreateObservation(context.Background(), o); err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestCreateObservation_MissingCode(t *testing.T) {
	svc := newTestService()
	o := &Observation{PatientID: uuid.New()}
	if err := svc.CreateObservation(context.Background(), o); err == nil {
		t.Fatal("expected error for missing code_value")
	}
}

func TestCreateObservation_InvalidStatus(t *testing.T) {
	svc := newTestService()
	o := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", Status: "bogus"}
	if err := svc.CreateObservation(context.Background(), o); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCreateObservation_ValidStatuses(t *testing.T) {
	for _, s := range []string{"registered", "preliminary", "final", "amended", "corrected", "cancelled", "entered-in-error", "unknown"} {
		svc := newTestService()
		o := &Observation{PatientID: uuid.New(), CodeValue: "8310-5", Status: s}
		if err := svc.CreateObservation(context.Background(), o); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

func TestAddObservationComponent(t *testing.T) {
	svc := newTestService()
	o := &Observation{PatientID: uuid.New(), CodeValue: "85354-9"}
	svc.CreateObservation(context.Background(), o)

	comp := &ObservationComponent{ObservationID: o.ID, CodeValue: "8480-6"}
	if err := svc.AddObservationComponent(context.Background(), comp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comps, err := svc.GetObservationComponents(context.Background(), o.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comps) != 1 {
		t.Errorf("expected 1 component, got %d", len(comps))
	}
}

func TestAddObservationComponent_MissingObsID(t *testing.T) {
	svc := newTestService()
	comp := &ObservationComponent{CodeValue: "8480-6"}
	if err := svc.AddObservationComponent(context.Background(), comp); err == nil {
		t.Fatal("expected error for missing observation_id")
	}
}

func TestAddObservationComponent_MissingCode(t *testing.T) {
	svc := newTestService()
	comp := &ObservationComponent{ObservationID: uuid.New()}
	if err := svc.AddObservationComponent(context.Background(), comp); err == nil {
		t.Fatal("expected error for missing code_value")
	}
}

// =========== Allergy Tests ===========

func TestCreateAllergy_Success(t *testing.T) {
	svc := newTestService()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	if err := svc.CreateAllergy(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ClinicalStatus == nil || *a.ClinicalStatus != "active" {
		t.Error("expected default clinical_status 'active'")
	}
}

func TestCreateAllergy_MissingPatient(t *testing.T) {
	svc := newTestService()
	a := &AllergyIntolerance{}
	if err := svc.CreateAllergy(context.Background(), a); err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestCreateAllergy_ExplicitStatus(t *testing.T) {
	svc := newTestService()
	status := "resolved"
	a := &AllergyIntolerance{PatientID: uuid.New(), ClinicalStatus: &status}
	if err := svc.CreateAllergy(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *a.ClinicalStatus != "resolved" {
		t.Errorf("expected 'resolved', got %q", *a.ClinicalStatus)
	}
}

func TestAddAllergyReaction(t *testing.T) {
	svc := newTestService()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	svc.CreateAllergy(context.Background(), a)

	r := &AllergyReaction{AllergyID: a.ID, ManifestationCode: "39579001"}
	if err := svc.AddAllergyReaction(context.Background(), r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reactions, err := svc.GetAllergyReactions(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reactions) != 1 {
		t.Errorf("expected 1 reaction, got %d", len(reactions))
	}
}

func TestAddAllergyReaction_MissingAllergyID(t *testing.T) {
	svc := newTestService()
	r := &AllergyReaction{ManifestationCode: "39579001"}
	if err := svc.AddAllergyReaction(context.Background(), r); err == nil {
		t.Fatal("expected error for missing allergy_id")
	}
}

func TestAddAllergyReaction_MissingCode(t *testing.T) {
	svc := newTestService()
	r := &AllergyReaction{AllergyID: uuid.New()}
	if err := svc.AddAllergyReaction(context.Background(), r); err == nil {
		t.Fatal("expected error for missing manifestation_code")
	}
}

func TestRemoveAllergyReaction(t *testing.T) {
	svc := newTestService()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	svc.CreateAllergy(context.Background(), a)

	r := &AllergyReaction{AllergyID: a.ID, ManifestationCode: "39579001"}
	svc.AddAllergyReaction(context.Background(), r)

	if err := svc.RemoveAllergyReaction(context.Background(), r.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	reactions, _ := svc.GetAllergyReactions(context.Background(), a.ID)
	if len(reactions) != 0 {
		t.Errorf("expected 0 reactions after remove, got %d", len(reactions))
	}
}

// =========== Procedure Tests ===========

func TestCreateProcedure_Success(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002"}
	if err := svc.CreateProcedure(context.Background(), p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != "completed" {
		t.Errorf("expected default status 'completed', got %q", p.Status)
	}
}

func TestCreateProcedure_MissingPatient(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{CodeValue: "80146002"}
	if err := svc.CreateProcedure(context.Background(), p); err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestCreateProcedure_MissingCode(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{PatientID: uuid.New()}
	if err := svc.CreateProcedure(context.Background(), p); err == nil {
		t.Fatal("expected error for missing code_value")
	}
}

func TestCreateProcedure_InvalidStatus(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", Status: "bogus"}
	if err := svc.CreateProcedure(context.Background(), p); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCreateProcedure_ValidStatuses(t *testing.T) {
	for _, s := range []string{"preparation", "in-progress", "not-done", "on-hold", "stopped", "completed", "entered-in-error", "unknown"} {
		svc := newTestService()
		p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002", Status: s}
		if err := svc.CreateProcedure(context.Background(), p); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

func TestAddProcedurePerformer(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002"}
	svc.CreateProcedure(context.Background(), p)

	pf := &ProcedurePerformer{ProcedureID: p.ID, PractitionerID: uuid.New()}
	if err := svc.AddProcedurePerformer(context.Background(), pf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	performers, err := svc.GetProcedurePerformers(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(performers) != 1 {
		t.Errorf("expected 1 performer, got %d", len(performers))
	}
}

func TestAddProcedurePerformer_MissingProcedureID(t *testing.T) {
	svc := newTestService()
	pf := &ProcedurePerformer{PractitionerID: uuid.New()}
	if err := svc.AddProcedurePerformer(context.Background(), pf); err == nil {
		t.Fatal("expected error for missing procedure_id")
	}
}

func TestAddProcedurePerformer_MissingPractitionerID(t *testing.T) {
	svc := newTestService()
	pf := &ProcedurePerformer{ProcedureID: uuid.New()}
	if err := svc.AddProcedurePerformer(context.Background(), pf); err == nil {
		t.Fatal("expected error for missing practitioner_id")
	}
}

func TestRemoveProcedurePerformer(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002"}
	svc.CreateProcedure(context.Background(), p)

	pf := &ProcedurePerformer{ProcedureID: p.ID, PractitionerID: uuid.New()}
	svc.AddProcedurePerformer(context.Background(), pf)

	if err := svc.RemoveProcedurePerformer(context.Background(), pf.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	performers, _ := svc.GetProcedurePerformers(context.Background(), p.ID)
	if len(performers) != 0 {
		t.Errorf("expected 0 performers after remove, got %d", len(performers))
	}
}

// =========== FHIR Conversion Tests ===========

func TestConditionToFHIR(t *testing.T) {
	c := &Condition{
		FHIRID:         "cond-1",
		PatientID:      uuid.New(),
		ClinicalStatus: "active",
		CodeValue:      "J06.9",
		CodeDisplay:    "Acute upper respiratory infection",
	}
	f := c.ToFHIR()
	if f["resourceType"] != "Condition" {
		t.Errorf("expected resourceType 'Condition', got %v", f["resourceType"])
	}
	if f["id"] != "cond-1" {
		t.Errorf("expected id 'cond-1', got %v", f["id"])
	}
}

func TestObservationToFHIR(t *testing.T) {
	qty := float64(98.6)
	o := &Observation{
		FHIRID:        "obs-1",
		PatientID:     uuid.New(),
		Status:        "final",
		CodeValue:     "8310-5",
		CodeDisplay:   "Body temperature",
		ValueQuantity: &qty,
	}
	unit := "degF"
	o.ValueUnit = &unit
	f := o.ToFHIR()
	if f["resourceType"] != "Observation" {
		t.Errorf("expected resourceType 'Observation', got %v", f["resourceType"])
	}
	vq, ok := f["valueQuantity"].(map[string]interface{})
	if !ok {
		t.Fatal("expected valueQuantity map")
	}
	if vq["value"] != qty {
		t.Errorf("expected value %v, got %v", qty, vq["value"])
	}
}

func TestAllergyToFHIR(t *testing.T) {
	status := "active"
	a := &AllergyIntolerance{
		FHIRID:         "allergy-1",
		PatientID:      uuid.New(),
		ClinicalStatus: &status,
	}
	f := a.ToFHIR()
	if f["resourceType"] != "AllergyIntolerance" {
		t.Errorf("expected resourceType 'AllergyIntolerance', got %v", f["resourceType"])
	}
}

func TestProcedureToFHIR(t *testing.T) {
	p := &ProcedureRecord{
		FHIRID:      "proc-1",
		PatientID:   uuid.New(),
		Status:      "completed",
		CodeValue:   "80146002",
		CodeDisplay: "Appendectomy",
	}
	f := p.ToFHIR()
	if f["resourceType"] != "Procedure" {
		t.Errorf("expected resourceType 'Procedure', got %v", f["resourceType"])
	}
	if f["status"] != "completed" {
		t.Errorf("expected status 'completed', got %v", f["status"])
	}
}

// =========== Additional Condition Tests ===========

func TestGetConditionByFHIRID(t *testing.T) {
	svc := newTestService()
	c := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	svc.CreateCondition(context.Background(), c)

	got, err := svc.GetConditionByFHIRID(context.Background(), c.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != c.ID {
		t.Errorf("expected ID %v, got %v", c.ID, got.ID)
	}
}

func TestGetConditionByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetConditionByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateCondition_ValidStatus(t *testing.T) {
	svc := newTestService()
	c := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI"}
	svc.CreateCondition(context.Background(), c)
	c.ClinicalStatus = "resolved"
	if err := svc.UpdateCondition(context.Background(), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateCondition_NotFound(t *testing.T) {
	svc := newTestService()
	c := &Condition{ID: uuid.New(), ClinicalStatus: "active"}
	if err := svc.UpdateCondition(context.Background(), c); err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestSearchConditions(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateCondition(context.Background(), &Condition{PatientID: pid, CodeValue: "J06.9", CodeDisplay: "URI"})
	items, total, err := svc.SearchConditions(context.Background(), map[string]string{"code": "J06.9"}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(items) < 1 {
		t.Error("expected items")
	}
}

func TestCreateCondition_ValidStatuses(t *testing.T) {
	for _, s := range []string{"active", "recurrence", "relapse", "inactive", "remission", "resolved"} {
		svc := newTestService()
		c := &Condition{PatientID: uuid.New(), CodeValue: "J06.9", CodeDisplay: "URI", ClinicalStatus: s}
		if err := svc.CreateCondition(context.Background(), c); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

// =========== Additional Observation Tests ===========

func TestGetObservation(t *testing.T) {
	svc := newTestService()
	o := &Observation{PatientID: uuid.New(), CodeValue: "8310-5"}
	svc.CreateObservation(context.Background(), o)

	got, err := svc.GetObservation(context.Background(), o.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != o.ID {
		t.Errorf("expected ID %v, got %v", o.ID, got.ID)
	}
}

func TestGetObservation_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetObservation(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetObservationByFHIRID(t *testing.T) {
	svc := newTestService()
	o := &Observation{PatientID: uuid.New(), CodeValue: "8310-5"}
	svc.CreateObservation(context.Background(), o)

	got, err := svc.GetObservationByFHIRID(context.Background(), o.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != o.ID {
		t.Errorf("expected ID %v, got %v", o.ID, got.ID)
	}
}

func TestGetObservationByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetObservationByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateObservation(t *testing.T) {
	svc := newTestService()
	o := &Observation{PatientID: uuid.New(), CodeValue: "8310-5"}
	svc.CreateObservation(context.Background(), o)
	o.Status = "amended"
	if err := svc.UpdateObservation(context.Background(), o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateObservation_NotFound(t *testing.T) {
	svc := newTestService()
	o := &Observation{ID: uuid.New()}
	if err := svc.UpdateObservation(context.Background(), o); err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestDeleteObservation(t *testing.T) {
	svc := newTestService()
	o := &Observation{PatientID: uuid.New(), CodeValue: "8310-5"}
	svc.CreateObservation(context.Background(), o)
	if err := svc.DeleteObservation(context.Background(), o.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetObservation(context.Background(), o.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListObservationsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateObservation(context.Background(), &Observation{PatientID: pid, CodeValue: "8310-5"})
	svc.CreateObservation(context.Background(), &Observation{PatientID: pid, CodeValue: "8462-4"})
	svc.CreateObservation(context.Background(), &Observation{PatientID: uuid.New(), CodeValue: "8480-6"})

	items, total, err := svc.ListObservationsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 observations, got %d", total)
	}
}

func TestSearchObservations(t *testing.T) {
	svc := newTestService()
	svc.CreateObservation(context.Background(), &Observation{PatientID: uuid.New(), CodeValue: "8310-5"})
	items, total, err := svc.SearchObservations(context.Background(), map[string]string{"code": "8310-5"}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(items) < 1 {
		t.Error("expected items")
	}
}

// =========== Additional Allergy Tests ===========

func TestGetAllergy(t *testing.T) {
	svc := newTestService()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	svc.CreateAllergy(context.Background(), a)

	got, err := svc.GetAllergy(context.Background(), a.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != a.ID {
		t.Errorf("expected ID %v, got %v", a.ID, got.ID)
	}
}

func TestGetAllergy_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetAllergy(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetAllergyByFHIRID(t *testing.T) {
	svc := newTestService()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	svc.CreateAllergy(context.Background(), a)

	got, err := svc.GetAllergyByFHIRID(context.Background(), a.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != a.ID {
		t.Errorf("expected ID %v, got %v", a.ID, got.ID)
	}
}

func TestGetAllergyByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetAllergyByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateAllergy(t *testing.T) {
	svc := newTestService()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	svc.CreateAllergy(context.Background(), a)
	resolved := "resolved"
	a.ClinicalStatus = &resolved
	if err := svc.UpdateAllergy(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateAllergy_NotFound(t *testing.T) {
	svc := newTestService()
	a := &AllergyIntolerance{ID: uuid.New()}
	if err := svc.UpdateAllergy(context.Background(), a); err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestDeleteAllergy(t *testing.T) {
	svc := newTestService()
	a := &AllergyIntolerance{PatientID: uuid.New()}
	svc.CreateAllergy(context.Background(), a)
	if err := svc.DeleteAllergy(context.Background(), a.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetAllergy(context.Background(), a.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListAllergiesByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateAllergy(context.Background(), &AllergyIntolerance{PatientID: pid})
	svc.CreateAllergy(context.Background(), &AllergyIntolerance{PatientID: pid})
	svc.CreateAllergy(context.Background(), &AllergyIntolerance{PatientID: uuid.New()})

	items, total, err := svc.ListAllergiesByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 allergies, got %d", total)
	}
}

func TestSearchAllergies(t *testing.T) {
	svc := newTestService()
	svc.CreateAllergy(context.Background(), &AllergyIntolerance{PatientID: uuid.New()})
	items, total, err := svc.SearchAllergies(context.Background(), map[string]string{}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(items) < 1 {
		t.Error("expected items")
	}
}

// =========== Additional Procedure Tests ===========

func TestGetProcedure(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002"}
	svc.CreateProcedure(context.Background(), p)

	got, err := svc.GetProcedure(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("expected ID %v, got %v", p.ID, got.ID)
	}
}

func TestGetProcedure_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetProcedure(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetProcedureByFHIRID(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002"}
	svc.CreateProcedure(context.Background(), p)

	got, err := svc.GetProcedureByFHIRID(context.Background(), p.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("expected ID %v, got %v", p.ID, got.ID)
	}
}

func TestGetProcedureByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetProcedureByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateProcedure(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002"}
	svc.CreateProcedure(context.Background(), p)
	p.Status = "in-progress"
	if err := svc.UpdateProcedure(context.Background(), p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateProcedure_NotFound(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{ID: uuid.New()}
	if err := svc.UpdateProcedure(context.Background(), p); err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestDeleteProcedure(t *testing.T) {
	svc := newTestService()
	p := &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002"}
	svc.CreateProcedure(context.Background(), p)
	if err := svc.DeleteProcedure(context.Background(), p.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetProcedure(context.Background(), p.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListProceduresByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateProcedure(context.Background(), &ProcedureRecord{PatientID: pid, CodeValue: "80146002"})
	svc.CreateProcedure(context.Background(), &ProcedureRecord{PatientID: pid, CodeValue: "44950"})
	svc.CreateProcedure(context.Background(), &ProcedureRecord{PatientID: uuid.New(), CodeValue: "47562"})

	items, total, err := svc.ListProceduresByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 procedures, got %d", total)
	}
}

func TestSearchProcedures(t *testing.T) {
	svc := newTestService()
	svc.CreateProcedure(context.Background(), &ProcedureRecord{PatientID: uuid.New(), CodeValue: "80146002"})
	items, total, err := svc.SearchProcedures(context.Background(), map[string]string{"code": "80146002"}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total < 1 {
		t.Errorf("expected at least 1, got %d", total)
	}
	if len(items) < 1 {
		t.Error("expected items")
	}
}
