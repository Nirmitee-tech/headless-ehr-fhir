package substancespecification

import (
	"context"

	"github.com/google/uuid"
)

type SubstanceSpecificationRepository interface {
	Create(ctx context.Context, s *SubstanceSpecification) error
	GetByID(ctx context.Context, id uuid.UUID) (*SubstanceSpecification, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceSpecification, error)
	Update(ctx context.Context, s *SubstanceSpecification) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SubstanceSpecification, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceSpecification, int, error)
}
