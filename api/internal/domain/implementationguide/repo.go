package implementationguide

import (
	"context"

	"github.com/google/uuid"
)

type ImplementationGuideRepository interface {
	Create(ctx context.Context, ig *ImplementationGuide) error
	GetByID(ctx context.Context, id uuid.UUID) (*ImplementationGuide, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ImplementationGuide, error)
	Update(ctx context.Context, ig *ImplementationGuide) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ImplementationGuide, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ImplementationGuide, int, error)
}
