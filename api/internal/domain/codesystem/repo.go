package codesystem

import (
	"context"

	"github.com/google/uuid"
)

type CodeSystemRepository interface {
	Create(ctx context.Context, cs *CodeSystem) error
	GetByID(ctx context.Context, id uuid.UUID) (*CodeSystem, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*CodeSystem, error)
	Update(ctx context.Context, cs *CodeSystem) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*CodeSystem, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CodeSystem, int, error)
}
