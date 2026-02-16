package medproductauthorization

import (
	"context"
	"github.com/google/uuid"
)

type MedicinalProductAuthorizationRepository interface {
	Create(ctx context.Context, m *MedicinalProductAuthorization) error
	GetByID(ctx context.Context, id uuid.UUID) (*MedicinalProductAuthorization, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*MedicinalProductAuthorization, error)
	Update(ctx context.Context, m *MedicinalProductAuthorization) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*MedicinalProductAuthorization, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*MedicinalProductAuthorization, int, error)
}
