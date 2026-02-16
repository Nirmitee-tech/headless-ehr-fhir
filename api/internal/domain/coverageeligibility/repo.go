package coverageeligibility

import (
	"context"

	"github.com/google/uuid"
)

type CoverageEligibilityRequestRepository interface {
	Create(ctx context.Context, r *CoverageEligibilityRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*CoverageEligibilityRequest, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*CoverageEligibilityRequest, error)
	Update(ctx context.Context, r *CoverageEligibilityRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*CoverageEligibilityRequest, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CoverageEligibilityRequest, int, error)
}

type CoverageEligibilityResponseRepository interface {
	Create(ctx context.Context, r *CoverageEligibilityResponse) error
	GetByID(ctx context.Context, id uuid.UUID) (*CoverageEligibilityResponse, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*CoverageEligibilityResponse, error)
	Update(ctx context.Context, r *CoverageEligibilityResponse) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*CoverageEligibilityResponse, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CoverageEligibilityResponse, int, error)
}
