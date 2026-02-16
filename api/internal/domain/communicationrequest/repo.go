package communicationrequest

import (
	"context"

	"github.com/google/uuid"
)

type CommunicationRequestRepository interface {
	Create(ctx context.Context, cr *CommunicationRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*CommunicationRequest, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*CommunicationRequest, error)
	Update(ctx context.Context, cr *CommunicationRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*CommunicationRequest, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*CommunicationRequest, int, error)
}
