package behavioral

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// -- Mock Repositories --

type mockPsychAssessmentRepo struct {
	records map[uuid.UUID]*PsychiatricAssessment
}

func newMockPsychAssessmentRepo() *mockPsychAssessmentRepo {
	return &mockPsychAssessmentRepo{records: make(map[uuid.UUID]*PsychiatricAssessment)}
}

func (m *mockPsychAssessmentRepo) Create(_ context.Context, a *PsychiatricAssessment) error {
	a.ID = uuid.New()
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	m.records[a.ID] = a
	return nil
}
func (m *mockPsychAssessmentRepo) GetByID(_ context.Context, id uuid.UUID) (*PsychiatricAssessment, error) {
	a, ok := m.records[id]
	if !ok { return nil, fmt.Errorf("not found") }
	return a, nil
}
func (m *mockPsychAssessmentRepo) Update(_ context.Context, a *PsychiatricAssessment) error { m.records[a.ID] = a; return nil }
func (m *mockPsychAssessmentRepo) Delete(_ context.Context, id uuid.UUID) error { delete(m.records, id); return nil }
func (m *mockPsychAssessmentRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*PsychiatricAssessment, int, error) {
	var result []*PsychiatricAssessment
	for _, a := range m.records { if a.PatientID == patientID { result = append(result, a) } }
	return result, len(result), nil
}
func (m *mockPsychAssessmentRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*PsychiatricAssessment, int, error) {
	var result []*PsychiatricAssessment
	for _, a := range m.records { result = append(result, a) }
	return result, len(result), nil
}

type mockSafetyPlanRepo struct {
	records map[uuid.UUID]*SafetyPlan
}

func newMockSafetyPlanRepo() *mockSafetyPlanRepo {
	return &mockSafetyPlanRepo{records: make(map[uuid.UUID]*SafetyPlan)}
}

func (m *mockSafetyPlanRepo) Create(_ context.Context, s *SafetyPlan) error {
	s.ID = uuid.New(); s.CreatedAt = time.Now(); s.UpdatedAt = time.Now(); m.records[s.ID] = s; return nil
}
func (m *mockSafetyPlanRepo) GetByID(_ context.Context, id uuid.UUID) (*SafetyPlan, error) {
	s, ok := m.records[id]; if !ok { return nil, fmt.Errorf("not found") }; return s, nil
}
func (m *mockSafetyPlanRepo) Update(_ context.Context, s *SafetyPlan) error { m.records[s.ID] = s; return nil }
func (m *mockSafetyPlanRepo) Delete(_ context.Context, id uuid.UUID) error { delete(m.records, id); return nil }
func (m *mockSafetyPlanRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*SafetyPlan, int, error) {
	var result []*SafetyPlan
	for _, s := range m.records { if s.PatientID == patientID { result = append(result, s) } }
	return result, len(result), nil
}
func (m *mockSafetyPlanRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*SafetyPlan, int, error) {
	var result []*SafetyPlan
	for _, s := range m.records { result = append(result, s) }
	return result, len(result), nil
}

type mockLegalHoldRepo struct {
	records map[uuid.UUID]*LegalHold
}

func newMockLegalHoldRepo() *mockLegalHoldRepo {
	return &mockLegalHoldRepo{records: make(map[uuid.UUID]*LegalHold)}
}

func (m *mockLegalHoldRepo) Create(_ context.Context, h *LegalHold) error {
	h.ID = uuid.New(); h.CreatedAt = time.Now(); h.UpdatedAt = time.Now(); m.records[h.ID] = h; return nil
}
func (m *mockLegalHoldRepo) GetByID(_ context.Context, id uuid.UUID) (*LegalHold, error) {
	h, ok := m.records[id]; if !ok { return nil, fmt.Errorf("not found") }; return h, nil
}
func (m *mockLegalHoldRepo) Update(_ context.Context, h *LegalHold) error { m.records[h.ID] = h; return nil }
func (m *mockLegalHoldRepo) Delete(_ context.Context, id uuid.UUID) error { delete(m.records, id); return nil }
func (m *mockLegalHoldRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*LegalHold, int, error) {
	var result []*LegalHold
	for _, h := range m.records { if h.PatientID == patientID { result = append(result, h) } }
	return result, len(result), nil
}
func (m *mockLegalHoldRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*LegalHold, int, error) {
	var result []*LegalHold
	for _, h := range m.records { result = append(result, h) }
	return result, len(result), nil
}

