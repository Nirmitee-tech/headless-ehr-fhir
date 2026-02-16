package evidencevariable

import (
	"context"

	"github.com/google/uuid"
)

type EvidenceVariableRepository interface {
	Create(ctx context.Context, e *EvidenceVariable) error
	GetByID(ctx context.Context, id uuid.UUID) (*EvidenceVariable, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*EvidenceVariable, error)
	Update(ctx context.Context, e *EvidenceVariable) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*EvidenceVariable, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EvidenceVariable, int, error)
}
