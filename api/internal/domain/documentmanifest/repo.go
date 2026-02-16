package documentmanifest

import (
	"context"

	"github.com/google/uuid"
)

type DocumentManifestRepository interface {
	Create(ctx context.Context, d *DocumentManifest) error
	GetByID(ctx context.Context, id uuid.UUID) (*DocumentManifest, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*DocumentManifest, error)
	Update(ctx context.Context, d *DocumentManifest) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*DocumentManifest, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DocumentManifest, int, error)
}