type mockSeclusionRestraintRepo struct {
	records map[uuid.UUID]*SeclusionRestraintEvent
}

func newMockSeclusionRestraintRepo() *mockSeclusionRestraintRepo {
	return &mockSeclusionRestraintRepo{records: make(map[uuid.UUID]*SeclusionRestraintEvent)}
}

func (m *mockSeclusionRestraintRepo) Create(_ context.Context, e *SeclusionRestraintEvent) error {
	e.ID = uuid.New(); e.CreatedAt = time.Now(); e.UpdatedAt = time.Now(); m.records[e.ID] = e; return nil
}
func (m *mockSeclusionRestraintRepo) GetByID(_ context.Context, id uuid.UUID) (*SeclusionRestraintEvent, error) {
	e, ok := m.records[id]; if !ok { return nil, fmt.Errorf("not found") }; return e, nil
}
func (m *mockSeclusionRestraintRepo) Update(_ context.Context, e *SeclusionRestraintEvent) error { m.records[e.ID] = e; return nil }
func (m *mockSeclusionRestraintRepo) Delete(_ context.Context, id uuid.UUID) error { delete(m.records, id); return nil }
func (m *mockSeclusionRestraintRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*SeclusionRestraintEvent, int, error) {
	var result []*SeclusionRestraintEvent
	for _, e := range m.records { if e.PatientID == patientID { result = append(result, e) } }
	return result, len(result), nil
}
func (m *mockSeclusionRestraintRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*SeclusionRestraintEvent, int, error) {
	var result []*SeclusionRestraintEvent
	for _, e := range m.records { result = append(result, e) }
	return result, len(result), nil
}

type mockGroupTherapyRepo struct {
	records    map[uuid.UUID]*GroupTherapySession
	attendance map[uuid.UUID]*GroupTherapyAttendance
}

func newMockGroupTherapyRepo() *mockGroupTherapyRepo {
	return &mockGroupTherapyRepo{
		records:    make(map[uuid.UUID]*GroupTherapySession),
		attendance: make(map[uuid.UUID]*GroupTherapyAttendance),
	}
}

func (m *mockGroupTherapyRepo) Create(_ context.Context, s *GroupTherapySession) error {
	s.ID = uuid.New(); s.CreatedAt = time.Now(); s.UpdatedAt = time.Now(); m.records[s.ID] = s; return nil
}
func (m *mockGroupTherapyRepo) GetByID(_ context.Context, id uuid.UUID) (*GroupTherapySession, error) {
	s, ok := m.records[id]; if !ok { return nil, fmt.Errorf("not found") }; return s, nil
}
func (m *mockGroupTherapyRepo) Update(_ context.Context, s *GroupTherapySession) error { m.records[s.ID] = s; return nil }
func (m *mockGroupTherapyRepo) Delete(_ context.Context, id uuid.UUID) error { delete(m.records, id); return nil }
func (m *mockGroupTherapyRepo) List(_ context.Context, limit, offset int) ([]*GroupTherapySession, int, error) {
	var result []*GroupTherapySession
	for _, s := range m.records { result = append(result, s) }
	return result, len(result), nil
}
func (m *mockGroupTherapyRepo) Search(_ context.Context, _ map[string]string, limit, offset int) ([]*GroupTherapySession, int, error) {
	return m.List(context.Background(), limit, offset)
}
func (m *mockGroupTherapyRepo) AddAttendance(_ context.Context, a *GroupTherapyAttendance) error {
	a.ID = uuid.New(); m.attendance[a.ID] = a; return nil
}
func (m *mockGroupTherapyRepo) GetAttendance(_ context.Context, sessionID uuid.UUID) ([]*GroupTherapyAttendance, error) {
	var result []*GroupTherapyAttendance
	for _, a := range m.attendance { if a.SessionID == sessionID { result = append(result, a) } }
	return result, nil
}

