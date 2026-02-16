package medicinalproduct

import (
	"context"
	"github.com/google/uuid"
)

type MedicinalProductRepository interface {
	Create(ctx context.Context, m *MedicinalProduct) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProduct, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProduct, error)
	Update(ctx context.Context, m *MedicinalProduct) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicinalProduct, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProduct, int, error)
}
