package biologicallyderivedproduct

import (
	"context"

	"github.com/google/uuid"
)

type BiologicallyDerivedProductRepository interface {
	Create(ctx context.Context, b *BiologicallyDerivedProduct) error
	GetByID(ctx context.Context, id uuid.UUID) (*BiologicallyDerivedProduct, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*BiologicallyDerivedProduct, error)
	Update(ctx context.Context, b *BiologicallyDerivedProduct) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*BiologicallyDerivedProduct, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*BiologicallyDerivedProduct, int, error)
}
