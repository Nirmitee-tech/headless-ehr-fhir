package evidence

import (
	"context"

	"github.com/google/uuid"
)

type EvidenceRepository interface {
	Create(ctx context.Context, e *Evidence) error
	GetByID(ctx context.Context, id uuid.UUID) (*Evidence, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Evidence, error)
	Update(ctx context.Context, e *Evidence) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Evidence, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Evidence, int, error)
}
