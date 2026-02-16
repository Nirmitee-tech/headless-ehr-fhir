package linkage

import (
	"context"

	"github.com/google/uuid"
)

type LinkageRepository interface {
	Create(ctx context.Context, l *Linkage) error
	GetByID(ctx context.Context, id uuid.UUID) (*Linkage, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Linkage, error)
	Update(ctx context.Context, l *Linkage) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Linkage, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Linkage, int, error)
}
