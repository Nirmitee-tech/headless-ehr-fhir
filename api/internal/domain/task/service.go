package task

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	tasks TaskRepository
	vt    *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(tasks TaskRepository) *Service {
	return &Service{tasks: tasks}
}

var validTaskStatuses = map[string]bool{
	"draft":            true,
	"requested":        true,
	"received":         true,
	"accepted":         true,
	"rejected":         true,
	"ready":            true,
	"cancelled":        true,
	"in-progress":      true,
	"on-hold":          true,
	"failed":           true,
	"completed":        true,
	"entered-in-error": true,
}

var validTaskIntents = map[string]bool{
	"unknown":        true,
	"proposal":       true,
	"plan":           true,
	"order":          true,
	"original-order": true,
	"reflex-order":   true,
	"filler-order":   true,
	"instance-order": true,
	"option":         true,
}

var validTaskPriorities = map[string]bool{
	"routine": true,
	"urgent":  true,
	"asap":    true,
	"stat":    true,
}

func (s *Service) CreateTask(ctx context.Context, t *Task) error {
	if t.Intent == "" {
		return fmt.Errorf("intent is required")
	}
	if !validTaskIntents[t.Intent] {
		return fmt.Errorf("invalid intent: %s", t.Intent)
	}
	if t.Status == "" {
		t.Status = "draft"
	}
	if !validTaskStatuses[t.Status] {
		return fmt.Errorf("invalid status: %s", t.Status)
	}
	if t.Priority != nil && !validTaskPriorities[*t.Priority] {
		return fmt.Errorf("invalid priority: %s", *t.Priority)
	}
	if err := s.tasks.Create(ctx, t); err != nil {
		return err
	}
	t.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Task", t.FHIRID, t.ToFHIR())
	}
	return nil
}

func (s *Service) GetTask(ctx context.Context, id uuid.UUID) (*Task, error) {
	return s.tasks.GetByID(ctx, id)
}

func (s *Service) GetTaskByFHIRID(ctx context.Context, fhirID string) (*Task, error) {
	return s.tasks.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateTask(ctx context.Context, t *Task) error {
	if t.Status != "" && !validTaskStatuses[t.Status] {
		return fmt.Errorf("invalid status: %s", t.Status)
	}
	if t.Priority != nil && !validTaskPriorities[*t.Priority] {
		return fmt.Errorf("invalid priority: %s", *t.Priority)
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Task", t.FHIRID, t.VersionID, t.ToFHIR())
		if err == nil {
			t.VersionID = newVer
		}
	}
	return s.tasks.Update(ctx, t)
}

func (s *Service) DeleteTask(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		t, err := s.tasks.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Task", t.FHIRID, t.VersionID)
		}
	}
	return s.tasks.Delete(ctx, id)
}

func (s *Service) ListTasksByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Task, int, error) {
	return s.tasks.ListByPatient(ctx, patientID, limit, offset)
}

func (s *Service) ListTasksByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Task, int, error) {
	return s.tasks.ListByOwner(ctx, ownerID, limit, offset)
}

func (s *Service) SearchTasks(ctx context.Context, params map[string]string, limit, offset int) ([]*Task, int, error) {
	return s.tasks.Search(ctx, params, limit, offset)
}
