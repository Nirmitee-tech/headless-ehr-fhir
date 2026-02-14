package relatedperson

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// =========== Mock Repository ===========

type mockRelatedPersonRepo struct {
	store          map[uuid.UUID]*RelatedPerson
	communications map[uuid.UUID][]*RelatedPersonCommunication
}

func newMockRelatedPersonRepo() *mockRelatedPersonRepo {
	return &mockRelatedPersonRepo{store: make(map[uuid.UUID]*RelatedPerson), communications: make(map[uuid.UUID][]*RelatedPersonCommunication)}
}

func (m *mockRelatedPersonRepo) Create(_ context.Context, rp *RelatedPerson) error {
	rp.ID = uuid.New()
	if rp.FHIRID == "" {
		rp.FHIRID = rp.ID.String()
	}
	m.store[rp.ID] = rp
	return nil
}

func (m *mockRelatedPersonRepo) GetByID(_ context.Context, id uuid.UUID) (*RelatedPerson, error) {
	rp, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return rp, nil
}

func (m *mockRelatedPersonRepo) GetByFHIRID(_ context.Context, fhirID string) (*RelatedPerson, error) {
	for _, rp := range m.store {
		if rp.FHIRID == fhirID {
			return rp, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockRelatedPersonRepo) Update(_ context.Context, rp *RelatedPerson) error {
	if _, ok := m.store[rp.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[rp.ID] = rp
	return nil
}

func (m *mockRelatedPersonRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockRelatedPersonRepo) List(_ context.Context, limit, offset int) ([]*RelatedPerson, int, error) {
	var result []*RelatedPerson
	for _, rp := range m.store {
		result = append(result, rp)
	}
	return result, len(result), nil
}

func (m *mockRelatedPersonRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*RelatedPerson, int, error) {
	var result []*RelatedPerson
	for _, rp := range m.store {
		if rp.PatientID == patientID {
			result = append(result, rp)
		}
	}
	return result, len(result), nil
}

func (m *mockRelatedPersonRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*RelatedPerson, int, error) {
	var result []*RelatedPerson
	for _, rp := range m.store {
		result = append(result, rp)
	}
	return result, len(result), nil
}

func (m *mockRelatedPersonRepo) AddCommunication(_ context.Context, c *RelatedPersonCommunication) error {
	c.ID = uuid.New()
	m.communications[c.RelatedPersonID] = append(m.communications[c.RelatedPersonID], c)
	return nil
}

func (m *mockRelatedPersonRepo) GetCommunications(_ context.Context, relatedPersonID uuid.UUID) ([]*RelatedPersonCommunication, error) {
	return m.communications[relatedPersonID], nil
}

// =========== Helper ===========

func newTestService() *Service {
	return NewService(newMockRelatedPersonRepo())
}

// =========== RelatedPerson Tests ===========

func TestCreateRelatedPerson_Success(t *testing.T) {
	svc := newTestService()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	if err := svc.CreateRelatedPerson(context.Background(), rp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateRelatedPerson_MissingPatient(t *testing.T) {
	svc := newTestService()
	rp := &RelatedPerson{RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	if err := svc.CreateRelatedPerson(context.Background(), rp); err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestCreateRelatedPerson_MissingRelationshipCode(t *testing.T) {
	svc := newTestService()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipDisplay: "Wife"}
	if err := svc.CreateRelatedPerson(context.Background(), rp); err == nil {
		t.Fatal("expected error for missing relationship_code")
	}
}

func TestCreateRelatedPerson_MissingRelationshipDisplay(t *testing.T) {
	svc := newTestService()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE"}
	if err := svc.CreateRelatedPerson(context.Background(), rp); err == nil {
		t.Fatal("expected error for missing relationship_display")
	}
}

func TestGetRelatedPerson(t *testing.T) {
	svc := newTestService()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	svc.CreateRelatedPerson(context.Background(), rp)

	got, err := svc.GetRelatedPerson(context.Background(), rp.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != rp.ID {
		t.Errorf("expected ID %v, got %v", rp.ID, got.ID)
	}
}

func TestGetRelatedPerson_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetRelatedPerson(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetRelatedPersonByFHIRID(t *testing.T) {
	svc := newTestService()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	svc.CreateRelatedPerson(context.Background(), rp)

	got, err := svc.GetRelatedPersonByFHIRID(context.Background(), rp.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != rp.ID {
		t.Errorf("expected ID %v, got %v", rp.ID, got.ID)
	}
}

func TestGetRelatedPersonByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetRelatedPersonByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestDeleteRelatedPerson(t *testing.T) {
	svc := newTestService()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	svc.CreateRelatedPerson(context.Background(), rp)
	if err := svc.DeleteRelatedPerson(context.Background(), rp.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetRelatedPerson(context.Background(), rp.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListRelatedPersonsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateRelatedPerson(context.Background(), &RelatedPerson{PatientID: pid, RelationshipCode: "WIFE", RelationshipDisplay: "Wife"})
	svc.CreateRelatedPerson(context.Background(), &RelatedPerson{PatientID: pid, RelationshipCode: "CHILD", RelationshipDisplay: "Child"})
	svc.CreateRelatedPerson(context.Background(), &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "SIB", RelationshipDisplay: "Sibling"})

	items, total, err := svc.ListRelatedPersonsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 related persons, got %d", total)
	}
}

func TestSearchRelatedPersons(t *testing.T) {
	svc := newTestService()
	svc.CreateRelatedPerson(context.Background(), &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"})
	items, total, err := svc.SearchRelatedPersons(context.Background(), map[string]string{"relationship": "WIFE"}, 10, 0)
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

// =========== Communication Tests ===========

func TestAddCommunication_Success(t *testing.T) {
	svc := newTestService()
	rp := &RelatedPerson{PatientID: uuid.New(), RelationshipCode: "WIFE", RelationshipDisplay: "Wife"}
	svc.CreateRelatedPerson(context.Background(), rp)
	c := &RelatedPersonCommunication{RelatedPersonID: rp.ID, LanguageCode: "en", LanguageDisplay: "English"}
	if err := svc.AddCommunication(context.Background(), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	comms, _ := svc.GetCommunications(context.Background(), rp.ID)
	if len(comms) != 1 {
		t.Errorf("expected 1 communication, got %d", len(comms))
	}
}

func TestAddCommunication_MissingRelatedPersonID(t *testing.T) {
	svc := newTestService()
	c := &RelatedPersonCommunication{LanguageCode: "en", LanguageDisplay: "English"}
	if err := svc.AddCommunication(context.Background(), c); err == nil {
		t.Fatal("expected error for missing related_person_id")
	}
}

func TestAddCommunication_MissingLanguageCode(t *testing.T) {
	svc := newTestService()
	c := &RelatedPersonCommunication{RelatedPersonID: uuid.New(), LanguageDisplay: "English"}
	if err := svc.AddCommunication(context.Background(), c); err == nil {
		t.Fatal("expected error for missing language_code")
	}
}
