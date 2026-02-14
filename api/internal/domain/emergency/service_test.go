package emergency

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockTriageRepo struct {
	records map[uuid.UUID]*TriageRecord
}

func newMockTriageRepo() *mockTriageRepo {
	return &mockTriageRepo{records: make(map[uuid.UUID]*TriageRecord)}
}

func (m *mockTriageRepo) Create(_ context.Context, t *TriageRecord) error {
	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	m.records[t.ID] = t
	return nil
}

func (m *mockTriageRepo) GetByID(_ context.Context, id uuid.UUID) (*TriageRecord, error) {
	t, ok := m.records[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

func (m *mockTriageRepo) Update(_ context.Context, t *TriageRecord) error {
	m.records[t.ID] = t
	return nil
}

func (m *mockTriageRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.records, id)
	return nil
}

func (m *mockTriageRepo) List(_ context.Context, limit, offset int) ([]*TriageRecord, int, error) {
	var result []*TriageRecord
	for _, t := range m.records {
		result = append(result, t)
	}
	return result, len(result), nil
}

func (m *mockTriageRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*TriageRecord, int, error) {
	var result []*TriageRecord
	for _, t := range m.records {
		if t.PatientID == patientID {
			result = append(result, t)
		}
	}
	return result, len(result), nil
}

func (m *mockTriageRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*TriageRecord, int, error) {
	return m.List(context.Background(), limit, offset)
}

type mockEDTrackingRepo struct {
	trackings map[uuid.UUID]*EDTracking
	history   map[uuid.UUID]*EDStatusHistory
}

func newMockEDTrackingRepo() *mockEDTrackingRepo {
	return &mockEDTrackingRepo{
		trackings: make(map[uuid.UUID]*EDTracking),
		history:   make(map[uuid.UUID]*EDStatusHistory),
	}
}

func (m *mockEDTrackingRepo) Create(_ context.Context, t *EDTracking) error {
	t.ID = uuid.New()
	m.trackings[t.ID] = t
	return nil
}

func (m *mockEDTrackingRepo) GetByID(_ context.Context, id uuid.UUID) (*EDTracking, error) {
	t, ok := m.trackings[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

func (m *mockEDTrackingRepo) Update(_ context.Context, t *EDTracking) error {
	m.trackings[t.ID] = t
	return nil
}

func (m *mockEDTrackingRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.trackings, id)
	return nil
}

func (m *mockEDTrackingRepo) List(_ context.Context, limit, offset int) ([]*EDTracking, int, error) {
	var result []*EDTracking
	for _, t := range m.trackings {
		result = append(result, t)
	}
	return result, len(result), nil
}

func (m *mockEDTrackingRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*EDTracking, int, error) {
	var result []*EDTracking
	for _, t := range m.trackings {
		if t.PatientID == patientID {
			result = append(result, t)
		}
	}
	return result, len(result), nil
}

func (m *mockEDTrackingRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*EDTracking, int, error) {
	return m.List(context.Background(), limit, offset)
}

func (m *mockEDTrackingRepo) AddStatusHistory(_ context.Context, h *EDStatusHistory) error {
	h.ID = uuid.New()
	m.history[h.ID] = h
	return nil
}

func (m *mockEDTrackingRepo) GetStatusHistory(_ context.Context, trackingID uuid.UUID) ([]*EDStatusHistory, error) {
	var result []*EDStatusHistory
	for _, h := range m.history {
		if h.EDTrackingID == trackingID {
			result = append(result, h)
		}
	}
	return result, nil
}

type mockTraumaRepo struct {
	activations map[uuid.UUID]*TraumaActivation
}

func newMockTraumaRepo() *mockTraumaRepo {
	return &mockTraumaRepo{activations: make(map[uuid.UUID]*TraumaActivation)}
}

func (m *mockTraumaRepo) Create(_ context.Context, t *TraumaActivation) error {
	t.ID = uuid.New()
	m.activations[t.ID] = t
	return nil
}

func (m *mockTraumaRepo) GetByID(_ context.Context, id uuid.UUID) (*TraumaActivation, error) {
	t, ok := m.activations[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

func (m *mockTraumaRepo) Update(_ context.Context, t *TraumaActivation) error {
	m.activations[t.ID] = t
	return nil
}

func (m *mockTraumaRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.activations, id)
	return nil
}

func (m *mockTraumaRepo) List(_ context.Context, limit, offset int) ([]*TraumaActivation, int, error) {
	var result []*TraumaActivation
	for _, t := range m.activations {
		result = append(result, t)
	}
	return result, len(result), nil
}

func (m *mockTraumaRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*TraumaActivation, int, error) {
	var result []*TraumaActivation
	for _, t := range m.activations {
		if t.PatientID == patientID {
			result = append(result, t)
		}
	}
	return result, len(result), nil
}

func (m *mockTraumaRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*TraumaActivation, int, error) {
	return m.List(context.Background(), limit, offset)
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockTriageRepo(), newMockEDTrackingRepo(), newMockTraumaRepo())
}

func TestCreateTriageRecord(t *testing.T) {
	svc := newTestService()
	tr := &TriageRecord{PatientID: uuid.New(), EncounterID: uuid.New(), TriageNurseID: uuid.New(), ChiefComplaint: "chest pain"}
	err := svc.CreateTriageRecord(context.Background(), tr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if tr.TriageTime == nil {
		t.Error("expected triage_time to be defaulted")
	}
}

func TestCreateTriageRecord_PatientRequired(t *testing.T) {
	svc := newTestService()
	tr := &TriageRecord{EncounterID: uuid.New(), TriageNurseID: uuid.New(), ChiefComplaint: "pain"}
	err := svc.CreateTriageRecord(context.Background(), tr)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateTriageRecord_ChiefComplaintRequired(t *testing.T) {
	svc := newTestService()
	tr := &TriageRecord{PatientID: uuid.New(), EncounterID: uuid.New(), TriageNurseID: uuid.New()}
	err := svc.CreateTriageRecord(context.Background(), tr)
	if err == nil {
		t.Error("expected error for missing chief_complaint")
	}
}

func TestGetTriageRecord(t *testing.T) {
	svc := newTestService()
	tr := &TriageRecord{PatientID: uuid.New(), EncounterID: uuid.New(), TriageNurseID: uuid.New(), ChiefComplaint: "pain"}
	svc.CreateTriageRecord(context.Background(), tr)

	fetched, err := svc.GetTriageRecord(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched.ChiefComplaint != "pain" {
		t.Errorf("expected chief_complaint 'pain', got %s", fetched.ChiefComplaint)
	}
}

func TestDeleteTriageRecord(t *testing.T) {
	svc := newTestService()
	tr := &TriageRecord{PatientID: uuid.New(), EncounterID: uuid.New(), TriageNurseID: uuid.New(), ChiefComplaint: "pain"}
	svc.CreateTriageRecord(context.Background(), tr)
	err := svc.DeleteTriageRecord(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GetTriageRecord(context.Background(), tr.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestCreateEDTracking(t *testing.T) {
	svc := newTestService()
	ed := &EDTracking{PatientID: uuid.New(), EncounterID: uuid.New()}
	err := svc.CreateEDTracking(context.Background(), ed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ed.CurrentStatus != "waiting" {
		t.Errorf("expected default status 'waiting', got %s", ed.CurrentStatus)
	}
}

func TestCreateEDTracking_PatientRequired(t *testing.T) {
	svc := newTestService()
	ed := &EDTracking{EncounterID: uuid.New()}
	err := svc.CreateEDTracking(context.Background(), ed)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestAddEDStatusHistory(t *testing.T) {
	svc := newTestService()
	trackingID := uuid.New()
	h := &EDStatusHistory{EDTrackingID: trackingID, Status: "triaged"}
	err := svc.AddEDStatusHistory(context.Background(), h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.ChangedAt.IsZero() {
		t.Error("expected changed_at to be defaulted")
	}
}

func TestAddEDStatusHistory_StatusRequired(t *testing.T) {
	svc := newTestService()
	h := &EDStatusHistory{EDTrackingID: uuid.New()}
	err := svc.AddEDStatusHistory(context.Background(), h)
	if err == nil {
		t.Error("expected error for missing status")
	}
}

func TestCreateTraumaActivation(t *testing.T) {
	svc := newTestService()
	ta := &TraumaActivation{PatientID: uuid.New(), ActivationLevel: "level-1"}
	err := svc.CreateTraumaActivation(context.Background(), ta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ta.ActivationTime.IsZero() {
		t.Error("expected activation_time to be defaulted")
	}
}

func TestCreateTraumaActivation_PatientRequired(t *testing.T) {
	svc := newTestService()
	ta := &TraumaActivation{ActivationLevel: "level-1"}
	err := svc.CreateTraumaActivation(context.Background(), ta)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestCreateTraumaActivation_LevelRequired(t *testing.T) {
	svc := newTestService()
	ta := &TraumaActivation{PatientID: uuid.New()}
	err := svc.CreateTraumaActivation(context.Background(), ta)
	if err == nil {
		t.Error("expected error for missing activation_level")
	}
}
