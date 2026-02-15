package careteam

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// -- Mock Repository --

type mockCareTeamRepo struct {
	store        map[uuid.UUID]*CareTeam
	participants map[uuid.UUID][]*CareTeamParticipant
}

func newMockCareTeamRepo() *mockCareTeamRepo {
	return &mockCareTeamRepo{
		store:        make(map[uuid.UUID]*CareTeam),
		participants: make(map[uuid.UUID][]*CareTeamParticipant),
	}
}

func (m *mockCareTeamRepo) Create(_ context.Context, ct *CareTeam) error {
	ct.ID = uuid.New()
	if ct.FHIRID == "" {
		ct.FHIRID = ct.ID.String()
	}
	m.store[ct.ID] = ct
	return nil
}

func (m *mockCareTeamRepo) GetByID(_ context.Context, id uuid.UUID) (*CareTeam, error) {
	ct, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return ct, nil
}

func (m *mockCareTeamRepo) GetByFHIRID(_ context.Context, fhirID string) (*CareTeam, error) {
	for _, ct := range m.store {
		if ct.FHIRID == fhirID {
			return ct, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockCareTeamRepo) Update(_ context.Context, ct *CareTeam) error {
	if _, ok := m.store[ct.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[ct.ID] = ct
	return nil
}

func (m *mockCareTeamRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockCareTeamRepo) ListByPatient(_ context.Context, pid uuid.UUID, limit, offset int) ([]*CareTeam, int, error) {
	var r []*CareTeam
	for _, ct := range m.store {
		if ct.PatientID == pid {
			r = append(r, ct)
		}
	}
	return r, len(r), nil
}

func (m *mockCareTeamRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*CareTeam, int, error) {
	var r []*CareTeam
	for _, ct := range m.store {
		r = append(r, ct)
	}
	return r, len(r), nil
}

func (m *mockCareTeamRepo) AddParticipant(_ context.Context, careTeamID uuid.UUID, p *CareTeamParticipant) error {
	p.ID = uuid.New()
	p.CareTeamID = careTeamID
	m.participants[careTeamID] = append(m.participants[careTeamID], p)
	return nil
}

func (m *mockCareTeamRepo) RemoveParticipant(_ context.Context, careTeamID uuid.UUID, participantID uuid.UUID) error {
	parts := m.participants[careTeamID]
	for i, p := range parts {
		if p.ID == participantID {
			m.participants[careTeamID] = append(parts[:i], parts[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("participant not found")
}

func (m *mockCareTeamRepo) GetParticipants(_ context.Context, careTeamID uuid.UUID) ([]*CareTeamParticipant, error) {
	return m.participants[careTeamID], nil
}

func newTestService() *Service {
	return NewService(newMockCareTeamRepo())
}

// -- Service Tests --

func TestCreateCareTeam_Success(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	if err := svc.CreateCareTeam(context.Background(), ct); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if ct.FHIRID == "" {
		t.Error("expected FHIRID to be set")
	}
}

func TestCreateCareTeam_MissingStatus(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New()}
	if err := svc.CreateCareTeam(context.Background(), ct); err == nil {
		t.Fatal("expected error for missing status")
	}
}

func TestCreateCareTeam_MissingPatient(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{Status: "active"}
	if err := svc.CreateCareTeam(context.Background(), ct); err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestCreateCareTeam_InvalidStatus(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New(), Status: "bogus"}
	if err := svc.CreateCareTeam(context.Background(), ct); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCreateCareTeam_ValidStatuses(t *testing.T) {
	for _, s := range []string{"proposed", "active", "suspended", "inactive", "entered-in-error"} {
		svc := newTestService()
		ct := &CareTeam{PatientID: uuid.New(), Status: s}
		if err := svc.CreateCareTeam(context.Background(), ct); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

func TestGetCareTeam_Success(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	svc.CreateCareTeam(context.Background(), ct)
	got, err := svc.GetCareTeam(context.Background(), ct.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != ct.ID {
		t.Error("ID mismatch")
	}
}

func TestGetCareTeam_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetCareTeam(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateCareTeam_Success(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	svc.CreateCareTeam(context.Background(), ct)
	ct.Status = "inactive"
	if err := svc.UpdateCareTeam(context.Background(), ct); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetCareTeam(context.Background(), ct.ID)
	if got.Status != "inactive" {
		t.Errorf("expected status 'inactive', got %q", got.Status)
	}
}

func TestUpdateCareTeam_InvalidStatus(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	svc.CreateCareTeam(context.Background(), ct)
	ct.Status = "invalid"
	if err := svc.UpdateCareTeam(context.Background(), ct); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestDeleteCareTeam_Success(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	svc.CreateCareTeam(context.Background(), ct)
	if err := svc.DeleteCareTeam(context.Background(), ct.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.GetCareTeam(context.Background(), ct.ID); err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListCareTeamsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateCareTeam(context.Background(), &CareTeam{PatientID: pid, Status: "active"})
	svc.CreateCareTeam(context.Background(), &CareTeam{PatientID: pid, Status: "proposed"})
	svc.CreateCareTeam(context.Background(), &CareTeam{PatientID: uuid.New(), Status: "active"})
	items, total, err := svc.ListCareTeamsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 care teams, got %d", total)
	}
}

func TestSearchCareTeams(t *testing.T) {
	svc := newTestService()
	svc.CreateCareTeam(context.Background(), &CareTeam{PatientID: uuid.New(), Status: "active"})
	svc.CreateCareTeam(context.Background(), &CareTeam{PatientID: uuid.New(), Status: "proposed"})
	items, total, err := svc.SearchCareTeams(context.Background(), map[string]string{"status": "active"}, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 care teams from search, got %d", total)
	}
}

func TestAddParticipant_Success(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	svc.CreateCareTeam(context.Background(), ct)
	p := &CareTeamParticipant{
		MemberID:   uuid.New(),
		MemberType: "Practitioner",
	}
	if err := svc.AddParticipant(context.Background(), ct.ID, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	participants, _ := svc.GetParticipants(context.Background(), ct.ID)
	if len(participants) != 1 {
		t.Errorf("expected 1 participant, got %d", len(participants))
	}
}

func TestAddParticipant_MissingMember(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	svc.CreateCareTeam(context.Background(), ct)
	p := &CareTeamParticipant{
		MemberType: "Practitioner",
	}
	if err := svc.AddParticipant(context.Background(), ct.ID, p); err == nil {
		t.Fatal("expected error for missing member_id")
	}
}

func TestRemoveParticipant_Success(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	svc.CreateCareTeam(context.Background(), ct)
	p := &CareTeamParticipant{
		MemberID:   uuid.New(),
		MemberType: "Practitioner",
	}
	svc.AddParticipant(context.Background(), ct.ID, p)
	if err := svc.RemoveParticipant(context.Background(), ct.ID, p.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	participants, _ := svc.GetParticipants(context.Background(), ct.ID)
	if len(participants) != 0 {
		t.Errorf("expected 0 participants after removal, got %d", len(participants))
	}
}

func TestGetCareTeamByFHIRID(t *testing.T) {
	svc := newTestService()
	ct := &CareTeam{PatientID: uuid.New(), Status: "active"}
	svc.CreateCareTeam(context.Background(), ct)
	got, err := svc.GetCareTeamByFHIRID(context.Background(), ct.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != ct.ID {
		t.Error("ID mismatch")
	}
}
