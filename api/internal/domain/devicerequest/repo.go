package devicerequest

import (
	"context"

	"github.com/google/uuid"
)

type DeviceRequestRepository interface {
	Create(ctx context.Context, d *DeviceRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*DeviceRequest, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*DeviceRequest, error)
	Update(ctx context.Context, d *DeviceRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*DeviceRequest, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceRequest, int, error)
}
