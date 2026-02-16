package compartmentdefinition

import (
	"context"

	"github.com/google/uuid"
)

type CompartmentDefinitionRepository interface {
	Create(ctx context.Context, cd *CompartmentDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*CompartmentDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*CompartmentDefinition, error)
	Update(ctx context.Context, cd *CompartmentDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*CompartmentDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CompartmentDefinition, int, error)
}
