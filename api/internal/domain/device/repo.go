package device

import (
	"context"

	"github.com/google/uuid"
)

type DeviceRepository interface {
	Create(ctx context.Context, d *Device) error
	GetByID(ctx context.Context, id uuid.UUID) (*Device, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Device, error)
	Update(ctx context.Context, d *Device) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Device, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*Device, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Device, int, error)
}
