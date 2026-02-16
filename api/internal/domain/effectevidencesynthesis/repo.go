package effectevidencesynthesis

import (
	"context"

	"github.com/google/uuid"
)

type EffectEvidenceSynthesisRepository interface {
	Create(ctx context.Context, e *EffectEvidenceSynthesis) error
	GetByID(ctx context.Context, id uuid.UUID) (*EffectEvidenceSynthesis, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*EffectEvidenceSynthesis, error)
	Update(ctx context.Context, e *EffectEvidenceSynthesis) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*EffectEvidenceSynthesis, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EffectEvidenceSynthesis, int, error)
}
