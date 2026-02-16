package researchdefinition

import (
	"context"

	"github.com/google/uuid"
)

type ResearchDefinitionRepository interface {
	Create(ctx context.Context, e *ResearchDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*ResearchDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ResearchDefinition, error)
	Update(ctx context.Context, e *ResearchDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ResearchDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ResearchDefinition, int, error)
}
