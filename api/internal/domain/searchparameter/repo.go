package searchparameter

import (
	"context"

	"github.com/google/uuid"
)

type SearchParameterRepository interface {
	Create(ctx context.Context, s *SearchParameter) error
	GetByID(ctx context.Context, id uuid.UUID) (*SearchParameter, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*SearchParameter, error)
	Update(ctx context.Context, s *SearchParameter) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SearchParameter, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SearchParameter, int, error)
}
