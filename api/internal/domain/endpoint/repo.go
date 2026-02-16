package endpoint

import (
	"context"

	"github.com/google/uuid"
)

type EndpointRepository interface {
	Create(ctx context.Context, e *Endpoint) error
	GetByID(ctx context.Context, id uuid.UUID) (*Endpoint, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Endpoint, error)
	Update(ctx context.Context, e *Endpoint) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Endpoint, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Endpoint, int, error)
}
