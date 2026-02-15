package workflow

import (
	"context"

	"github.com/google/uuid"
)

type ActivityDefinitionRepository interface {
	Create(ctx context.Context, a *ActivityDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*ActivityDefinition, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*ActivityDefinition, error)
	Update(ctx context.Context, a *ActivityDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*ActivityDefinition, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*ActivityDefinition, int, error)
}

type RequestGroupRepository interface {
	Create(ctx context.Context, rg *RequestGroup) error
	GetByID(ctx context.Context, id uuid.UUID) (*RequestGroup, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*RequestGroup, error)
	Update(ctx context.Context, rg *RequestGroup) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*RequestGroup, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*RequestGroup, int, error)
	// Actions
	AddAction(ctx context.Context, a *RequestGroupAction) error
	GetActions(ctx context.Context, requestGroupID uuid.UUID) ([]*RequestGroupAction, error)
}

type GuidanceResponseRepository interface {
	Create(ctx context.Context, gr *GuidanceResponse) error
	GetByID(ctx context.Context, id uuid.UUID) (*GuidanceResponse, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*GuidanceResponse, error)
	Update(ctx context.Context, gr *GuidanceResponse) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*GuidanceResponse, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*GuidanceResponse, int, error)
}