// -- Tests --

func newTestService() *Service {
	return NewService(newMockPsychAssessmentRepo(), newMockSafetyPlanRepo(), newMockLegalHoldRepo(), newMockSeclusionRestraintRepo(), newMockGroupTherapyRepo())
}

func TestCreatePsychAssessment(t *testing.T) {
	svc := newTestService()
	a := &PsychiatricAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), AssessorID: uuid.New()}
	err := svc.CreatePsychAssessment(context.Background(), a)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if a.ID == uuid.Nil { t.Error("expected ID to be set") }
}

func TestCreatePsychAssessment_PatientRequired(t *testing.T) {
	svc := newTestService()
	a := &PsychiatricAssessment{EncounterID: uuid.New(), AssessorID: uuid.New()}
	err := svc.CreatePsychAssessment(context.Background(), a)
	if err == nil { t.Error("expected error for missing patient_id") }
}

func TestCreatePsychAssessment_EncounterRequired(t *testing.T) {
	svc := newTestService()
	a := &PsychiatricAssessment{PatientID: uuid.New(), AssessorID: uuid.New()}
	err := svc.CreatePsychAssessment(context.Background(), a)
	if err == nil { t.Error("expected error for missing encounter_id") }
}

func TestCreatePsychAssessment_AssessorRequired(t *testing.T) {
	svc := newTestService()
	a := &PsychiatricAssessment{PatientID: uuid.New(), EncounterID: uuid.New()}
	err := svc.CreatePsychAssessment(context.Background(), a)
	if err == nil { t.Error("expected error for missing assessor_id") }
}

func TestGetPsychAssessment(t *testing.T) {
	svc := newTestService()
	a := &PsychiatricAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), AssessorID: uuid.New()}
	svc.CreatePsychAssessment(context.Background(), a)
	fetched, err := svc.GetPsychAssessment(context.Background(), a.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if fetched.PatientID != a.PatientID { t.Error("patient_id mismatch") }
}

func TestDeletePsychAssessment(t *testing.T) {
	svc := newTestService()
	a := &PsychiatricAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), AssessorID: uuid.New()}
	svc.CreatePsychAssessment(context.Background(), a)
	err := svc.DeletePsychAssessment(context.Background(), a.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	_, err = svc.GetPsychAssessment(context.Background(), a.ID)
	if err == nil { t.Error("expected error after deletion") }
}

func TestCreateSafetyPlan(t *testing.T) {
	svc := newTestService()
	sp := &SafetyPlan{PatientID: uuid.New(), CreatedByID: uuid.New()}
	err := svc.CreateSafetyPlan(context.Background(), sp)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if sp.Status != "active" { t.Errorf("expected default status 'active', got %s", sp.Status) }
}

func TestCreateSafetyPlan_PatientRequired(t *testing.T) {
	svc := newTestService()
	sp := &SafetyPlan{CreatedByID: uuid.New()}
	err := svc.CreateSafetyPlan(context.Background(), sp)
	if err == nil { t.Error("expected error for missing patient_id") }
}

func TestCreateSafetyPlan_CreatedByRequired(t *testing.T) {
	svc := newTestService()
	sp := &SafetyPlan{PatientID: uuid.New()}
	err := svc.CreateSafetyPlan(context.Background(), sp)
	if err == nil { t.Error("expected error for missing created_by_id") }
}

func TestCreateSafetyPlan_InvalidStatus(t *testing.T) {
	svc := newTestService()
	sp := &SafetyPlan{PatientID: uuid.New(), CreatedByID: uuid.New(), Status: "bogus"}
	err := svc.CreateSafetyPlan(context.Background(), sp)
	if err == nil { t.Error("expected error for invalid status") }
}

func TestCreateLegalHold(t *testing.T) {
	svc := newTestService()
	h := &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger to self"}
	err := svc.CreateLegalHold(context.Background(), h)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if h.Status != "active" { t.Errorf("expected default status 'active', got %s", h.Status) }
}

