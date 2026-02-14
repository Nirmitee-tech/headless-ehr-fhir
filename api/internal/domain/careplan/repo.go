package careplan

import (
	"context"

	"github.com/google/uuid"
)

type CarePlanRepository interface {
	Create(ctx context.Context, cp *CarePlan) error
	GetByID(ctx context.Context, id uuid.UUID) (*CarePlan, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*CarePlan, error)
	Update(ctx context.Context, cp *CarePlan) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*CarePlan, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*CarePlan, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CarePlan, int, error)
	// Activities
	AddActivity(ctx context.Context, a *CarePlanActivity) error
	GetActivities(ctx context.Context, carePlanID uuid.UUID) ([]*CarePlanActivity, error)
}

type GoalRepository interface {
	Create(ctx context.Context, g *Goal) error
	GetByID(ctx context.Context, id uuid.UUID) (*Goal, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Goal, error)
	Update(ctx context.Context, g *Goal) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Goal, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Goal, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Goal, int, error)
}
