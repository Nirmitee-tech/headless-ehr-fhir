package conceptmap

import (
	"context"

	"github.com/google/uuid"
)

type ConceptMapRepository interface {
	Create(ctx context.Context, cm *ConceptMap) error
	GetByID(ctx context.Context, id uuid.UUID) (*ConceptMap, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ConceptMap, error)
	Update(ctx context.Context, cm *ConceptMap) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ConceptMap, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ConceptMap, int, error)
}
