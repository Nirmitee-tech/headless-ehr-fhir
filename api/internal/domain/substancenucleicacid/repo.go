package substancenucleicacid

import (
	"context"
	"github.com/google/uuid"
)

type SubstanceNucleicAcidRepository interface {
	Create(ctx context.Context, m *SubstanceNucleicAcid) error
	GetByID(ctx context.Context, id uuid.UUID) (*SubstanceNucleicAcid, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceNucleicAcid, error)
	Update(ctx context.Context, m *SubstanceNucleicAcid) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SubstanceNucleicAcid, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceNucleicAcid, int, error)
}
