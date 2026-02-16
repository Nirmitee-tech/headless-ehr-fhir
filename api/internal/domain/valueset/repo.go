package valueset

import (
	"context"

	"github.com/google/uuid"
)

type ValueSetRepository interface {
	Create(ctx context.Context, vs *ValueSet) error
	GetByID(ctx context.Context, id uuid.UUID) (*ValueSet, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ValueSet, error)
	Update(ctx context.Context, vs *ValueSet) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ValueSet, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ValueSet, int, error)
}
