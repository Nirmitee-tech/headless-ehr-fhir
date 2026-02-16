package verificationresult

import (
	"context"

	"github.com/google/uuid"
)

type VerificationResultRepository interface {
	Create(ctx context.Context, v *VerificationResult) error
	GetByID(ctx context.Context, id uuid.UUID) (*VerificationResult, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*VerificationResult, error)
	Update(ctx context.Context, v *VerificationResult) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*VerificationResult, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*VerificationResult, int, error)
}
