package substance

import (
	"context"
)

type SubstanceRepository interface {
	Create(ctx context.Context, s *Substance) error
	GetByID(ctx context.Context, id string) (*Substance, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Substance, error)
	Update(ctx context.Context, s *Substance) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Substance, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Substance, int, error)
}
