package careplan

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

type mockCarePlanRepo struct {
	store      map[uuid.UUID]*CarePlan
	activities map[uuid.UUID][]*CarePlanActivity
}

func newMockCarePlanRepo() *mockCarePlanRepo {
	return &mockCarePlanRepo{store: make(map[uuid.UUID]*CarePlan), activities: make(map[uuid.UUID][]*CarePlanActivity)}
}
func (m *mockCarePlanRepo) Create(_ context.Context, cp *CarePlan) error {
	cp.ID = uuid.New(); if cp.FHIRID == "" { cp.FHIRID = cp.ID.String() }; m.store[cp.ID] = cp; return nil
}
func (m *mockCarePlanRepo) GetByID(_ context.Context, id uuid.UUID) (*CarePlan, error) {
	cp, ok := m.store[id]; if !ok { return nil, fmt.Errorf("not found") }; return cp, nil
}
func (m *mockCarePlanRepo) GetByFHIRID(_ context.Context, fhirID string) (*CarePlan, error) {
	for _, cp := range m.store { if cp.FHIRID == fhirID { return cp, nil } }; return nil, fmt.Errorf("not found")
}
func (m *mockCarePlanRepo) Update(_ context.Context, cp *CarePlan) error {
	if _, ok := m.store[cp.ID]; !ok { return fmt.Errorf("not found") }; m.store[cp.ID] = cp; return nil
}
func (m *mockCarePlanRepo) Delete(_ context.Context, id uuid.UUID) error { delete(m.store, id); return nil }
func (m *mockCarePlanRepo) List(_ context.Context, limit, offset int) ([]*CarePlan, int, error) {
	var r []*CarePlan; for _, cp := range m.store { r = append(r, cp) }; return r, len(r), nil
}
func (m *mockCarePlanRepo) ListByPatient(_ context.Context, pid uuid.UUID, limit, offset int) ([]*CarePlan, int, error) {
	var r []*CarePlan; for _, cp := range m.store { if cp.PatientID == pid { r = append(r, cp) } }; return r, len(r), nil
}
func (m *mockCarePlanRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*CarePlan, int, error) {
	var r []*CarePlan; for _, cp := range m.store { r = append(r, cp) }; return r, len(r), nil
}
func (m *mockCarePlanRepo) AddActivity(_ context.Context, a *CarePlanActivity) error {
	a.ID = uuid.New(); m.activities[a.CarePlanID] = append(m.activities[a.CarePlanID], a); return nil
}
func (m *mockCarePlanRepo) GetActivities(_ context.Context, cpID uuid.UUID) ([]*CarePlanActivity, error) {
	return m.activities[cpID], nil
}

type mockGoalRepo struct{ store map[uuid.UUID]*Goal }

func newMockGoalRepo() *mockGoalRepo { return &mockGoalRepo{store: make(map[uuid.UUID]*Goal)} }
func (m *mockGoalRepo) Create(_ context.Context, g *Goal) error {
	g.ID = uuid.New(); if g.FHIRID == "" { g.FHIRID = g.ID.String() }; m.store[g.ID] = g; return nil
}
func (m *mockGoalRepo) GetByID(_ context.Context, id uuid.UUID) (*Goal, error) {
	g, ok := m.store[id]; if !ok { return nil, fmt.Errorf("not found") }; return g, nil
}
func (m *mockGoalRepo) GetByFHIRID(_ context.Context, fhirID string) (*Goal, error) {
	for _, g := range m.store { if g.FHIRID == fhirID { return g, nil } }; return nil, fmt.Errorf("not found")
}
func (m *mockGoalRepo) Update(_ context.Context, g *Goal) error {
	if _, ok := m.store[g.ID]; !ok { return fmt.Errorf("not found") }; m.store[g.ID] = g; return nil
}
func (m *mockGoalRepo) Delete(_ context.Context, id uuid.UUID) error { delete(m.store, id); return nil }
func (m *mockGoalRepo) List(_ context.Context, limit, offset int) ([]*Goal, int, error) {
	var r []*Goal; for _, g := range m.store { r = append(r, g) }; return r, len(r), nil
}
func (m *mockGoalRepo) ListByPatient(_ context.Context, pid uuid.UUID, limit, offset int) ([]*Goal, int, error) {
	var r []*Goal; for _, g := range m.store { if g.PatientID == pid { r = append(r, g) } }; return r, len(r), nil
}
func (m *mockGoalRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Goal, int, error) {
	var r []*Goal; for _, g := range m.store { r = append(r, g) }; return r, len(r), nil
}

