package task

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// -- Mock Repository --

type mockTaskRepo struct {
	store map[uuid.UUID]*Task
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{store: make(map[uuid.UUID]*Task)}
}

func (m *mockTaskRepo) Create(_ context.Context, t *Task) error {
	t.ID = uuid.New()
	if t.FHIRID == "" {
		t.FHIRID = t.ID.String()
	}
	m.store[t.ID] = t
	return nil
}

func (m *mockTaskRepo) GetByID(_ context.Context, id uuid.UUID) (*Task, error) {
	t, ok := m.store[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return t, nil
}

func (m *mockTaskRepo) GetByFHIRID(_ context.Context, fhirID string) (*Task, error) {
	for _, t := range m.store {
		if t.FHIRID == fhirID {
			return t, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockTaskRepo) Update(_ context.Context, t *Task) error {
	if _, ok := m.store[t.ID]; !ok {
		return fmt.Errorf("not found")
	}
	m.store[t.ID] = t
	return nil
}

func (m *mockTaskRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.store, id)
	return nil
}

func (m *mockTaskRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*Task, int, error) {
	var r []*Task
	for _, t := range m.store {
		if t.ForPatientID == patientID {
			r = append(r, t)
		}
	}
	return r, len(r), nil
}

func (m *mockTaskRepo) ListByOwner(_ context.Context, ownerID uuid.UUID, limit, offset int) ([]*Task, int, error) {
	var r []*Task
	for _, t := range m.store {
		if t.OwnerID != nil && *t.OwnerID == ownerID {
			r = append(r, t)
		}
	}
	return r, len(r), nil
}

func (m *mockTaskRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*Task, int, error) {
	var r []*Task
	for _, t := range m.store {
		r = append(r, t)
	}
	return r, len(r), nil
}

func newTestService() *Service {
	return NewService(newMockTaskRepo())
}

// -- Service Tests --

func TestCreateTask_Success(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()
	tk := &Task{
		ForPatientID: patientID,
		Intent:       "order",
		Status:       "requested",
	}
	if err := svc.CreateTask(context.Background(), tk); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tk.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
	if tk.FHIRID == "" {
		t.Error("expected FHIRID to be set")
	}
	if tk.Status != "requested" {
		t.Errorf("expected status 'requested', got %q", tk.Status)
	}
}

func TestCreateTask_DefaultStatus(t *testing.T) {
	svc := newTestService()
	tk := &Task{
		ForPatientID: uuid.New(),
		Intent:       "order",
	}
	if err := svc.CreateTask(context.Background(), tk); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tk.Status != "draft" {
		t.Errorf("expected default status 'draft', got %q", tk.Status)
	}
}

func TestCreateTask_MissingStatus(t *testing.T) {
	// When status is empty, it should default to "draft" (not error)
	svc := newTestService()
	tk := &Task{
		ForPatientID: uuid.New(),
		Intent:       "order",
		Status:       "",
	}
	if err := svc.CreateTask(context.Background(), tk); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tk.Status != "draft" {
		t.Errorf("expected default status 'draft', got %q", tk.Status)
	}
}

func TestCreateTask_InvalidStatus(t *testing.T) {
	svc := newTestService()
	tk := &Task{
		ForPatientID: uuid.New(),
		Intent:       "order",
		Status:       "bogus",
	}
	if err := svc.CreateTask(context.Background(), tk); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestCreateTask_MissingIntent(t *testing.T) {
	svc := newTestService()
	tk := &Task{
		ForPatientID: uuid.New(),
		Status:       "draft",
	}
	if err := svc.CreateTask(context.Background(), tk); err == nil {
		t.Fatal("expected error for missing intent")
	}
}

func TestCreateTask_InvalidIntent(t *testing.T) {
	svc := newTestService()
	tk := &Task{
		ForPatientID: uuid.New(),
		Intent:       "bogus",
		Status:       "draft",
	}
	if err := svc.CreateTask(context.Background(), tk); err == nil {
		t.Fatal("expected error for invalid intent")
	}
}

func TestCreateTask_AllValidStatuses(t *testing.T) {
	validStatuses := []string{
		"draft", "requested", "received", "accepted", "rejected",
		"ready", "cancelled", "in-progress", "on-hold", "failed",
		"completed", "entered-in-error",
	}
	for _, s := range validStatuses {
		svc := newTestService()
		tk := &Task{
			ForPatientID: uuid.New(),
			Intent:       "order",
			Status:       s,
		}
		if err := svc.CreateTask(context.Background(), tk); err != nil {
			t.Errorf("status %q should be valid: %v", s, err)
		}
	}
}

func TestCreateTask_AllValidIntents(t *testing.T) {
	validIntents := []string{
		"unknown", "proposal", "plan", "order", "original-order",
		"reflex-order", "filler-order", "instance-order", "option",
	}
	for _, intent := range validIntents {
		svc := newTestService()
		tk := &Task{
			ForPatientID: uuid.New(),
			Intent:       intent,
			Status:       "draft",
		}
		if err := svc.CreateTask(context.Background(), tk); err != nil {
			t.Errorf("intent %q should be valid: %v", intent, err)
		}
	}
}

func TestCreateTask_ValidPriorities(t *testing.T) {
	validPriorities := []string{"routine", "urgent", "asap", "stat"}
	for _, p := range validPriorities {
		svc := newTestService()
		tk := &Task{
			ForPatientID: uuid.New(),
			Intent:       "order",
			Status:       "draft",
			Priority:     &p,
		}
		if err := svc.CreateTask(context.Background(), tk); err != nil {
			t.Errorf("priority %q should be valid: %v", p, err)
		}
	}
}

func TestCreateTask_InvalidPriority(t *testing.T) {
	svc := newTestService()
	bad := "critical"
	tk := &Task{
		ForPatientID: uuid.New(),
		Intent:       "order",
		Status:       "draft",
		Priority:     &bad,
	}
	if err := svc.CreateTask(context.Background(), tk); err == nil {
		t.Fatal("expected error for invalid priority")
	}
}

func TestGetTask_Success(t *testing.T) {
	svc := newTestService()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	svc.CreateTask(context.Background(), tk)

	got, err := svc.GetTask(context.Background(), tk.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != tk.ID {
		t.Errorf("ID mismatch: expected %v, got %v", tk.ID, got.ID)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetTask(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
}

func TestGetTaskByFHIRID_Success(t *testing.T) {
	svc := newTestService()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	svc.CreateTask(context.Background(), tk)

	got, err := svc.GetTaskByFHIRID(context.Background(), tk.FHIRID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.FHIRID != tk.FHIRID {
		t.Errorf("FHIRID mismatch")
	}
}

func TestGetTaskByFHIRID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetTaskByFHIRID(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent FHIR ID")
	}
}

func TestUpdateTask_Success(t *testing.T) {
	svc := newTestService()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	svc.CreateTask(context.Background(), tk)

	tk.Status = "requested"
	if err := svc.UpdateTask(context.Background(), tk); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetTask(context.Background(), tk.ID)
	if got.Status != "requested" {
		t.Errorf("expected status 'requested', got %q", got.Status)
	}
}

func TestUpdateTask_InvalidStatus(t *testing.T) {
	svc := newTestService()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	svc.CreateTask(context.Background(), tk)

	tk.Status = "invalid-status"
	if err := svc.UpdateTask(context.Background(), tk); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestUpdateTask_StatusTransition(t *testing.T) {
	// Valid transitions: draft -> requested -> accepted -> in-progress -> completed
	svc := newTestService()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	svc.CreateTask(context.Background(), tk)

	transitions := []string{"requested", "accepted", "in-progress", "completed"}
	for _, next := range transitions {
		tk.Status = next
		if err := svc.UpdateTask(context.Background(), tk); err != nil {
			t.Errorf("transition to %q should be valid: %v", next, err)
		}
	}
}

func TestUpdateTask_StatusTransition_ToFailed(t *testing.T) {
	svc := newTestService()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "in-progress"}
	svc.CreateTask(context.Background(), tk)

	tk.Status = "failed"
	if err := svc.UpdateTask(context.Background(), tk); err != nil {
		t.Fatalf("transition to 'failed' should be valid: %v", err)
	}
}

func TestUpdateTask_StatusTransition_ToCancelled(t *testing.T) {
	svc := newTestService()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "requested"}
	svc.CreateTask(context.Background(), tk)

	tk.Status = "cancelled"
	if err := svc.UpdateTask(context.Background(), tk); err != nil {
		t.Fatalf("transition to 'cancelled' should be valid: %v", err)
	}
}

func TestUpdateTask_StatusTransition_ToOnHold(t *testing.T) {
	svc := newTestService()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "in-progress"}
	svc.CreateTask(context.Background(), tk)

	tk.Status = "on-hold"
	if err := svc.UpdateTask(context.Background(), tk); err != nil {
		t.Fatalf("transition to 'on-hold' should be valid: %v", err)
	}
}

func TestDeleteTask_Success(t *testing.T) {
	svc := newTestService()
	tk := &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"}
	svc.CreateTask(context.Background(), tk)

	if err := svc.DeleteTask(context.Background(), tk.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.GetTask(context.Background(), tk.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestListTasksByPatient(t *testing.T) {
	svc := newTestService()
	patientID := uuid.New()

	svc.CreateTask(context.Background(), &Task{ForPatientID: patientID, Intent: "order", Status: "draft"})
	svc.CreateTask(context.Background(), &Task{ForPatientID: patientID, Intent: "order", Status: "requested"})
	svc.CreateTask(context.Background(), &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"})

	items, total, err := svc.ListTasksByPatient(context.Background(), patientID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 tasks for patient, got %d", total)
	}
}

func TestListTasksByOwner(t *testing.T) {
	svc := newTestService()
	ownerID := uuid.New()
	otherOwner := uuid.New()

	svc.CreateTask(context.Background(), &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft", OwnerID: &ownerID})
	svc.CreateTask(context.Background(), &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft", OwnerID: &ownerID})
	svc.CreateTask(context.Background(), &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft", OwnerID: &otherOwner})

	items, total, err := svc.ListTasksByOwner(context.Background(), ownerID, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Errorf("expected 2 tasks for owner, got %d", total)
	}
}

func TestSearchTasks(t *testing.T) {
	svc := newTestService()
	svc.CreateTask(context.Background(), &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"})
	svc.CreateTask(context.Background(), &Task{ForPatientID: uuid.New(), Intent: "order", Status: "completed"})

	params := map[string]string{"status": "draft"}
	items, total, err := svc.SearchTasks(context.Background(), params, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The mock returns all items regardless of params, so just ensure no error
	if total < 1 || len(items) < 1 {
		t.Errorf("expected at least 1 task in search results, got %d", total)
	}
}

func TestSearchTasks_EmptyParams(t *testing.T) {
	svc := newTestService()
	svc.CreateTask(context.Background(), &Task{ForPatientID: uuid.New(), Intent: "order", Status: "draft"})

	items, total, err := svc.SearchTasks(context.Background(), nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Errorf("expected 1 task, got %d", total)
	}
}
