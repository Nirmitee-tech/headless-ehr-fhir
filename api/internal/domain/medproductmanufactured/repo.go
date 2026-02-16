package medproductmanufactured

import (
	"context"
	"github.com/google/uuid"
)

type MedicinalProductManufacturedRepository interface {
	Create(ctx context.Context, m *MedicinalProductManufactured) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductManufactured, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductManufactured, error)
	Update(ctx context.Context, m *MedicinalProductManufactured) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicinalProductManufactured, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductManufactured, int, error)
}
