package devicemetric

import (
	"context"

	"github.com/google/uuid"
)

type DeviceMetricRepository interface {
	Create(ctx context.Context, m *DeviceMetric) error
	GetByID(ctx context.Context, id uuid.UUID) (*DeviceMetric, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*DeviceMetric, error)
	Update(ctx context.Context, m *DeviceMetric) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*DeviceMetric, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceMetric, int, error)
}
