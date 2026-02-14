package familyhistory

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// =========== Mock Repository ===========

type mockFMHRepo struct {
	store      map[uuid.UUID]*FamilyMemberHistory
	conditions map[uuid.UUID][]*FamilyMemberCondition
}

func newMockFMHRepo() *mockFMHRepo {
	return &mockFMHRepo{store: make(map[uuid.UUID]*FamilyMemberHistory), conditions: make(map[uuid.UUID][]*FamilyMemberCondition)}
}

func (m *mockFMHRepo) Create(_ context.Context, f *FamilyMemberHistory) error {
	f.ID = uuid.New()
	if f.FHIRID == "" {
		f.FHIRID = f.ID.String()
	}
	m.store[f.ID] = f
	return nil
}

func (m *mockFMHRepo) GetByID(_ context.Context, id uuid.UUID) (*FamilyMemberHistory, error) {
	f, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return f, nil
}

func (m *mockFMHRepo) GetByFHIRID(_ context.Context, fhirID string) (*FamilyMemberHistory, error) {
	for _, f := range m.store {
		if f.FHIRID == fhirID {
			return f, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockFMHRepo) Update(_ context.Context, f *FamilyMemberHistory) error {
	if _, ok := m.store[f.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[f.ID] = f
	return nil
}

func (m *mockFMHRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockFMHRepo) List(_ context.Context, limit, offset int) ([]*FamilyMemberHistory, int, error) {
	var result []*FamilyMemberHistory
	for _, f := range m.store {
		result = append(result, f)
	}
	return result, len(result), nil
}

func (m *mockFMHRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*FamilyMemberHistory, int, error) {
	var result []*FamilyMemberHistory
	for _, f := range m.store {
		if f.PatientID == patientID {
			result = append(result, f)
		}
	}
	return result, len(result), nil
}

func (m *mockFMHRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*FamilyMemberHistory, int, error) {
	var result []*FamilyMemberHistory
	for _, f := range m.store {
		result = append(result, f)
	}
	return result, len(result), nil
}

func (m *mockFMHRepo) AddCondition(_ context.Context, c *FamilyMemberCondition) error {
	c.ID = uuid.New()
	m.conditions[c.FamilyMemberID] = append(m.conditions[c.FamilyMemberID], c)
	return nil
}

func (m *mockFMHRepo) GetConditions(_ context.Context, familyMemberID uuid.UUID) ([]*FamilyMemberCondition, error) {
	return m.conditions[familyMemberID], nil
}

// =========== Helper ===========

func newTestService() *Service {
	return NewService(newMockFMHRepo())
}

// =========== FamilyMemberHistory Tests ===========

func TestCreateFamilyMemberHistory_Success(t *testing.T) {
	svc := newTestService()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	if err := svc.CreateFamilyMemberHistory(context.Background(), f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Status != "completed" {
		t.Errorf("expected default status 'completed', got %q", f.Status)
	}
}

func TestCreateFamilyMemberHistory_MissingPatient(t *testing.T) {
	svc := newTestService()
	f := &FamilyMemberHistory{RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	if err := svc.CreateFamilyMemberHistory(context.Background(), f); err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestCreateFamilyMemberHistory_MissingRelationshipCode(t *testing.T) {
	svc := newTestService()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipDisplay: "Father"}
	if err := svc.CreateFamilyMemberHistory(context.Background(), f); err == nil {
		t.Fatal("expected error for missing relationship_code")
	}
}

func TestCreateFamilyMemberHistory_MissingRelationshipDisplay(t *testing.T) {
	svc := newTestService()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH"}
	if err := svc.CreateFamilyMemberHistory(context.Background(), f); err == nil {
		t.Fatal("expected error for missing relationship_display")
	}
}

func TestCreateFamilyMemberHistory_InvalidStatus(t *testing.T) {
	svc := newTestService()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father", Status: "bogus"}
	if err := svc.CreateFamilyMemberHistory(context.Background(), f); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCreateFamilyMemberHistory_ValidStatuses(t *testing.T) {
	for _, s := range []string{"partial", "completed", "entered-in-error", "health-unknown"} {
		svc := newTestService()
		f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father", Status: s}
		if err := svc.CreateFamilyMemberHistory(context.Background(), f); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

func TestGetFamilyMemberHistory(t *testing.T) {
	svc := newTestService()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	svc.CreateFamilyMemberHistory(context.Background(), f)

	got, err := svc.GetFamilyMemberHistory(context.Background(), f.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != f.ID {
		t.Errorf("expected ID %v, got %v", f.ID, got.ID)
	}
}

func TestGetFamilyMemberHistory_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetFamilyMemberHistory(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetFamilyMemberHistoryByFHIRID(t *testing.T) {
	svc := newTestService()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	svc.CreateFamilyMemberHistory(context.Background(), f)

	got, err := svc.GetFamilyMemberHistoryByFHIRID(context.Background(), f.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != f.ID {
		t.Errorf("expected ID %v, got %v", f.ID, got.ID)
	}
}

func TestGetFamilyMemberHistoryByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetFamilyMemberHistoryByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateFamilyMemberHistory_InvalidStatus(t *testing.T) {
	svc := newTestService()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	svc.CreateFamilyMemberHistory(context.Background(), f)
	f.Status = "invalid"
	if err := svc.UpdateFamilyMemberHistory(context.Background(), f); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestDeleteFamilyMemberHistory(t *testing.T) {
	svc := newTestService()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	svc.CreateFamilyMemberHistory(context.Background(), f)
	if err := svc.DeleteFamilyMemberHistory(context.Background(), f.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetFamilyMemberHistory(context.Background(), f.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListFamilyMemberHistoriesByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateFamilyMemberHistory(context.Background(), &FamilyMemberHistory{PatientID: pid, RelationshipCode: "FTH", RelationshipDisplay: "Father"})
	svc.CreateFamilyMemberHistory(context.Background(), &FamilyMemberHistory{PatientID: pid, RelationshipCode: "MTH", RelationshipDisplay: "Mother"})
	svc.CreateFamilyMemberHistory(context.Background(), &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "SIS", RelationshipDisplay: "Sister"})

	items, total, err := svc.ListFamilyMemberHistoriesByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 family member histories, got %d", total)
	}
}

func TestSearchFamilyMemberHistories(t *testing.T) {
	svc := newTestService()
	svc.CreateFamilyMemberHistory(context.Background(), &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"})
	items, total, err := svc.SearchFamilyMemberHistories(context.Background(), map[string]string{"status": "completed"}, 10, 0)
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

// =========== Condition Tests ===========

func TestAddCondition_Success(t *testing.T) {
	svc := newTestService()
	f := &FamilyMemberHistory{PatientID: uuid.New(), RelationshipCode: "FTH", RelationshipDisplay: "Father"}
	svc.CreateFamilyMemberHistory(context.Background(), f)
	c := &FamilyMemberCondition{FamilyMemberID: f.ID, Code: "I25.10", Display: "Coronary artery disease"}
	if err := svc.AddCondition(context.Background(), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	conditions, _ := svc.GetConditions(context.Background(), f.ID)
	if len(conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(conditions))
	}
}

func TestAddCondition_MissingFamilyMemberID(t *testing.T) {
	svc := newTestService()
	c := &FamilyMemberCondition{Code: "I25.10", Display: "Coronary artery disease"}
	if err := svc.AddCondition(context.Background(), c); err == nil {
		t.Fatal("expected error for missing family_member_id")
	}
}

func TestAddCondition_MissingCode(t *testing.T) {
	svc := newTestService()
	c := &FamilyMemberCondition{FamilyMemberID: uuid.New(), Display: "Coronary artery disease"}
	if err := svc.AddCondition(context.Background(), c); err == nil {
		t.Fatal("expected error for missing code")
	}
}

func TestAddCondition_MissingDisplay(t *testing.T) {
	svc := newTestService()
	c := &FamilyMemberCondition{FamilyMemberID: uuid.New(), Code: "I25.10"}
	if err := svc.AddCondition(context.Background(), c); err == nil {
		t.Fatal("expected error for missing display")
	}
}
