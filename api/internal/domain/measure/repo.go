package measure

import (
	"context"

	"github.com/google/uuid"
)

type MeasureRepository interface {
	Create(ctx context.Context, m *Measure) error
	GetByID(ctx context.Context, id uuid.UUID) (*Measure, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Measure, error)
	Update(ctx context.Context, m *Measure) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Measure, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Measure, int, error)
}
