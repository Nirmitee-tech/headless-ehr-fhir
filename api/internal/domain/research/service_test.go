package research

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// ── Mock Repositories ──

type mockStudyRepo struct {
	data map[uuid.UUID]*ResearchStudy
	arms map[uuid.UUID][]*ResearchArm
}

func (m *mockStudyRepo) Create(_ context.Context, s *ResearchStudy) error {
	s.ID = uuid.New()
	if s.FHIRID == "" {
		s.FHIRID = "fhir-" + s.ID.String()
	}
	m.data[s.ID] = s
	return nil
}
func (m *mockStudyRepo) GetByID(_ context.Context, id uuid.UUID) (*ResearchStudy, error) {
	if s, ok := m.data[id]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockStudyRepo) GetByFHIRID(_ context.Context, fhirID string) (*ResearchStudy, error) {
	for _, s := range m.data {
		if s.FHIRID == fhirID {
			return s, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockStudyRepo) Update(_ context.Context, s *ResearchStudy) error {
	if _, ok := m.data[s.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[s.ID] = s
	return nil
}
func (m *mockStudyRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockStudyRepo) List(_ context.Context, limit, offset int) ([]*ResearchStudy, int, error) {
	var out []*ResearchStudy
	for _, s := range m.data {
		out = append(out, s)
	}
	return out, len(out), nil
}
func (m *mockStudyRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*ResearchStudy, int, error) {
	var out []*ResearchStudy
	for _, s := range m.data {
		out = append(out, s)
	}
	return out, len(out), nil
}
func (m *mockStudyRepo) AddArm(_ context.Context, a *ResearchArm) error {
	a.ID = uuid.New()
	m.arms[a.StudyID] = append(m.arms[a.StudyID], a)
	return nil
}
func (m *mockStudyRepo) GetArms(_ context.Context, studyID uuid.UUID) ([]*ResearchArm, error) {
	return m.arms[studyID], nil
}

type mockEnrollmentRepo struct {
	data map[uuid.UUID]*ResearchEnrollment
}

func (m *mockEnrollmentRepo) Create(_ context.Context, e *ResearchEnrollment) error {
	e.ID = uuid.New()
	m.data[e.ID] = e
	return nil
}
func (m *mockEnrollmentRepo) GetByID(_ context.Context, id uuid.UUID) (*ResearchEnrollment, error) {
	if e, ok := m.data[id]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockEnrollmentRepo) Update(_ context.Context, e *ResearchEnrollment) error {
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
func (m *mockEnrollmentRepo) ListByStudy(_ context.Context, studyID uuid.UUID, limit, offset int) ([]*ResearchEnrollment, int, error) {
	var out []*ResearchEnrollment
	for _, e := range m.data {
		if e.StudyID == studyID {
			out = append(out, e)
		}
	}
	return out, len(out), nil
}
func (m *mockEnrollmentRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*ResearchEnrollment, int, error) {
	var out []*ResearchEnrollment
	for _, e := range m.data {
		if e.PatientID == patientID {
			out = append(out, e)
		}
	}
	return out, len(out), nil
}

type mockAdverseEventRepo struct {
	data map[uuid.UUID]*ResearchAdverseEvent
}

func (m *mockAdverseEventRepo) Create(_ context.Context, ae *ResearchAdverseEvent) error {
	ae.ID = uuid.New()
	m.data[ae.ID] = ae
	return nil
}
func (m *mockAdverseEventRepo) GetByID(_ context.Context, id uuid.UUID) (*ResearchAdverseEvent, error) {
	if ae, ok := m.data[id]; ok {
		return ae, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockAdverseEventRepo) Update(_ context.Context, ae *ResearchAdverseEvent) error {
	if _, ok := m.data[ae.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[ae.ID] = ae
	return nil
}
func (m *mockAdverseEventRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockAdverseEventRepo) ListByEnrollment(_ context.Context, enrollmentID uuid.UUID, limit, offset int) ([]*ResearchAdverseEvent, int, error) {
	var out []*ResearchAdverseEvent
	for _, ae := range m.data {
		if ae.EnrollmentID == enrollmentID {
			out = append(out, ae)
		}
	}
	return out, len(out), nil
}

type mockDeviationRepo struct {
	data map[uuid.UUID]*ResearchProtocolDeviation
}

func (m *mockDeviationRepo) Create(_ context.Context, d *ResearchProtocolDeviation) error {
	d.ID = uuid.New()
	m.data[d.ID] = d
	return nil
}
func (m *mockDeviationRepo) GetByID(_ context.Context, id uuid.UUID) (*ResearchProtocolDeviation, error) {
	if d, ok := m.data[id]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockDeviationRepo) Update(_ context.Context, d *ResearchProtocolDeviation) error {
	if _, ok := m.data[d.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.data[d.ID] = d
	return nil
}
func (m *mockDeviationRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.data, id)
	return nil
}
func (m *mockDeviationRepo) ListByEnrollment(_ context.Context, enrollmentID uuid.UUID, limit, offset int) ([]*ResearchProtocolDeviation, int, error) {
	var out []*ResearchProtocolDeviation
	for _, d := range m.data {
		if d.EnrollmentID == enrollmentID {
			out = append(out, d)
		}
	}
	return out, len(out), nil
}

// ── Helper ──

func newTestService() *Service {
	return NewService(
		&mockStudyRepo{data: make(map[uuid.UUID]*ResearchStudy), arms: make(map[uuid.UUID][]*ResearchArm)},
		&mockEnrollmentRepo{data: make(map[uuid.UUID]*ResearchEnrollment)},
		&mockAdverseEventRepo{data: make(map[uuid.UUID]*ResearchAdverseEvent)},
		&mockDeviationRepo{data: make(map[uuid.UUID]*ResearchProtocolDeviation)},
	)
}

// ── Study Tests ──

func TestService_CreateStudy(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{ProtocolNumber: "PROTO-001", Title: "Test Study"}
	if err := svc.CreateStudy(nil, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if s.Status != "in-review" {
		t.Errorf("expected default status 'in-review', got %s", s.Status)
	}
}

func TestService_CreateStudy_MissingProtocolNumber(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{Title: "Test Study"}
	if err := svc.CreateStudy(nil, s); err == nil {
		t.Error("expected error for missing protocol_number")
	}
}

func TestService_CreateStudy_MissingTitle(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{ProtocolNumber: "PROTO-001"}
	if err := svc.CreateStudy(nil, s); err == nil {
		t.Error("expected error for missing title")
	}
}

func TestService_CreateStudy_InvalidStatus(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{ProtocolNumber: "PROTO-001", Title: "Test", Status: "bogus"}
	if err := svc.CreateStudy(nil, s); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_CreateStudy_ValidStatuses(t *testing.T) {
	statuses := []string{
		"in-review", "approved", "active-recruiting", "active-not-recruiting",
		"temporarily-closed", "closed", "completed", "withdrawn", "suspended",
	}
	for _, status := range statuses {
		svc := newTestService()
		s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T", Status: status}
		if err := svc.CreateStudy(nil, s); err != nil {
			t.Errorf("status %q should be valid, got error: %v", status, err)
		}
	}
}

func TestService_GetStudy(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	svc.CreateStudy(nil, s)
	got, err := svc.GetStudy(nil, s.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "T" {
		t.Errorf("expected title T, got %s", got.Title)
	}
}

func TestService_GetStudy_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetStudy(nil, uuid.New()); err == nil {
		t.Error("expected error for not found")
	}
}

func TestService_GetStudyByFHIRID(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	svc.CreateStudy(nil, s)
	got, err := svc.GetStudyByFHIRID(nil, s.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != s.ID {
		t.Errorf("expected same ID")
	}
}

func TestService_UpdateStudy(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	svc.CreateStudy(nil, s)
	s.Status = "active-recruiting"
	if err := svc.UpdateStudy(nil, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdateStudy_InvalidStatus(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	svc.CreateStudy(nil, s)
	s.Status = "invalid-status"
	if err := svc.UpdateStudy(nil, s); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_DeleteStudy(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	svc.CreateStudy(nil, s)
	if err := svc.DeleteStudy(nil, s.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.GetStudy(nil, s.ID); err == nil {
		t.Error("expected not found after delete")
	}
}

func TestService_ListStudies(t *testing.T) {
	svc := newTestService()
	svc.CreateStudy(nil, &ResearchStudy{ProtocolNumber: "P-1", Title: "T1"})
	svc.CreateStudy(nil, &ResearchStudy{ProtocolNumber: "P-2", Title: "T2"})
	items, total, err := svc.ListStudies(nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestService_SearchStudies(t *testing.T) {
	svc := newTestService()
	svc.CreateStudy(nil, &ResearchStudy{ProtocolNumber: "P-1", Title: "T1"})
	items, total, err := svc.SearchStudies(nil, map[string]string{"status": "in-review"}, 10, 0)
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

// ── Arm Tests ──

func TestService_AddStudyArm(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	svc.CreateStudy(nil, s)
	arm := &ResearchArm{StudyID: s.ID, Name: "Treatment A"}
	if err := svc.AddStudyArm(nil, arm); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if arm.ID == uuid.Nil {
		t.Error("expected arm ID to be set")
	}
}

func TestService_AddStudyArm_MissingStudyID(t *testing.T) {
	svc := newTestService()
	arm := &ResearchArm{Name: "Treatment A"}
	if err := svc.AddStudyArm(nil, arm); err == nil {
		t.Error("expected error for missing study_id")
	}
}

func TestService_AddStudyArm_MissingName(t *testing.T) {
	svc := newTestService()
	arm := &ResearchArm{StudyID: uuid.New()}
	if err := svc.AddStudyArm(nil, arm); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestService_GetStudyArms(t *testing.T) {
	svc := newTestService()
	s := &ResearchStudy{ProtocolNumber: "P-1", Title: "T"}
	svc.CreateStudy(nil, s)
	svc.AddStudyArm(nil, &ResearchArm{StudyID: s.ID, Name: "Arm A"})
	svc.AddStudyArm(nil, &ResearchArm{StudyID: s.ID, Name: "Arm B"})
	arms, err := svc.GetStudyArms(nil, s.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arms) != 2 {
		t.Errorf("expected 2 arms, got %d", len(arms))
	}
}

// ── Enrollment Tests ──

func TestService_CreateEnrollment(t *testing.T) {
	svc := newTestService()
	e := &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New()}
	if err := svc.CreateEnrollment(nil, e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Status != "pre-screening" {
		t.Errorf("expected default status 'pre-screening', got %s", e.Status)
	}
}

func TestService_CreateEnrollment_MissingStudyID(t *testing.T) {
	svc := newTestService()
	e := &ResearchEnrollment{PatientID: uuid.New()}
	if err := svc.CreateEnrollment(nil, e); err == nil {
		t.Error("expected error for missing study_id")
	}
}

func TestService_CreateEnrollment_MissingPatientID(t *testing.T) {
	svc := newTestService()
	e := &ResearchEnrollment{StudyID: uuid.New()}
	if err := svc.CreateEnrollment(nil, e); err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestService_CreateEnrollment_InvalidStatus(t *testing.T) {
	svc := newTestService()
	e := &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New(), Status: "bogus"}
	if err := svc.CreateEnrollment(nil, e); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_CreateEnrollment_ValidStatuses(t *testing.T) {
	statuses := []string{
		"pre-screening", "screening", "screen-fail", "enrolled", "active",
		"on-study-treatment", "follow-up", "completed", "early-termination",
		"withdrawn", "lost-to-followup", "deceased",
	}
	for _, status := range statuses {
		svc := newTestService()
		e := &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New(), Status: status}
		if err := svc.CreateEnrollment(nil, e); err != nil {
			t.Errorf("status %q should be valid, got error: %v", status, err)
		}
	}
}

func TestService_GetEnrollment(t *testing.T) {
	svc := newTestService()
	e := &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New()}
	svc.CreateEnrollment(nil, e)
	got, err := svc.GetEnrollment(nil, e.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != e.ID {
		t.Error("ID mismatch")
	}
}

func TestService_UpdateEnrollment(t *testing.T) {
	svc := newTestService()
	e := &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New()}
	svc.CreateEnrollment(nil, e)
	e.Status = "enrolled"
	if err := svc.UpdateEnrollment(nil, e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdateEnrollment_InvalidStatus(t *testing.T) {
	svc := newTestService()
	e := &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New()}
	svc.CreateEnrollment(nil, e)
	e.Status = "bad"
	if err := svc.UpdateEnrollment(nil, e); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_DeleteEnrollment(t *testing.T) {
	svc := newTestService()
	e := &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New()}
	svc.CreateEnrollment(nil, e)
	if err := svc.DeleteEnrollment(nil, e.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListEnrollmentsByStudy(t *testing.T) {
	svc := newTestService()
	studyID := uuid.New()
	svc.CreateEnrollment(nil, &ResearchEnrollment{StudyID: studyID, PatientID: uuid.New()})
	svc.CreateEnrollment(nil, &ResearchEnrollment{StudyID: studyID, PatientID: uuid.New()})
	svc.CreateEnrollment(nil, &ResearchEnrollment{StudyID: uuid.New(), PatientID: uuid.New()})
	items, total, err := svc.ListEnrollmentsByStudy(nil, studyID, 10, 0)
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

func TestService_ListEnrollmentsByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	svc.CreateEnrollment(nil, &ResearchEnrollment{StudyID: uuid.New(), PatientID: patientID})
	items, total, err := svc.ListEnrollmentsByPatient(nil, patientID, 10, 0)
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

// ── Adverse Event Tests ──

func TestService_CreateAdverseEvent(t *testing.T) {
	svc := newTestService()
	ae := &ResearchAdverseEvent{EnrollmentID: uuid.New(), Description: "Nausea"}
	if err := svc.CreateAdverseEvent(nil, ae); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ae.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestService_CreateAdverseEvent_MissingEnrollmentID(t *testing.T) {
	svc := newTestService()
	ae := &ResearchAdverseEvent{Description: "Nausea"}
	if err := svc.CreateAdverseEvent(nil, ae); err == nil {
		t.Error("expected error for missing enrollment_id")
	}
}

func TestService_CreateAdverseEvent_MissingDescription(t *testing.T) {
	svc := newTestService()
	ae := &ResearchAdverseEvent{EnrollmentID: uuid.New()}
	if err := svc.CreateAdverseEvent(nil, ae); err == nil {
		t.Error("expected error for missing description")
	}
}

func TestService_GetAdverseEvent(t *testing.T) {
	svc := newTestService()
	ae := &ResearchAdverseEvent{EnrollmentID: uuid.New(), Description: "Nausea"}
	svc.CreateAdverseEvent(nil, ae)
	got, err := svc.GetAdverseEvent(nil, ae.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Description != "Nausea" {
		t.Errorf("expected 'Nausea', got %s", got.Description)
	}
}

func TestService_UpdateAdverseEvent(t *testing.T) {
	svc := newTestService()
	ae := &ResearchAdverseEvent{EnrollmentID: uuid.New(), Description: "Nausea"}
	svc.CreateAdverseEvent(nil, ae)
	ae.Description = "Severe Nausea"
	if err := svc.UpdateAdverseEvent(nil, ae); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_DeleteAdverseEvent(t *testing.T) {
	svc := newTestService()
	ae := &ResearchAdverseEvent{EnrollmentID: uuid.New(), Description: "Nausea"}
	svc.CreateAdverseEvent(nil, ae)
	if err := svc.DeleteAdverseEvent(nil, ae.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListAdverseEventsByEnrollment(t *testing.T) {
	svc := newTestService()
	eid := uuid.New()
	svc.CreateAdverseEvent(nil, &ResearchAdverseEvent{EnrollmentID: eid, Description: "AE1"})
	svc.CreateAdverseEvent(nil, &ResearchAdverseEvent{EnrollmentID: eid, Description: "AE2"})
	items, total, err := svc.ListAdverseEventsByEnrollment(nil, eid, 10, 0)
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

// ── Protocol Deviation Tests ──

func TestService_CreateDeviation(t *testing.T) {
	svc := newTestService()
	d := &ResearchProtocolDeviation{EnrollmentID: uuid.New(), Description: "Wrong dosage"}
	if err := svc.CreateDeviation(nil, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

func TestService_CreateDeviation_MissingEnrollmentID(t *testing.T) {
	svc := newTestService()
	d := &ResearchProtocolDeviation{Description: "Wrong dosage"}
	if err := svc.CreateDeviation(nil, d); err == nil {
		t.Error("expected error for missing enrollment_id")
	}
}

func TestService_CreateDeviation_MissingDescription(t *testing.T) {
	svc := newTestService()
	d := &ResearchProtocolDeviation{EnrollmentID: uuid.New()}
	if err := svc.CreateDeviation(nil, d); err == nil {
		t.Error("expected error for missing description")
	}
}

func TestService_GetDeviation(t *testing.T) {
	svc := newTestService()
	d := &ResearchProtocolDeviation{EnrollmentID: uuid.New(), Description: "Wrong dosage"}
	svc.CreateDeviation(nil, d)
	got, err := svc.GetDeviation(nil, d.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Description != "Wrong dosage" {
		t.Errorf("expected 'Wrong dosage', got %s", got.Description)
	}
}

func TestService_UpdateDeviation(t *testing.T) {
	svc := newTestService()
	d := &ResearchProtocolDeviation{EnrollmentID: uuid.New(), Description: "Wrong dosage"}
	svc.CreateDeviation(nil, d)
	d.Description = "Corrected dosage issue"
	if err := svc.UpdateDeviation(nil, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_DeleteDeviation(t *testing.T) {
	svc := newTestService()
	d := &ResearchProtocolDeviation{EnrollmentID: uuid.New(), Description: "Wrong dosage"}
	svc.CreateDeviation(nil, d)
	if err := svc.DeleteDeviation(nil, d.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_ListDeviationsByEnrollment(t *testing.T) {
	svc := newTestService()
	eid := uuid.New()
	svc.CreateDeviation(nil, &ResearchProtocolDeviation{EnrollmentID: eid, Description: "D1"})
	svc.CreateDeviation(nil, &ResearchProtocolDeviation{EnrollmentID: eid, Description: "D2"})
	items, total, err := svc.ListDeviationsByEnrollment(nil, eid, 10, 0)
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