func newTestService() *Service { return NewService(newMockCarePlanRepo(), newMockGoalRepo()) }

func TestCreateCarePlan_Success(t *testing.T) {
	svc := newTestService()
	cp := &CarePlan{PatientID: uuid.New(), Intent: "plan"}
	if err := svc.CreateCarePlan(context.Background(), cp); err != nil { t.Fatalf("unexpected error: %v", err) }
	if cp.Status != "draft" { t.Errorf("expected default status 'draft', got %q", cp.Status) }
}

func TestCreateCarePlan_MissingPatient(t *testing.T) {
	svc := newTestService()
	if err := svc.CreateCarePlan(context.Background(), &CarePlan{Intent: "plan"}); err == nil { t.Fatal("expected error") }
}

func TestCreateCarePlan_MissingIntent(t *testing.T) {
	svc := newTestService()
	if err := svc.CreateCarePlan(context.Background(), &CarePlan{PatientID: uuid.New()}); err == nil { t.Fatal("expected error") }
}

func TestCreateCarePlan_InvalidStatus(t *testing.T) {
	svc := newTestService()
	if err := svc.CreateCarePlan(context.Background(), &CarePlan{PatientID: uuid.New(), Intent: "plan", Status: "bogus"}); err == nil { t.Fatal("expected error") }
}

func TestCreateCarePlan_InvalidIntent(t *testing.T) {
	svc := newTestService()
	if err := svc.CreateCarePlan(context.Background(), &CarePlan{PatientID: uuid.New(), Intent: "bogus"}); err == nil { t.Fatal("expected error") }
}

func TestCreateCarePlan_ValidStatuses(t *testing.T) {
	for _, s := range []string{"draft", "active", "on-hold", "completed", "revoked", "entered-in-error"} {
		svc := newTestService()
		cp := &CarePlan{PatientID: uuid.New(), Intent: "plan", Status: s}
		if err := svc.CreateCarePlan(context.Background(), cp); err != nil { t.Errorf("status %q should be valid: %v", s, err) }
	}
}

func TestGetCarePlan(t *testing.T) {
	svc := newTestService()
	cp := &CarePlan{PatientID: uuid.New(), Intent: "plan"}
	svc.CreateCarePlan(context.Background(), cp)
	got, err := svc.GetCarePlan(context.Background(), cp.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if got.ID != cp.ID { t.Errorf("ID mismatch") }
}

func TestGetCarePlan_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetCarePlan(context.Background(), uuid.New()); err == nil { t.Fatal("expected error") }
}

func TestDeleteCarePlan(t *testing.T) {
	svc := newTestService()
	cp := &CarePlan{PatientID: uuid.New(), Intent: "plan"}
	svc.CreateCarePlan(context.Background(), cp)
	if err := svc.DeleteCarePlan(context.Background(), cp.ID); err != nil { t.Fatalf("unexpected error: %v", err) }
	if _, err := svc.GetCarePlan(context.Background(), cp.ID); err == nil { t.Fatal("expected error after delete") }
}

func TestUpdateCarePlan_InvalidStatus(t *testing.T) {
	svc := newTestService()
	cp := &CarePlan{PatientID: uuid.New(), Intent: "plan"}
	svc.CreateCarePlan(context.Background(), cp)
	cp.Status = "invalid"
	if err := svc.UpdateCarePlan(context.Background(), cp); err == nil { t.Fatal("expected error") }
}

