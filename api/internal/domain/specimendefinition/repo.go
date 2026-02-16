package specimendefinition

import (
	"context"

	"github.com/google/uuid"
)

type SpecimenDefinitionRepository interface {
	Create(ctx context.Context, s *SpecimenDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*SpecimenDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*SpecimenDefinition, error)
	Update(ctx context.Context, s *SpecimenDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SpecimenDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SpecimenDefinition, int, error)
}
