package medproductcontraindication

import (
	"context"
	"github.com/google/uuid"
)

type MedicinalProductContraindicationRepository interface {
	Create(ctx context.Context, m *MedicinalProductContraindication) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductContraindication, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductContraindication, error)
	Update(ctx context.Context, m *MedicinalProductContraindication) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicinalProductContraindication, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductContraindication, int, error)
}
