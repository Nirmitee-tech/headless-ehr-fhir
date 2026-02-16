package structuredefinition

import (
	"context"

	"github.com/google/uuid"
)

type StructureDefinitionRepository interface {
	Create(ctx context.Context, s *StructureDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*StructureDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*StructureDefinition, error)
	Update(ctx context.Context, s *StructureDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*StructureDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*StructureDefinition, int, error)
}
