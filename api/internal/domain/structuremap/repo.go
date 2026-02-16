package structuremap

import (
	"context"

	"github.com/google/uuid"
)

type StructureMapRepository interface {
	Create(ctx context.Context, sm *StructureMap) error
	GetByID(ctx context.Context, id uuid.UUID) (*StructureMap, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*StructureMap, error)
	Update(ctx context.Context, sm *StructureMap) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*StructureMap, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*StructureMap, int, error)
}
