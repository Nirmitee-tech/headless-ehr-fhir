package eventdefinition

import (
	"context"

	"github.com/google/uuid"
)

type EventDefinitionRepository interface {
	Create(ctx context.Context, e *EventDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*EventDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*EventDefinition, error)
	Update(ctx context.Context, e *EventDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*EventDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*EventDefinition, int, error)
}
