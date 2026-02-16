package media

import (
	"context"

	"github.com/google/uuid"
)

type MediaRepository interface {
	Create(ctx context.Context, m *Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*Media, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Media, error)
	Update(ctx context.Context, m *Media) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Media, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Media, int, error)
}
