package substancepolymer

import (
	"context"
	"github.com/google/uuid"
)

type SubstancePolymerRepository interface {
	Create(ctx context.Context, m *SubstancePolymer) error
	GetByID(ctx context.Context, id uuid.UUID) (*SubstancePolymer, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*SubstancePolymer, error)
	Update(ctx context.Context, m *SubstancePolymer) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SubstancePolymer, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstancePolymer, int, error)
}
