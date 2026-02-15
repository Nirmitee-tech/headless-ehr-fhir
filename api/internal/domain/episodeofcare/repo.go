package episodeofcare

import (
	"context"

	"github.com/google/uuid"
)

type EpisodeOfCareRepository interface {
	Create(ctx context.Context, e *EpisodeOfCare) error
	GetByID(ctx context.Context, id uuid.UUID) (*EpisodeOfCare, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*EpisodeOfCare, error)
	Update(ctx context.Context, e *EpisodeOfCare) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*EpisodeOfCare, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*EpisodeOfCare, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EpisodeOfCare, int, error)
}
