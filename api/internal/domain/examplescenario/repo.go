package examplescenario

import (
	"context"

	"github.com/google/uuid"
)

type ExampleScenarioRepository interface {
	Create(ctx context.Context, e *ExampleScenario) error
	GetByID(ctx context.Context, id uuid.UUID) (*ExampleScenario, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ExampleScenario, error)
	Update(ctx context.Context, e *ExampleScenario) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ExampleScenario, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ExampleScenario, int, error)
}
