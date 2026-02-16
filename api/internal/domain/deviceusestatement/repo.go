package deviceusestatement

import (
	"context"

	"github.com/google/uuid"
)

type DeviceUseStatementRepository interface {
	Create(ctx context.Context, d *DeviceUseStatement) error
	GetByID(ctx context.Context, id uuid.UUID) (*DeviceUseStatement, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*DeviceUseStatement, error)
	Update(ctx context.Context, d *DeviceUseStatement) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*DeviceUseStatement, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*DeviceUseStatement, int, error)
}
