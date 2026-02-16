package immunizationevaluation

import (
	"context"

	"github.com/google/uuid"
)

type ImmunizationEvaluationRepository interface {
	Create(ctx context.Context, ie *ImmunizationEvaluation) error
	GetByID(ctx context.Context, id uuid.UUID) (*ImmunizationEvaluation, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ImmunizationEvaluation, error)
	Update(ctx context.Context, ie *ImmunizationEvaluation) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ImmunizationEvaluation, int, error)
}
