package medproductundesirableeffect

import (
	"context"
	"github.com/google/uuid"
)

type MedicinalProductUndesirableEffectRepository interface {
	Create(ctx context.Context, m *MedicinalProductUndesirableEffect) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductUndesirableEffect, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductUndesirableEffect, error)
	Update(ctx context.Context, m *MedicinalProductUndesirableEffect) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicinalProductUndesirableEffect, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductUndesirableEffect, int, error)
}
