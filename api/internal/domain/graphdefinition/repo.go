package graphdefinition

import (
	"context"

	"github.com/google/uuid"
)

type GraphDefinitionRepository interface {
	Create(ctx context.Context, g *GraphDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*GraphDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*GraphDefinition, error)
	Update(ctx context.Context, g *GraphDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*GraphDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*GraphDefinition, int, error)
}
