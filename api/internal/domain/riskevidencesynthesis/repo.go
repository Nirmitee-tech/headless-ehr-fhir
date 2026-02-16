package riskevidencesynthesis

import (
	"context"

	"github.com/google/uuid"
)

type RiskEvidenceSynthesisRepository interface {
	Create(ctx context.Context, e *RiskEvidenceSynthesis) error
	GetByID(ctx context.Context, id uuid.UUID) (*RiskEvidenceSynthesis, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*RiskEvidenceSynthesis, error)
	Update(ctx context.Context, e *RiskEvidenceSynthesis) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*RiskEvidenceSynthesis, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*RiskEvidenceSynthesis, int, error)
}
