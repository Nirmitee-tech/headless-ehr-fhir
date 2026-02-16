package medproductinteraction

import (
	"context"
	"github.com/google/uuid"
)

type MedicinalProductInteractionRepository interface {
	Create(ctx context.Context, m *MedicinalProductInteraction) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductInteraction, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductInteraction, error)
	Update(ctx context.Context, m *MedicinalProductInteraction) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicinalProductInteraction, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductInteraction, int, error)
}
