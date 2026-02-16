package medproductpackaged

import (
	"context"
	"github.com/google/uuid"
)

type MedicinalProductPackagedRepository interface {
	Create(ctx context.Context, m *MedicinalProductPackaged) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductPackaged, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductPackaged, error)
	Update(ctx context.Context, m *MedicinalProductPackaged) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicinalProductPackaged, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductPackaged, int, error)
}
