package medproductindication

import (
	"context"
	"github.com/google/uuid"
)

type MedicinalProductIndicationRepository interface {
	Create(ctx context.Context, m *MedicinalProductIndication) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductIndication, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductIndication, error)
	Update(ctx context.Context, m *MedicinalProductIndication) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicinalProductIndication, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductIndication, int, error)
}
