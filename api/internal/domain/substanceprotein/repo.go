package substanceprotein

import (
	"context"
	"github.com/google/uuid"
)

type SubstanceProteinRepository interface {
	Create(ctx context.Context, m *SubstanceProtein) error
	GetByID(ctx context.Context, id uuid.UUID) (*SubstanceProtein, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceProtein, error)
	Update(ctx context.Context, m *SubstanceProtein) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SubstanceProtein, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceProtein, int, error)
}
