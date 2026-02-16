package medproductpharmaceutical

import (
	"context"
	"github.com/google/uuid"
)

type MedicinalProductPharmaceuticalRepository interface {
	Create(ctx context.Context, m *MedicinalProductPharmaceutical) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductPharmaceutical, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductPharmaceutical, error)
	Update(ctx context.Context, m *MedicinalProductPharmaceutical) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicinalProductPharmaceutical, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductPharmaceutical, int, error)
}
