package molecularsequence

import (
	"context"

	"github.com/google/uuid"
)

type MolecularSequenceRepository interface {
	Create(ctx context.Context, m *MolecularSequence) error
	GetByID(ctx context.Context, id uuid.UUID) (*MolecularSequence, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MolecularSequence, error)
	Update(ctx context.Context, m *MolecularSequence) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MolecularSequence, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MolecularSequence, int, error)
}
