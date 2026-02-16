package researchelementdefinition

import (
	"context"

	"github.com/google/uuid"
)

type ResearchElementDefinitionRepository interface {
	Create(ctx context.Context, e *ResearchElementDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*ResearchElementDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ResearchElementDefinition, error)
	Update(ctx context.Context, e *ResearchElementDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ResearchElementDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ResearchElementDefinition, int, error)
}