func TestAddActivity(t *testing.T) {
	svc := newTestService()
	cp := &CarePlan{PatientID: uuid.New(), Intent: "plan"}
	svc.CreateCarePlan(context.Background(), cp)
	a := &CarePlanActivity{CarePlanID: cp.ID, Status: "scheduled"}
	if err := svc.AddActivity(context.Background(), a); err != nil { t.Fatalf("unexpected error: %v", err) }
	acts, _ := svc.GetActivities(context.Background(), cp.ID)
	if len(acts) != 1 { t.Errorf("expected 1 activity, got %d", len(acts)) }
}

func TestAddActivity_MissingCarePlanID(t *testing.T) {
	svc := newTestService()
	if err := svc.AddActivity(context.Background(), &CarePlanActivity{Status: "scheduled"}); err == nil { t.Fatal("expected error") }
}

func TestAddActivity_MissingStatus(t *testing.T) {
	svc := newTestService()
	if err := svc.AddActivity(context.Background(), &CarePlanActivity{CarePlanID: uuid.New()}); err == nil { t.Fatal("expected error") }
}

// -- Goal Tests --

func TestCreateGoal_Success(t *testing.T) {
	svc := newTestService()
	g := &Goal{PatientID: uuid.New(), Description: "Reduce A1C"}
	if err := svc.CreateGoal(context.Background(), g); err != nil { t.Fatalf("unexpected error: %v", err) }
	if g.LifecycleStatus != "proposed" { t.Errorf("expected default status 'proposed', got %q", g.LifecycleStatus) }
}

func TestCreateGoal_MissingPatient(t *testing.T) {
	svc := newTestService()
	if err := svc.CreateGoal(context.Background(), &Goal{Description: "test"}); err == nil { t.Fatal("expected error") }
}

func TestCreateGoal_MissingDescription(t *testing.T) {
	svc := newTestService()
	if err := svc.CreateGoal(context.Background(), &Goal{PatientID: uuid.New()}); err == nil { t.Fatal("expected error") }
}

func TestCreateGoal_InvalidStatus(t *testing.T) {
	svc := newTestService()
	if err := svc.CreateGoal(context.Background(), &Goal{PatientID: uuid.New(), Description: "test", LifecycleStatus: "bogus"}); err == nil { t.Fatal("expected error") }
}

func TestCreateGoal_ValidStatuses(t *testing.T) {
	for _, s := range []string{"proposed", "planned", "accepted", "active", "on-hold", "completed", "cancelled", "entered-in-error", "rejected"} {
		svc := newTestService()
		g := &Goal{PatientID: uuid.New(), Description: "test", LifecycleStatus: s}
		if err := svc.CreateGoal(context.Background(), g); err != nil { t.Errorf("status %q should be valid: %v", s, err) }
	}
}

func TestGetGoal(t *testing.T) {
	svc := newTestService()
	g := &Goal{PatientID: uuid.New(), Description: "test"}
	svc.CreateGoal(context.Background(), g)
	got, err := svc.GetGoal(context.Background(), g.ID)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if got.ID != g.ID { t.Error("ID mismatch") }
}

func TestGetGoal_NotFound(t *testing.T) {
	svc := newTestService()
	if _, err := svc.GetGoal(context.Background(), uuid.New()); err == nil { t.Fatal("expected error") }
}

func TestDeleteGoal(t *testing.T) {
	svc := newTestService()
	g := &Goal{PatientID: uuid.New(), Description: "test"}
	svc.CreateGoal(context.Background(), g)
	if err := svc.DeleteGoal(context.Background(), g.ID); err != nil { t.Fatalf("unexpected error: %v", err) }
	if _, err := svc.GetGoal(context.Background(), g.ID); err == nil { t.Fatal("expected error after delete") }
}

func TestListGoalsByPatient(t *testing.T) {
	svc := newTestService()
	pid := uuid.New()
	svc.CreateGoal(context.Background(), &Goal{PatientID: pid, Description: "G1"})
	svc.CreateGoal(context.Background(), &Goal{PatientID: pid, Description: "G2"})
	svc.CreateGoal(context.Background(), &Goal{PatientID: uuid.New(), Description: "G3"})
	items, total, err := svc.ListGoalsByPatient(context.Background(), pid, 10, 0)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if total != 2 || len(items) != 2 { t.Errorf("expected 2 goals, got %d", total) }
}
