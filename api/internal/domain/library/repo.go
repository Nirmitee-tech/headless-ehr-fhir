package library

import (
	"context"

	"github.com/google/uuid"
)

type LibraryRepository interface {
	Create(ctx context.Context, l *Library) error
	GetByID(ctx context.Context, id uuid.UUID) (*Library, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Library, error)
	Update(ctx context.Context, l *Library) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Library, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Library, int, error)
}
