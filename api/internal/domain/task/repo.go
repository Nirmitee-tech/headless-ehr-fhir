package task

import (
	"context"

	"github.com/google/uuid"
)

type TaskRepository interface {
	Create(ctx context.Context, t *Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*Task, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Task, error)
	Update(ctx context.Context, t *Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Task, int, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*Task, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Task, int, error)
}
