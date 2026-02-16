package basic

import (
	"context"

	"github.com/google/uuid"
)

type BasicRepository interface {
	Create(ctx context.Context, b *Basic) error
	GetByID(ctx context.Context, id uuid.UUID) (*Basic, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Basic, error)
	Update(ctx context.Context, b *Basic) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Basic, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Basic, int, error)
}
