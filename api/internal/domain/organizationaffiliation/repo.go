package organizationaffiliation

import (
	"context"

	"github.com/google/uuid"
)

type OrganizationAffiliationRepository interface {
	Create(ctx context.Context, o *OrganizationAffiliation) error
	GetByID(ctx context.Context, id uuid.UUID) (*OrganizationAffiliation, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*OrganizationAffiliation, error)
	Update(ctx context.Context, o *OrganizationAffiliation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*OrganizationAffiliation, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*OrganizationAffiliation, int, error)
}
