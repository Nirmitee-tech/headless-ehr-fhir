package substancesourcematerial

import (
	"context"
	"github.com/google/uuid"
)

type SubstanceSourceMaterialRepository interface {
	Create(ctx context.Context, m *SubstanceSourceMaterial) error
	GetByID(ctx context.Context, id uuid.UUID) (*SubstanceSourceMaterial, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceSourceMaterial, error)
	Update(ctx context.Context, m *SubstanceSourceMaterial) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SubstanceSourceMaterial, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceSourceMaterial, int, error)
}
