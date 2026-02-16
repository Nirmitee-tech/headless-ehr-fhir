package devicedefinition

import (
	"context"

	"github.com/google/uuid"
)

type DeviceDefinitionRepository interface {
	Create(ctx context.Context, d *DeviceDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*DeviceDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*DeviceDefinition, error)
	Update(ctx context.Context, d *DeviceDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*DeviceDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceDefinition, int, error)
}
