package supply

import (
	"context"

	"github.com/google/uuid"
)

type SupplyRequestRepository interface {
	Create(ctx context.Context, s *SupplyRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*SupplyRequest, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*SupplyRequest, error)
	Update(ctx context.Context, s *SupplyRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SupplyRequest, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SupplyRequest, int, error)
}

type SupplyDeliveryRepository interface {
	Create(ctx context.Context, s *SupplyDelivery) error
	GetByID(ctx context.Context, id uuid.UUID) (*SupplyDelivery, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*SupplyDelivery, error)
	Update(ctx context.Context, s *SupplyDelivery) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SupplyDelivery, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*SupplyDelivery, int, error)
}
