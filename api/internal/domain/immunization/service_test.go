package immunization

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// =========== Mock Repositories ===========

type mockImmunizationRepo struct {
	store map[uuid.UUID]*Immunization
}

func newMockImmunizationRepo() *mockImmunizationRepo {
	return &mockImmunizationRepo{store: make(map[uuid.UUID]*Immunization)}
}

func (m *mockImmunizationRepo) Create(_ context.Context, im *Immunization) error {
	im.ID = uuid.New()
	if im.FHIRID == "" {
		im.FHIRID = im.ID.String()
	}
	m.store[im.ID] = im
	return nil
}

func (m *mockImmunizationRepo) GetByID(_ context.Context, id uuid.UUID) (*Immunization, error) {
	im, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return im, nil
}

func (m *mockImmunizationRepo) GetByFHIRID(_ context.Context, fhirID string) (*Immunization, error) {
	for _, im := range m.store {
		if im.FHIRID == fhirID {
			return im, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockImmunizationRepo) Update(_ context.Context, im *Immunization) error {
	if _, ok := m.store[im.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[im.ID] = im
	return nil
}

func (m *mockImmunizationRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockImmunizationRepo) List(_ context.Context, limit, offset int) ([]*Immunization, int, error) {
	var result []*Immunization
	for _, im := range m.store {
		result = append(result, im)
	}
	return result, len(result), nil
}

func (m *mockImmunizationRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Immunization, int, error) {
	var result []*Immunization
	for _, im := range m.store {
		if im.PatientID == patientID {
			result = append(result, im)
		}
	}
	return result, len(result), nil
}

func (m *mockImmunizationRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Immunization, int, error) {
	var result []*Immunization
	for _, im := range m.store {
		result = append(result, im)
	}
	return result, len(result), nil
}

// -- Mock Recommendation Repo --

type mockRecommendationRepo struct {
	store map[uuid.UUID]*ImmunizationRecommendation
}

func newMockRecommendationRepo() *mockRecommendationRepo {
	return &mockRecommendationRepo{store: make(map[uuid.UUID]*ImmunizationRecommendation)}
}

func (m *mockRecommendationRepo) Create(_ context.Context, r *ImmunizationRecommendation) error {
	r.ID = uuid.New()
	if r.FHIRID == "" {
		r.FHIRID = r.ID.String()
	}
	m.store[r.ID] = r
	return nil
}

func (m *mockRecommendationRepo) GetByID(_ context.Context, id uuid.UUID) (*ImmunizationRecommendation, error) {
	r, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

func (m *mockRecommendationRepo) GetByFHIRID(_ context.Context, fhirID string) (*ImmunizationRecommendation, error) {
	for _, r := range m.store {
		if r.FHIRID == fhirID {
			return r, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockRecommendationRepo) Update(_ context.Context, r *ImmunizationRecommendation) error {
	if _, ok := m.store[r.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[r.ID] = r
	return nil
}

func (m *mockRecommendationRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockRecommendationRepo) List(_ context.Context, limit, offset int) ([]*ImmunizationRecommendation, int, error) {
	var result []*ImmunizationRecommendation
	for _, r := range m.store {
		result = append(result, r)
	}
	return result, len(result), nil
}

func (m *mockRecommendationRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*ImmunizationRecommendation, int, error) {
	var result []*ImmunizationRecommendation
	for _, r := range m.store {
		if r.PatientID == patientID {
			result = append(result, r)
		}
	}
	return result, len(result), nil
}

func (m *mockRecommendationRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*ImmunizationRecommendation, int, error) {
	var result []*ImmunizationRecommendation
	for _, r := range m.store {
		result = append(result, r)
	}
	return result, len(result), nil
}

// =========== Helper ===========

func newTestService() *Service {
	return NewService(newMockImmunizationRepo(), newMockRecommendationRepo())
}

// =========== Immunization Tests ===========

func TestCreateImmunization_Success(t *testing.T) {
	svc := newTestService()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	if err := svc.CreateImmunization(context.Background(), im); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if im.Status != "completed" {
		t.Errorf("expected default status 'completed', got %q", im.Status)
	}
}

func TestCreateImmunization_MissingPatient(t *testing.T) {
	svc := newTestService()
	im := &Immunization{VaccineCode: "08", VaccineDisplay: "Hep B"}
	if err := svc.CreateImmunization(context.Background(), im); err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestCreateImmunization_MissingVaccineCode(t *testing.T) {
	svc := newTestService()
	im := &Immunization{PatientID: uuid.New(), VaccineDisplay: "Hep B"}
	if err := svc.CreateImmunization(context.Background(), im); err == nil {
		t.Fatal("expected error for missing vaccine_code")
	}
}

func TestCreateImmunization_MissingVaccineDisplay(t *testing.T) {
	svc := newTestService()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08"}
	if err := svc.CreateImmunization(context.Background(), im); err == nil {
		t.Fatal("expected error for missing vaccine_display")
	}
}

func TestCreateImmunization_InvalidStatus(t *testing.T) {
	svc := newTestService()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B", Status: "bogus"}
	if err := svc.CreateImmunization(context.Background(), im); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCreateImmunization_ValidStatuses(t *testing.T) {
	for _, s := range []string{"completed", "entered-in-error", "not-done"} {
		svc := newTestService()
		im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B", Status: s}
		if err := svc.CreateImmunization(context.Background(), im); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

func TestGetImmunization(t *testing.T) {
	svc := newTestService()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	svc.CreateImmunization(context.Background(), im)

	got, err := svc.GetImmunization(context.Background(), im.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != im.ID {
		t.Errorf("expected ID %v, got %v", im.ID, got.ID)
	}
}

func TestGetImmunization_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetImmunization(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetImmunizationByFHIRID(t *testing.T) {
	svc := newTestService()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	svc.CreateImmunization(context.Background(), im)

	got, err := svc.GetImmunizationByFHIRID(context.Background(), im.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != im.ID {
		t.Errorf("expected ID %v, got %v", im.ID, got.ID)
	}
}

func TestGetImmunizationByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetImmunizationByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateImmunization_InvalidStatus(t *testing.T) {
	svc := newTestService()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	svc.CreateImmunization(context.Background(), im)
	im.Status = "invalid"
	if err := svc.UpdateImmunization(context.Background(), im); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestDeleteImmunization(t *testing.T) {
	svc := newTestService()
	im := &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"}
	svc.CreateImmunization(context.Background(), im)
	if err := svc.DeleteImmunization(context.Background(), im.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetImmunization(context.Background(), im.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListImmunizationsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateImmunization(context.Background(), &Immunization{PatientID: pid, VaccineCode: "08", VaccineDisplay: "Hep B"})
	svc.CreateImmunization(context.Background(), &Immunization{PatientID: pid, VaccineCode: "03", VaccineDisplay: "MMR"})
	svc.CreateImmunization(context.Background(), &Immunization{PatientID: uuid.New(), VaccineCode: "10", VaccineDisplay: "IPV"})

	items, total, err := svc.ListImmunizationsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 immunizations, got %d", total)
	}
}

func TestSearchImmunizations(t *testing.T) {
	svc := newTestService()
	svc.CreateImmunization(context.Background(), &Immunization{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B"})
	items, total, err := svc.SearchImmunizations(context.Background(), map[string]string{"vaccine-code": "08"}, 10, 0)
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

// =========== Recommendation Tests ===========

func TestCreateRecommendation_Success(t *testing.T) {
	svc := newTestService()
	r := &ImmunizationRecommendation{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B", ForecastStatus: "due", Date: time.Now()}
	if err := svc.CreateRecommendation(context.Background(), r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateRecommendation_MissingPatient(t *testing.T) {
	svc := newTestService()
	r := &ImmunizationRecommendation{VaccineCode: "08", VaccineDisplay: "Hep B", ForecastStatus: "due", Date: time.Now()}
	if err := svc.CreateRecommendation(context.Background(), r); err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestCreateRecommendation_MissingVaccineCode(t *testing.T) {
	svc := newTestService()
	r := &ImmunizationRecommendation{PatientID: uuid.New(), ForecastStatus: "due", Date: time.Now()}
	if err := svc.CreateRecommendation(context.Background(), r); err == nil {
		t.Fatal("expected error for missing vaccine_code")
	}
}

func TestCreateRecommendation_MissingForecastStatus(t *testing.T) {
	svc := newTestService()
	r := &ImmunizationRecommendation{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B", Date: time.Now()}
	if err := svc.CreateRecommendation(context.Background(), r); err == nil {
		t.Fatal("expected error for missing forecast_status")
	}
}

func TestGetRecommendation(t *testing.T) {
	svc := newTestService()
	r := &ImmunizationRecommendation{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B", ForecastStatus: "due", Date: time.Now()}
	svc.CreateRecommendation(context.Background(), r)

	got, err := svc.GetRecommendation(context.Background(), r.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != r.ID {
		t.Errorf("expected ID %v, got %v", r.ID, got.ID)
	}
}

func TestGetRecommendation_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetRecommendation(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestDeleteRecommendation(t *testing.T) {
	svc := newTestService()
	r := &ImmunizationRecommendation{PatientID: uuid.New(), VaccineCode: "08", VaccineDisplay: "Hep B", ForecastStatus: "due", Date: time.Now()}
	svc.CreateRecommendation(context.Background(), r)
	if err := svc.DeleteRecommendation(context.Background(), r.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := svc.GetRecommendation(context.Background(), r.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListRecommendationsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateRecommendation(context.Background(), &ImmunizationRecommendation{PatientID: pid, VaccineCode: "08", VaccineDisplay: "Hep B", ForecastStatus: "due", Date: time.Now()})
	svc.CreateRecommendation(context.Background(), &ImmunizationRecommendation{PatientID: pid, VaccineCode: "03", VaccineDisplay: "MMR", ForecastStatus: "due", Date: time.Now()})
	svc.CreateRecommendation(context.Background(), &ImmunizationRecommendation{PatientID: uuid.New(), VaccineCode: "10", VaccineDisplay: "IPV", ForecastStatus: "due", Date: time.Now()})

	items, total, err := svc.ListRecommendationsByPatient(context.Background(), pid, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 recommendations, got %d", total)
	}
}
