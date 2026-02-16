package observationdefinition

import (
	"context"

	"github.com/google/uuid"
)

type ObservationDefinitionRepository interface {
	Create(ctx context.Context, od *ObservationDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*ObservationDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ObservationDefinition, error)
	Update(ctx context.Context, od *ObservationDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ObservationDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ObservationDefinition, int, error)
}
