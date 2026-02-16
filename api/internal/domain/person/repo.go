package person

import (
	"context"

	"github.com/google/uuid"
)

type PersonRepository interface {
	Create(ctx context.Context, p *Person) error
	GetByID(ctx context.Context, id uuid.UUID) (*Person, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Person, error)
	Update(ctx context.Context, p *Person) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Person, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Person, int, error)
}
