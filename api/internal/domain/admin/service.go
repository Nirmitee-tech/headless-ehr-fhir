package admin

import (
	"context"
	"fmt"

	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/google/uuid"
)

type Service struct {
	orgs  OrganizationRepository
	depts DepartmentRepository
	locs  LocationRepository
	users SystemUserRepository
	vt    *fhir.VersionTracker
}

// SetVersionTracker attaches an optional VersionTracker to the service.
func (s *Service) SetVersionTracker(vt *fhir.VersionTracker) {
	s.vt = vt
}

// VersionTracker returns the service's VersionTracker (may be nil).
func (s *Service) VersionTracker() *fhir.VersionTracker {
	return s.vt
}

func NewService(orgs OrganizationRepository, depts DepartmentRepository, locs LocationRepository, users SystemUserRepository) *Service {
	return &Service{orgs: orgs, depts: depts, locs: locs, users: users}
}

// -- Organization --

func (s *Service) CreateOrganization(ctx context.Context, org *Organization) error {
	if org.Name == "" {
		return fmt.Errorf("organization name is required")
	}
	if org.TypeCode == "" {
		org.TypeCode = "prov"
	}
	org.Active = true
	if err := s.orgs.Create(ctx, org); err != nil {
		return err
	}
	org.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Organization", org.FHIRID, org.ToFHIR())
	}
	return nil
}

func (s *Service) GetOrganization(ctx context.Context, id uuid.UUID) (*Organization, error) {
	return s.orgs.GetByID(ctx, id)
}

func (s *Service) GetOrganizationByFHIRID(ctx context.Context, fhirID string) (*Organization, error) {
	return s.orgs.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateOrganization(ctx context.Context, org *Organization) error {
	if org.Name == "" {
		return fmt.Errorf("organization name is required")
	}
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Organization", org.FHIRID, org.VersionID, org.ToFHIR())
		if err == nil {
			org.VersionID = newVer
		}
	}
	return s.orgs.Update(ctx, org)
}

func (s *Service) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		org, err := s.orgs.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Organization", org.FHIRID, org.VersionID)
		}
	}
	return s.orgs.Delete(ctx, id)
}

func (s *Service) ListOrganizations(ctx context.Context, limit, offset int) ([]*Organization, int, error) {
	return s.orgs.List(ctx, limit, offset)
}

func (s *Service) SearchOrganizations(ctx context.Context, params map[string]string, limit, offset int) ([]*Organization, int, error) {
	return s.orgs.Search(ctx, params, limit, offset)
}

// -- Department --

func (s *Service) CreateDepartment(ctx context.Context, dept *Department) error {
	if dept.Name == "" {
		return fmt.Errorf("department name is required")
	}
	if dept.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}
	dept.Active = true
	return s.depts.Create(ctx, dept)
}

func (s *Service) GetDepartment(ctx context.Context, id uuid.UUID) (*Department, error) {
	return s.depts.GetByID(ctx, id)
}

func (s *Service) UpdateDepartment(ctx context.Context, dept *Department) error {
	return s.depts.Update(ctx, dept)
}

func (s *Service) DeleteDepartment(ctx context.Context, id uuid.UUID) error {
	return s.depts.Delete(ctx, id)
}

func (s *Service) ListDepartments(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*Department, int, error) {
	return s.depts.ListByOrganization(ctx, orgID, limit, offset)
}

// -- Location --

func (s *Service) CreateLocation(ctx context.Context, loc *Location) error {
	if loc.Name == "" {
		return fmt.Errorf("location name is required")
	}
	if loc.Status == "" {
		loc.Status = "active"
	}
	if err := s.locs.Create(ctx, loc); err != nil {
		return err
	}
	loc.VersionID = 1
	if s.vt != nil {
		_ = s.vt.RecordCreate(ctx, "Location", loc.FHIRID, loc.ToFHIR())
	}
	return nil
}

func (s *Service) GetLocation(ctx context.Context, id uuid.UUID) (*Location, error) {
	return s.locs.GetByID(ctx, id)
}

func (s *Service) GetLocationByFHIRID(ctx context.Context, fhirID string) (*Location, error) {
	return s.locs.GetByFHIRID(ctx, fhirID)
}

func (s *Service) UpdateLocation(ctx context.Context, loc *Location) error {
	if s.vt != nil {
		newVer, err := s.vt.RecordUpdate(ctx, "Location", loc.FHIRID, loc.VersionID, loc.ToFHIR())
		if err == nil {
			loc.VersionID = newVer
		}
	}
	return s.locs.Update(ctx, loc)
}

func (s *Service) DeleteLocation(ctx context.Context, id uuid.UUID) error {
	if s.vt != nil {
		loc, err := s.locs.GetByID(ctx, id)
		if err == nil {
			_ = s.vt.RecordDelete(ctx, "Location", loc.FHIRID, loc.VersionID)
		}
	}
	return s.locs.Delete(ctx, id)
}

func (s *Service) ListLocations(ctx context.Context, limit, offset int) ([]*Location, int, error) {
	return s.locs.List(ctx, limit, offset)
}

// -- System User --

func (s *Service) CreateSystemUser(ctx context.Context, user *SystemUser) error {
	if user.Username == "" {
		return fmt.Errorf("username is required")
	}
	if user.UserType == "" {
		return fmt.Errorf("user_type is required")
	}
	if user.Status == "" {
		user.Status = "active"
	}
	return s.users.Create(ctx, user)
}

func (s *Service) GetSystemUser(ctx context.Context, id uuid.UUID) (*SystemUser, error) {
	return s.users.GetByID(ctx, id)
}

func (s *Service) GetSystemUserByUsername(ctx context.Context, username string) (*SystemUser, error) {
	return s.users.GetByUsername(ctx, username)
}

func (s *Service) UpdateSystemUser(ctx context.Context, user *SystemUser) error {
	return s.users.Update(ctx, user)
}

func (s *Service) DeleteSystemUser(ctx context.Context, id uuid.UUID) error {
	return s.users.Delete(ctx, id)
}

func (s *Service) ListSystemUsers(ctx context.Context, limit, offset int) ([]*SystemUser, int, error) {
	return s.users.List(ctx, limit, offset)
}

func (s *Service) AssignRole(ctx context.Context, assignment *UserRoleAssignment) error {
	if assignment.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}
	if assignment.RoleName == "" {
		return fmt.Errorf("role_name is required")
	}
	assignment.Active = true
	return s.users.AssignRole(ctx, assignment)
}

func (s *Service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRoleAssignment, error) {
	return s.users.GetRoles(ctx, userID)
}

func (s *Service) RemoveRole(ctx context.Context, assignmentID uuid.UUID) error {
	return s.users.RemoveRole(ctx, assignmentID)
}
