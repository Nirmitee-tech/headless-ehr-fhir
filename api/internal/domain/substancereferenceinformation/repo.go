package substancereferenceinformation

import (
	"context"
	"github.com/google/uuid"
)

type SubstanceReferenceInformationRepository interface {
	Create(ctx context.Context, m *SubstanceReferenceInformation) error
	GetByID(ctx context.Context, id uuid.UUID) (*SubstanceReferenceInformation, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*SubstanceReferenceInformation, error)
	Update(ctx context.Context, m *SubstanceReferenceInformation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SubstanceReferenceInformation, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SubstanceReferenceInformation, int, error)
}
