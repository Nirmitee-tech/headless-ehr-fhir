package admin

import (
	"context"

	"github.com/google/uuid"
)

// OrganizationRepository defines the persistence interface for organizations.
type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Organization, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*Organization, int, error)
}

// DepartmentRepository defines the persistence interface for departments.
type DepartmentRepository interface {
	Create(ctx context.Context, dept *Department) error
	GetByID(ctx context.Context, id uuid.UUID) (*Department, error)
	Update(ctx context.Context, dept *Department) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*Department, int, error)
}

// LocationRepository defines the persistence interface for locations.
type LocationRepository interface {
	Create(ctx context.Context, loc *Location) error
	GetByID(ctx context.Context, id uuid.UUID) (*Location, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*Location, error)
	Update(ctx context.Context, loc *Location) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Location, int, error)
}

// SystemUserRepository defines the persistence interface for system users.
type SystemUserRepository interface {
	Create(ctx context.Context, user *SystemUser) error
	GetByID(ctx context.Context, id uuid.UUID) (*SystemUser, error)
	GetByUsername(ctx context.Context, username string) (*SystemUser, error)
	Update(ctx context.Context, user *SystemUser) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*SystemUser, int, error)
	AssignRole(ctx context.Context, assignment *UserRoleAssignment) error
	GetRoles(ctx context.Context, userID uuid.UUID) ([]*UserRoleAssignment, error)
	RemoveRole(ctx context.Context, assignmentID uuid.UUID) error
}
