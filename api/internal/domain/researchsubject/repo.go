package researchsubject

import (
	"context"

	"github.com/google/uuid"
)

type ResearchSubjectRepository interface {
	Create(ctx context.Context, r *ResearchSubject) error
	GetByID(ctx context.Context, id uuid.UUID) (*ResearchSubject, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ResearchSubject, error)
	Update(ctx context.Context, r *ResearchSubject) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ResearchSubject, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ResearchSubject, int, error)
}