func TestCreateLegalHold_PatientRequired(t *testing.T) {
	svc := newTestService()
	h := &LegalHold{InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger"}
	err := svc.CreateLegalHold(context.Background(), h)
	if err == nil { t.Error("expected error for missing patient_id") }
}

func TestCreateLegalHold_HoldTypeRequired(t *testing.T) {
	svc := newTestService()
	h := &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), Reason: "danger"}
	err := svc.CreateLegalHold(context.Background(), h)
	if err == nil { t.Error("expected error for missing hold_type") }
}

func TestCreateLegalHold_ReasonRequired(t *testing.T) {
	svc := newTestService()
	h := &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), HoldType: "5150"}
	err := svc.CreateLegalHold(context.Background(), h)
	if err == nil { t.Error("expected error for missing reason") }
}

func TestCreateSeclusionRestraint(t *testing.T) {
	svc := newTestService()
	e := &SeclusionRestraintEvent{PatientID: uuid.New(), OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"}
	err := svc.CreateSeclusionRestraint(context.Background(), e)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if e.ID == uuid.Nil { t.Error("expected ID to be set") }
}

func TestCreateSeclusionRestraint_PatientRequired(t *testing.T) {
	svc := newTestService()
	e := &SeclusionRestraintEvent{OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"}
	err := svc.CreateSeclusionRestraint(context.Background(), e)
	if err == nil { t.Error("expected error for missing patient_id") }
}

func TestCreateSeclusionRestraint_EventTypeRequired(t *testing.T) {
	svc := newTestService()
	e := &SeclusionRestraintEvent{PatientID: uuid.New(), OrderedByID: uuid.New(), Reason: "agitated"}
	err := svc.CreateSeclusionRestraint(context.Background(), e)
	if err == nil { t.Error("expected error for missing event_type") }
}

func TestCreateSeclusionRestraint_ReasonRequired(t *testing.T) {
	svc := newTestService()
	e := &SeclusionRestraintEvent{PatientID: uuid.New(), OrderedByID: uuid.New(), EventType: "seclusion"}
	err := svc.CreateSeclusionRestraint(context.Background(), e)
	if err == nil { t.Error("expected error for missing reason") }
}

func TestCreateGroupTherapySession(t *testing.T) {
	svc := newTestService()
	gs := &GroupTherapySession{SessionName: "CBT Group", FacilitatorID: uuid.New()}
	err := svc.CreateGroupTherapySession(context.Background(), gs)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if gs.Status != "scheduled" { t.Errorf("expected default status 'scheduled', got %s", gs.Status) }
}

func TestCreateGroupTherapySession_NameRequired(t *testing.T) {
	svc := newTestService()
	gs := &GroupTherapySession{FacilitatorID: uuid.New()}
	err := svc.CreateGroupTherapySession(context.Background(), gs)
	if err == nil { t.Error("expected error for missing session_name") }
}

func TestCreateGroupTherapySession_FacilitatorRequired(t *testing.T) {
	svc := newTestService()
	gs := &GroupTherapySession{SessionName: "CBT Group"}
	err := svc.CreateGroupTherapySession(context.Background(), gs)
	if err == nil { t.Error("expected error for missing facilitator_id") }
}

func TestCreateGroupTherapySession_InvalidStatus(t *testing.T) {
	svc := newTestService()
	gs := &GroupTherapySession{SessionName: "CBT Group", FacilitatorID: uuid.New(), Status: "bogus"}
	err := svc.CreateGroupTherapySession(context.Background(), gs)
	if err == nil { t.Error("expected error for invalid status") }
}

func TestDeleteGroupTherapySession(t *testing.T) {
	svc := newTestService()
	gs := &GroupTherapySession{SessionName: "CBT Group", FacilitatorID: uuid.New()}
	svc.CreateGroupTherapySession(context.Background(), gs)
	err := svc.DeleteGroupTherapySession(context.Background(), gs.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	_, err = svc.GetGroupTherapySession(context.Background(), gs.ID)
	if err == nil { t.Error("expected error after deletion") }
}

func TestAddGroupTherapyAttendance(t *testing.T) {
	svc := newTestService()
	a := &GroupTherapyAttendance{SessionID: uuid.New(), PatientID: uuid.New()}
	err := svc.AddGroupTherapyAttendance(context.Background(), a)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if a.ID == uuid.Nil { t.Error("expected ID to be set") }
}

func TestAddGroupTherapyAttendance_SessionRequired(t *testing.T) {
	svc := newTestService()
	a := &GroupTherapyAttendance{PatientID: uuid.New()}
	err := svc.AddGroupTherapyAttendance(context.Background(), a)
	if err == nil { t.Error("expected error for missing session_id") }
}

func TestAddGroupTherapyAttendance_PatientRequired(t *testing.T) {
	svc := newTestService()
	a := &GroupTherapyAttendance{SessionID: uuid.New()}
	err := svc.AddGroupTherapyAttendance(context.Background(), a)
	if err == nil { t.Error("expected error for missing patient_id") }
}

// -- Additional PsychAssessment Tests --

func TestUpdatePsychAssessment(t *testing.T) {
	svc := newTestService()
	a := &PsychiatricAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), AssessorID: uuid.New()}
	svc.CreatePsychAssessment(context.Background(), a)
	err := svc.UpdatePsychAssessment(context.Background(), a)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestListPsychAssessmentsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreatePsychAssessment(context.Background(), &PsychiatricAssessment{PatientID: patientID, EncounterID: uuid.New(), AssessorID: uuid.New()})
	items, total, err := svc.ListPsychAssessmentsByPatient(context.Background(), patientID, 20, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total != 1 { t.Errorf("expected 1, got %d", total) }
	if len(items) != 1 { t.Errorf("expected 1 item, got %d", len(items)) }
}

func TestSearchPsychAssessments(t *testing.T) {
	svc := newTestService()
	svc.CreatePsychAssessment(context.Background(), &PsychiatricAssessment{PatientID: uuid.New(), EncounterID: uuid.New(), AssessorID: uuid.New()})
	items, total, err := svc.SearchPsychAssessments(context.Background(), map[string]string{}, 20, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total < 1 { t.Errorf("expected at least 1, got %d", total) }
	if len(items) < 1 { t.Error("expected items") }
}

// -- Additional SafetyPlan Tests --

func TestGetSafetyPlan(t *testing.T) {
	svc := newTestService()
	sp := &SafetyPlan{PatientID: uuid.New(), CreatedByID: uuid.New()}
	svc.CreateSafetyPlan(context.Background(), sp)
	fetched, err := svc.GetSafetyPlan(context.Background(), sp.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if fetched.ID != sp.ID { t.Error("unexpected ID mismatch") }
}

func TestGetSafetyPlan_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetSafetyPlan(context.Background(), uuid.New())
	if err == nil { t.Error("expected error for not found") }
}

func TestUpdateSafetyPlan(t *testing.T) {
	svc := newTestService()
	sp := &SafetyPlan{PatientID: uuid.New(), CreatedByID: uuid.New()}
	svc.CreateSafetyPlan(context.Background(), sp)
	sp.Status = "superseded"
	err := svc.UpdateSafetyPlan(context.Background(), sp)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestUpdateSafetyPlan_InvalidStatus(t *testing.T) {
	svc := newTestService()
	sp := &SafetyPlan{PatientID: uuid.New(), CreatedByID: uuid.New()}
	svc.CreateSafetyPlan(context.Background(), sp)
	sp.Status = "bogus"
	err := svc.UpdateSafetyPlan(context.Background(), sp)
	if err == nil { t.Error("expected error for invalid status") }
}

func TestDeleteSafetyPlan(t *testing.T) {
	svc := newTestService()
	sp := &SafetyPlan{PatientID: uuid.New(), CreatedByID: uuid.New()}
	svc.CreateSafetyPlan(context.Background(), sp)
	err := svc.DeleteSafetyPlan(context.Background(), sp.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	_, err = svc.GetSafetyPlan(context.Background(), sp.ID)
	if err == nil { t.Error("expected error after deletion") }
}

func TestListSafetyPlansByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateSafetyPlan(context.Background(), &SafetyPlan{PatientID: patientID, CreatedByID: uuid.New()})
	items, total, err := svc.ListSafetyPlansByPatient(context.Background(), patientID, 20, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total != 1 { t.Errorf("expected 1, got %d", total) }
	if len(items) != 1 { t.Errorf("expected 1 item, got %d", len(items)) }
}

func TestSearchSafetyPlans(t *testing.T) {
	svc := newTestService()
	svc.CreateSafetyPlan(context.Background(), &SafetyPlan{PatientID: uuid.New(), CreatedByID: uuid.New()})
	items, total, err := svc.SearchSafetyPlans(context.Background(), map[string]string{}, 20, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total < 1 { t.Errorf("expected at least 1, got %d", total) }
	if len(items) < 1 { t.Error("expected items") }
}

// -- Additional LegalHold Tests --

func TestGetLegalHold(t *testing.T) {
	svc := newTestService()
	h := &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger"}
	svc.CreateLegalHold(context.Background(), h)
	fetched, err := svc.GetLegalHold(context.Background(), h.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if fetched.ID != h.ID { t.Error("unexpected ID mismatch") }
}

func TestGetLegalHold_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetLegalHold(context.Background(), uuid.New())
	if err == nil { t.Error("expected error for not found") }
}

func TestUpdateLegalHold(t *testing.T) {
	svc := newTestService()
	h := &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger"}
	svc.CreateLegalHold(context.Background(), h)
	h.Status = "expired"
	err := svc.UpdateLegalHold(context.Background(), h)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestUpdateLegalHold_InvalidStatus(t *testing.T) {
	svc := newTestService()
	h := &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger"}
	svc.CreateLegalHold(context.Background(), h)
	h.Status = "bogus"
	err := svc.UpdateLegalHold(context.Background(), h)
	if err == nil { t.Error("expected error for invalid status") }
}

func TestDeleteLegalHold(t *testing.T) {
	svc := newTestService()
	h := &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger"}
	svc.CreateLegalHold(context.Background(), h)
	err := svc.DeleteLegalHold(context.Background(), h.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	_, err = svc.GetLegalHold(context.Background(), h.ID)
	if err == nil { t.Error("expected error after deletion") }
}

func TestListLegalHoldsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateLegalHold(context.Background(), &LegalHold{PatientID: patientID, InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger"})
	items, total, err := svc.ListLegalHoldsByPatient(context.Background(), patientID, 20, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total != 1 { t.Errorf("expected 1, got %d", total) }
	if len(items) != 1 { t.Errorf("expected 1 item, got %d", len(items)) }
}

func TestSearchLegalHolds(t *testing.T) {
	svc := newTestService()
	svc.CreateLegalHold(context.Background(), &LegalHold{PatientID: uuid.New(), InitiatedByID: uuid.New(), HoldType: "5150", Reason: "danger"})
	items, total, err := svc.SearchLegalHolds(context.Background(), map[string]string{}, 20, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total < 1 { t.Errorf("expected at least 1, got %d", total) }
	if len(items) < 1 { t.Error("expected items") }
}

// -- Additional SeclusionRestraint Tests --

func TestGetSeclusionRestraint(t *testing.T) {
	svc := newTestService()
	e := &SeclusionRestraintEvent{PatientID: uuid.New(), OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"}
	svc.CreateSeclusionRestraint(context.Background(), e)
	fetched, err := svc.GetSeclusionRestraint(context.Background(), e.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if fetched.ID != e.ID { t.Error("unexpected ID mismatch") }
}

func TestGetSeclusionRestraint_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetSeclusionRestraint(context.Background(), uuid.New())
	if err == nil { t.Error("expected error for not found") }
}

func TestUpdateSeclusionRestraint(t *testing.T) {
	svc := newTestService()
	e := &SeclusionRestraintEvent{PatientID: uuid.New(), OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"}
	svc.CreateSeclusionRestraint(context.Background(), e)
	err := svc.UpdateSeclusionRestraint(context.Background(), e)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestDeleteSeclusionRestraint(t *testing.T) {
	svc := newTestService()
	e := &SeclusionRestraintEvent{PatientID: uuid.New(), OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"}
	svc.CreateSeclusionRestraint(context.Background(), e)
	err := svc.DeleteSeclusionRestraint(context.Background(), e.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	_, err = svc.GetSeclusionRestraint(context.Background(), e.ID)
	if err == nil { t.Error("expected error after deletion") }
}

func TestListSeclusionRestraintsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateSeclusionRestraint(context.Background(), &SeclusionRestraintEvent{PatientID: patientID, OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"})
	items, total, err := svc.ListSeclusionRestraintsByPatient(context.Background(), patientID, 20, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total != 1 { t.Errorf("expected 1, got %d", total) }
	if len(items) != 1 { t.Errorf("expected 1 item, got %d", len(items)) }
}

func TestSearchSeclusionRestraints(t *testing.T) {
	svc := newTestService()
	svc.CreateSeclusionRestraint(context.Background(), &SeclusionRestraintEvent{PatientID: uuid.New(), OrderedByID: uuid.New(), EventType: "seclusion", Reason: "agitated"})
	items, total, err := svc.SearchSeclusionRestraints(context.Background(), map[string]string{}, 20, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total < 1 { t.Errorf("expected at least 1, got %d", total) }
	if len(items) < 1 { t.Error("expected items") }
}

// -- Additional GroupTherapy Tests --

func TestGetGroupTherapySession(t *testing.T) {
	svc := newTestService()
	gs := &GroupTherapySession{SessionName: "CBT Group", FacilitatorID: uuid.New()}
	svc.CreateGroupTherapySession(context.Background(), gs)
	fetched, err := svc.GetGroupTherapySession(context.Background(), gs.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if fetched.ID != gs.ID { t.Error("unexpected ID mismatch") }
}

func TestUpdateGroupTherapySession(t *testing.T) {
	svc := newTestService()
	gs := &GroupTherapySession{SessionName: "CBT Group", FacilitatorID: uuid.New()}
	svc.CreateGroupTherapySession(context.Background(), gs)
	gs.Status = "completed"
	err := svc.UpdateGroupTherapySession(context.Background(), gs)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
}

func TestUpdateGroupTherapySession_InvalidStatus(t *testing.T) {
	svc := newTestService()
	gs := &GroupTherapySession{SessionName: "CBT Group", FacilitatorID: uuid.New()}
	svc.CreateGroupTherapySession(context.Background(), gs)
	gs.Status = "bogus"
	err := svc.UpdateGroupTherapySession(context.Background(), gs)
	if err == nil { t.Error("expected error for invalid status") }
}

func TestListGroupTherapySessions(t *testing.T) {
	svc := newTestService()
	svc.CreateGroupTherapySession(context.Background(), &GroupTherapySession{SessionName: "Group A", FacilitatorID: uuid.New()})
	svc.CreateGroupTherapySession(context.Background(), &GroupTherapySession{SessionName: "Group B", FacilitatorID: uuid.New()})
	items, total, err := svc.ListGroupTherapySessions(context.Background(), 20, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total != 2 { t.Errorf("expected 2, got %d", total) }
	if len(items) != 2 { t.Errorf("expected 2, got %d", len(items)) }
}

func TestSearchGroupTherapySessions(t *testing.T) {
	svc := newTestService()
	svc.CreateGroupTherapySession(context.Background(), &GroupTherapySession{SessionName: "CBT", FacilitatorID: uuid.New()})
	items, total, err := svc.SearchGroupTherapySessions(context.Background(), map[string]string{}, 20, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total < 1 { t.Errorf("expected at least 1, got %d", total) }
	if len(items) < 1 { t.Error("expected items") }
}

func TestGetGroupTherapyAttendance(t *testing.T) {
	svc := newTestService()
	sessionID := uuid.New()
	svc.AddGroupTherapyAttendance(context.Background(), &GroupTherapyAttendance{SessionID: sessionID, PatientID: uuid.New()})
	attendance, err := svc.GetGroupTherapyAttendance(context.Background(), sessionID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if len(attendance) != 1 { t.Errorf("expected 1 attendance, got %d", len(attendance)) }
}
