package bodystructure

import (
	"context"

	"github.com/google/uuid"
)

type BodyStructureRepository interface {
	Create(ctx context.Context, b *BodyStructure) error
	GetByID(ctx context.Context, id uuid.UUID) (*BodyStructure, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*BodyStructure, error)
	Update(ctx context.Context, b *BodyStructure) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*BodyStructure, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*BodyStructure, int, error)
}
