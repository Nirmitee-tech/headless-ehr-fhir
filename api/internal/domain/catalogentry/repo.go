package catalogentry

import (
	"context"

	"github.com/google/uuid"
)

type CatalogEntryRepository interface {
	Create(ctx context.Context, ce *CatalogEntry) error
	GetByID(ctx context.Context, id uuid.UUID) (*CatalogEntry, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*CatalogEntry, error)
	Update(ctx context.Context, ce *CatalogEntry) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*CatalogEntry, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CatalogEntry, int, error)
}
