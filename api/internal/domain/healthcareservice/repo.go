package healthcareservice

import (
	"context"

	"github.com/google/uuid"
)

type HealthcareServiceRepository interface {
	Create(ctx context.Context, hs *HealthcareService) error
	GetByID(ctx context.Context, id uuid.UUID) (*HealthcareService, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*HealthcareService, error)
	Update(ctx context.Context, hs *HealthcareService) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*HealthcareService, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*HealthcareService, int, error)
}
